package astdyn

import (
	"fmt"
	"go/ast"
	"strconv"

	"github.com/thinkgos/astgo"
)

type Option struct {
	AllowDeleteBody     bool
	AllowEmptyPatchBody bool
	UseEncoding         bool
}

type InterfaceMetadata struct {
	Name    string
	Type    *ast.InterfaceType
	Doc     *ast.CommentGroup
	Methods []*MethodMetadata
}

func ParserInterface(typeSpec *ast.TypeSpec, doc *ast.CommentGroup, o *Option) (*InterfaceMetadata, error) {
	if o == nil {
		o = &Option{}
	}
	name := typeSpec.Name.Name
	typ := typeSpec.Type.(*ast.InterfaceType)
	methods := make([]*MethodMetadata, 0, len(typ.Methods.List))
	for _, m := range typ.Methods.List {
		method, err := parserMethod(name, m, o)
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}
	return &InterfaceMetadata{
		Name:    name,
		Type:    typ,
		Doc:     doc,
		Methods: methods,
	}, nil
}

//	type {{service}}HTTPServer interface {
//	   func MethodName(context.Context, *{{method_param}}) ({{method_result}}, error)
//	}
func (m *InterfaceMetadata) GenInterfaceGenDecl_GinHttpServer() *ast.GenDecl {
	newName := httpServerInterfaceName(m.Name)
	newMethods := make([]*ast.Field, 0, len(m.Methods))
	for _, method := range m.Methods {
		newMethods = append(newMethods,
			// func MethodName(context.Context, *{{method_param}}) ({{method_result}}, error)
			astgo.
				NewField(
					astgo.NewFuncType().
						Param(
							NewField_Context().FuncField(),
							NewFieldCloneExpr(method.Param.Type).FuncField(),
						).
						Result(
							NewFieldCloneExpr(method.Result.Type).FuncField(),
							NewField_Error().FuncField(),
						).
						Build(),
				).
				Name(method.Name).
				Doc(astgo.CloneCommentGroup(true, method.Doc)).
				Comment(astgo.CloneCommentGroup(false, method.Comment)).
				FuncField(),
		)
	}
	return astgo.NewGenDeclType(astgo.NewTypeSpec(newName, astgo.NewInterfaceType(newMethods...))).
		Doc(astgo.CloneCommentGroup(true, m.Doc)).
		Build()
}

//	func Register{{service}}HTTPServer(g *gin.RouterGroup, srv {{service}}HTTPServer) {
//	    r := g.Group("")
//	    {
//	        r.{{method}}("{{path}}", _{{ifc_name}}_{{method_name}}{{num}}_HTTP_Handler(srv))
//	    }
//	}
func (m *InterfaceMetadata) GenFuncDecl_GinRegisterHttpServer() *ast.FuncDecl {
	newName := httpServerInterfaceName(m.Name)
	stmts := make([]ast.Stmt, 0, len(m.Methods))
	for _, method := range m.Methods {
		for num, rule := range method.HttpRule.Rules {
			// r.{{method}}("{{path}}", _{{ifc_name}}_{{method_name}}{{num}}_HTTP_Handler(srv))
			stmts = append(stmts,
				astgo.NewExprStmt(
					astgo.NewChainExprIdent("r").
						Selector(rule.Method).
						Call(
							astgo.NewLitString(strconv.Quote(rule.Path)),
							astgo.NewChainExprIdent(httpServerHandlerFuncName(m.Name, method.Name, num)).
								Call(astgo.NewIdent("srv")).
								Build(),
						).
						Build(),
				),
			)
		}
	}
	return astgo.NewFuncDecl(
		registerHttpServerFuncName(m.Name),
		astgo.NewFuncType().
			Param(
				astgo.NewField(NewExpr_StarGinRouterGroup()).Name("g").FuncField(),
				astgo.NewField(astgo.NewIdent(newName)).Name("srv").FuncField(),
			).
			Build(),
		astgo.NewBlockStmt(
			astgo.NewShortVarDeclStmt().
				Lhs(astgo.NewIdent("r")).
				Rhs(astgo.NewChainExprIdent("g").Selector("Group").Call(astgo.IdentEmptyString()).Build()).
				Build(),
			astgo.NewBlockStmt(stmts...),
		),
	)
}

func (m *InterfaceMetadata) GenDecl_GinHandler() []ast.Decl {
	decls := make([]ast.Decl, 0, len(m.Methods)*2)
	for _, method := range m.Methods {
		decls = append(decls, method.GenDecl_GinHandler(m.Name)...)
	}
	return decls
}

func httpServerInterfaceName(name string) string {
	return name + "HTTPServer"
}

func registerHttpServerFuncName(name string) string {
	return "Register" + name + "HTTPServer"
}

func httpServerHandlerFuncName(ifcTypeName, methodName string, num int) string {
	return fmt.Sprintf("_%s_%s%d_HTTP_Handler", ifcTypeName, methodName, num)
}
