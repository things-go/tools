package main

import (
	"go/ast"

	"github.com/thinkgos/astgo"
)

// Expr `context.Context`
func Expr_Context() ast.Expr {
	return astgo.NewChainExpr(astgo.NewIdent("context")).Selector("Context").Build()
}

// Expr `gin.HandlerFunc`
func Expr_GinHandlerFunc() ast.Expr {
	return astgo.NewChainExpr(astgo.NewIdent("gin")).Selector("HandlerFunc").Build()
}

// Expr `*gin.Context`
func Expr_StarGinContext() ast.Expr {
	return astgo.NewChainExpr(astgo.NewIdent("gin")).Selector("Context").Star().Build()
}

func Field_Error() *ast.Field {
	return astgo.NewField(astgo.IdentError()).FuncField()
}
