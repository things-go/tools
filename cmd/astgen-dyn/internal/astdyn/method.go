package astdyn

import (
	"fmt"
	"go/ast"
	"go/token"
	"net/http"

	"github.com/things-go/tools/proc_http"
	"github.com/thinkgos/astgo"
)

type Rule struct {
	Method       string // http method {get|put|post|delete|patch}
	Path         string // http path "/v1/hello"
	Body         string // http request body
	ResponseBody string // http response body
	HasVars      bool   // 是否有url参数
	HasBody      bool   // 是否有消息体
	Custom       bool   // http method and path is customize
}

type HttpRule struct {
	// Rules http rules
	Rules []*Rule
}

type MethodMetadata struct {
	Name     string
	FuncType *ast.FuncType
	Param    *ast.Field
	Result   *ast.Field
	Doc      *ast.CommentGroup
	Comment  *ast.CommentGroup
	HttpRule *HttpRule
	opt      *Option
}

func parserMethod(ifcName string, method *ast.Field, opt *Option) (*MethodMetadata, error) {
	name := astgo.IdentName(method.Names)
	funcType := method.Type.(*ast.FuncType)
	if funcType.Params == nil || len(funcType.Params.List) > 1 ||
		funcType.Results == nil || len(funcType.Results.List) > 1 {
		return nil, fmt.Errorf("Definition of the interface(%v)'s method(%v) param and result must be only one.", ifcName, name)
	}
	comments := astgo.FromComment(method.Doc)
	httpRule, err := proc_http.Parse(comments, false)
	if err != nil {
		return nil, err
	}
	if httpRule == nil {
		return nil, fmt.Errorf("Definition of the interface(%v)'s method(%v) must have a http annotation.", ifcName, name)
	}

	rule, err := buildHTTPRule(httpRule, opt)
	if err != nil {
		return nil, err
	}
	return &MethodMetadata{
		Name:     name,
		FuncType: funcType,
		Param:    funcType.Params.List[0],
		Result:   funcType.Results.List[0],
		Doc:      method.Doc,
		HttpRule: rule,
		opt:      opt,
	}, nil
}

func buildHTTPRule(httpRule *proc_http.HttpRule, opt *Option) (*HttpRule, error) {
	rules := make([]*Rule, 0, len(httpRule.Rules))
	for _, ru := range httpRule.Rules {
		rule := &Rule{
			Method:       ru.Method,
			Path:         transformPathParams(ru.Path),
			Body:         ru.Body,
			ResponseBody: ru.ResponseBody,
			HasVars:      len(buildPathVars(ru.Path)) > 0,
			HasBody:      false,
			Custom:       ru.Custom,
		}
		body := rule.Body
		switch {
		case rule.Method == http.MethodGet:
			if body != "" {
				return nil, fmt.Errorf("%s %s body should not be declared.", ru.Method, ru.Path)
			}
			rule.HasBody = false
		case ru.Method == http.MethodDelete:
			if body != "" {
				rule.HasBody = true
				if !opt.AllowDeleteBody {
					rule.HasBody = false
					return nil, fmt.Errorf("%s %s body should not be declared.", ru.Method, ru.Path)
				}
			} else {
				rule.HasBody = false
			}
		case ru.Method == http.MethodPatch:
			if body != "" {
				rule.HasBody = true
			} else {
				rule.HasBody = false
				if !opt.AllowEmptyPatchBody {
					return nil, fmt.Errorf("%s %s is does not declare a body.", ru.Method, ru.Path)
				}
			}
		case body == "*":
			rule.HasBody = true
			rule.Body = ""
		case body != "":
			rule.HasBody = true
			rule.Body = "." + camelCaseVars(body)
		default:
			rule.HasBody = false
			return nil, fmt.Errorf("%s %s is does not declare a body.", ru.Method, ru.Path)
		}
		if rule.ResponseBody == "*" {
			rule.ResponseBody = ""
		} else if rule.ResponseBody != "" {
			rule.ResponseBody = "." + camelCaseVars(rule.ResponseBody)
		}
		rules = append(rules, rule)
	}
	return &HttpRule{Rules: rules}, nil
}

func (method *MethodMetadata) GenDecl_GinHandler(ifcName string) []ast.Decl {
	decls := make([]ast.Decl, 0, len(method.HttpRule.Rules))
	for num, rule := range method.HttpRule.Rules {
		decls = append(decls, method.genGinHandler(ifcName, rule, num))
	}
	return decls
}

func (method *MethodMetadata) genGinHandler(ifcTypeName string, rule *Rule, num int) ast.Decl {
	blockStmt := astgo.NewBlockStmtBuilder()

	// carrier := http.FromCarrier(c.Request.Context())
	blockStmt.Stmt(
		astgo.NewShortVarDeclStmt().
			Lhs(astgo.NewIdent("carrier")).
			Rhs(
				astgo.NewChainExprIdent("http").
					Selector("FromCarrier").
					Call(NewExpr_CRequestContextCall()).
					Build(),
			).
			Build(),
	)
	if method.opt.UseEncoding && rule.HasVars {
		// c.Request = carrier.WithValueUri(c.Request, c.Params)
		blockStmt.Stmt(
			astgo.NewAssignStmt().
				Lhs(astgo.NewChainExprIdent("c").Selector("Request").Build()).
				Rhs(
					astgo.NewChainExprIdent("carrier").
						Selector("WithValueUri").
						Call(
							astgo.NewChainExprIdent("c").Selector("Request").Build(),
							astgo.NewChainExprIdent("c").Selector("Params").Build(),
						).
						Build(),
				).
				Build(),
		)
	}
	// binding
	{
		// shouldBind := func(req *AddPoRequest) error {
		// 	if err := c.ShouldBind(req); err != nil {
		// 		return err
		// 	}
		// 	return carrier.Validate(c.Request.Context(), req)
		// }
		shouldBindBodyStmts := astgo.NewBlockStmtBuilder()
		if method.opt.UseEncoding {
			if rule.HasBody {
				shouldBindBodyStmts.Stmt(NewStmt_BindIf("Bind", "req"+rule.Body))
				if rule.Body != "" {
					shouldBindBodyStmts.Stmt(NewStmt_BindIf("BindQuery", "req"))
				}
			} else {
				if rule.Method != http.MethodPatch {
					shouldBindBodyStmts.Stmt(NewStmt_BindIf("BindQuery", "req"+rule.Body))
				}
			}
			if rule.HasVars {
				shouldBindBodyStmts.Stmt(NewStmt_BindIf("BindUri", "req"))
			}
		} else {
			if rule.HasBody {
				shouldBindBodyStmts.Stmt(NewStmt_GinBindIf("ShouldBind", "req"+rule.Body))
				if rule.Body != "" {
					shouldBindBodyStmts.Stmt(NewStmt_GinBindIf("ShouldBindQuery", "req"))
				}
			} else {
				if rule.Method != http.MethodPatch {
					shouldBindBodyStmts.Stmt(NewStmt_GinBindIf("ShouldBindQuery", "req"+rule.Body))
				}
			}
			if rule.HasVars {
				shouldBindBodyStmts.Stmt(NewStmt_GinBindIf("ShouldBindUri", "req"))
			}
		}

		shouldBindBodyStmts.Stmt(
			astgo.NewReturnStmt(
				astgo.NewChainExprIdent("carrier").
					Selector("Validate").
					Call(
						NewExpr_CRequestContextCall(),
						astgo.NewIdent("req"),
					).
					Build(),
			),
		)
		blockStmt.Stmt(
			astgo.NewShortVarDeclStmt().
				Lhs(astgo.NewIdent("shouldBind")).
				Rhs(
					astgo.NewFuncLit(
						astgo.NewFuncType().
							Param(
								astgo.NewField(astgo.CloneExpr(method.Param.Type)).Name("req").FuncField(),
							).
							Result(astgo.NewField(astgo.IdentError()).FuncField()).
							Build(),
						shouldBindBodyStmts.Build(),
					),
				).
				Build(),
		)
	}

	blockStmt.Stmt(
		// var err error
		// var req AddPoRequest
		// var reply *emptypb.Empty
		astgo.NewDeclStmt(astgo.NewGenDeclVar(astgo.NewTypeSpec("err", astgo.IdentError())).Build()),
		astgo.NewDeclStmt(astgo.NewGenDeclVar(astgo.NewTypeSpec("req", astgo.UnStar(astgo.CloneExpr(method.Param.Type)))).Build()),
		astgo.NewDeclStmt(astgo.NewGenDeclVar(astgo.NewTypeSpec("reply", astgo.CloneExpr(method.Result.Type))).Build()),
	)
	blockStmt.Stmt(
		// if err = shouldBind(&req); err != nil {
		// 	carrier.Error(c, err)
		// 	return
		// }
		astgo.NewIfStmt().
			Init(
				astgo.NewAssignStmt().
					Lhs(astgo.IdentErr()).
					Rhs(astgo.NewChainExprIdent("shouldBind").Call(astgo.NewUnaryExpr(token.AND, astgo.NewIdent("req"))).Build()).
					Build(),
			).
			Cond(astgo.NewBinaryExpr(token.NEQ, astgo.IdentErr(), astgo.IdentNil())).
			Body(astgo.NewBlockStmt(
				NewStmt_CarrierError(),
				astgo.NewReturnStmt(),
			)).
			Build(),
	)
	blockStmt.Stmt(
		// reply, err = srv.{{method_name}}(c.Request.Context(), &req)
		astgo.NewAssignStmt().
			Lhs(astgo.NewIdent("reply"), astgo.IdentErr()).
			Rhs(
				astgo.NewChainExprIdent("srv").
					Selector(method.Name).
					Call(
						NewExpr_CRequestContextCall(),
						astgo.NewUnaryExpr(token.AND, astgo.NewIdent("req")),
					).
					Build(),
			).Build(),
		// if err != nil {
		// 	carrier.Error(c, err)
		// 	return
		// }
		astgo.NewIfStmt().
			Cond(astgo.NewBinaryExpr(token.NEQ, astgo.IdentErr(), astgo.IdentNil())).
			Body(astgo.NewBlockStmt(
				NewStmt_CarrierError(),
				astgo.NewReturnStmt(),
			)).
			Build(),
		// carrier.Render(c, reply)
		astgo.NewExprStmt(
			astgo.NewChainExprIdent("carrier").Selector("Render").Call(
				astgo.NewIdent("c"), astgo.NewIdent("reply"+rule.ResponseBody),
			).Build(),
		),
	)

	ifcTypeNewName := httpServerInterfaceName(ifcTypeName)
	// func _{{ifc_name}}_{{method_name}}{{num}}_HTTP_Handler(srv {{ifc_name}}HTTPServer) gin.HandlerFunc {
	// 		return func(c *gin.Context) {
	// 			...block stmt
	// 		}
	// }
	return astgo.NewFuncDecl(
		httpServerHandlerFuncName(ifcTypeName, method.Name, num),
		astgo.NewFuncType().
			Param(astgo.NewField(astgo.NewIdent(ifcTypeNewName)).Name("srv").FuncField()).
			Result(astgo.NewField(NewExpr_GinHandlerFunc()).FuncField()).
			Build(),
		astgo.NewBlockStmt(
			// return func(c *gin.Context) {
			// 	...block stmt
			// }
			astgo.NewReturnStmt(
				astgo.NewFuncLit(
					astgo.NewFuncType().
						Param(astgo.NewField(NewExpr_StarGinContext()).Name("c").FuncField()).
						Build(),
					blockStmt.Build(),
				),
			),
		),
	)
}
