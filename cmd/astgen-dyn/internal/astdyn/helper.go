package astdyn

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/thinkgos/astgo"
)

// Expr `context.Context`
func NewExpr_Context() ast.Expr {
	return astgo.NewChainExpr(astgo.NewIdent("context")).Selector("Context").Build()
}

// Expr `gin.HandlerFunc`
func NewExpr_GinHandlerFunc() ast.Expr {
	return astgo.NewChainExpr(astgo.NewIdent("gin")).Selector("HandlerFunc").Build()
}

// Expr `*gin.Context`
func NewExpr_StarGinContext() ast.Expr {
	return astgo.NewChainExpr(astgo.NewIdent("gin")).Selector("Context").Star().Build()
}

// Expr `*gin.RouterGroup“
func NewExpr_StarGinRouterGroup() ast.Expr {
	return astgo.NewChainExprIdent("gin").Selector("RouterGroup").Star().Build()
}

// Expr `c.Request.Context()`
func NewExpr_CRequestContextCall() ast.Expr {
	return astgo.NewChainExprIdent("c").
		Selector("Request").
		Selector("Context").
		Call().
		Build()
}

// Field `error`
func NewField_Error() *astgo.FieldBuilder {
	return astgo.NewField(astgo.IdentError())
}

// Field `context.Context`
func NewField_Context() *astgo.FieldBuilder {
	return astgo.NewField(NewExpr_Context())
}

func NewFieldCloneExpr(typ ast.Expr) *astgo.FieldBuilder {
	return astgo.NewField(astgo.CloneExpr(typ))
}

// Stmt `carrier.Error(c, err)`
func NewStmt_CarrierError() ast.Stmt {
	return astgo.NewExprStmt(
		astgo.NewChainExprIdent("carrier").Selector("Error").
			Call(
				astgo.NewIdent("c"), astgo.IdentErr(),
			).
			Build(),
	)
}

//	if err := carrier.{{bindName}}(c, {{reqName}}); err != nil {
//		return err
//	}
func NewStmt_BindIf(bindName, reqName string) ast.Stmt {
	return astgo.NewIfStmt().
		Init(
			astgo.NewShortVarDeclStmt().
				Lhs(astgo.IdentErr()).
				Rhs(
					astgo.NewChainExprIdent("carrier").
						Selector(bindName).
						Call(astgo.NewIdent("c"), astgo.NewIdent(reqName)).
						Build(),
				).
				Build()).
		Cond(astgo.NewChainExpr(astgo.IdentErr()).Binary(token.NEQ, astgo.IdentNil()).Build()).
		Body(astgo.NewBlockStmt(astgo.NewReturnStmt(astgo.IdentErr()))).
		Build()
}

//	if err := c.{{bindName}}({{reqName}}); err != nil {
//		return err
//	}
func NewStmt_GinBindIf(bindName, reqName string) ast.Stmt {
	return astgo.NewIfStmt().
		Init(
			astgo.NewShortVarDeclStmt().
				Lhs(astgo.IdentErr()).
				Rhs(
					astgo.NewChainExprIdent("c").
						Selector(bindName).
						Call(astgo.NewIdent(reqName)).
						Build(),
				).
				Build()).
		Cond(astgo.NewChainExpr(astgo.IdentErr()).Binary(token.NEQ, astgo.IdentNil()).Build()).
		Body(astgo.NewBlockStmt(astgo.NewReturnStmt(astgo.IdentErr()))).
		Build()
}

// TransformPathParams 路由 {xx} --> :xx
func transformPathParams(path string) string {
	paths := strings.Split(path, "/")
	for i, p := range paths {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") || strings.HasPrefix(p, ":") {
			paths[i] = ":" + p[1:len(p)-1]
		}
	}
	return strings.Join(paths, "/")
}

func buildPathVars(path string) (res []string) {
	for _, v := range strings.Split(path, "/") {
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			res = append(res, strings.TrimSuffix(strings.TrimPrefix(v, "{"), "}"))
		}
	}
	return
}

func camelCaseVars(s string) string {
	vars := make([]string, 0)
	subs := strings.Split(s, ".")
	for _, sub := range subs {
		vars = append(vars, astgo.CamelCase(sub))
	}
	return strings.Join(vars, ".")
}
