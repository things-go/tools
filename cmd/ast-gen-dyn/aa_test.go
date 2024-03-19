package main

import (
	"go/format"
	"go/token"
	"os"
	"testing"

	"github.com/thinkgos/astgo"
)

func Test(t *testing.T) {
	file := token.NewFileSet()
	format.Node(
		os.Stdout,
		file,
		astgo.NewFuncDecl(
			"funcName",
			astgo.NewFuncType().
				Param(astgo.NewField(astgo.NewIdent("ifcTypeNewName")).Name("srv").FuncField()).
				Result(astgo.NewField(astgo.NewSelectorExpr(astgo.NewIdent("gin"), "HandlerFunc")).FuncField()).
				Build(),
			astgo.NewBlockStmt(),
		),

		// NewFuncDecl(
		// 	"Hello",
		// 	NewFuncType(
		// 		[]*ast.Field{
		// 			NewField(astgo.NewIdent("int")).
		// 				Name("req").
		// 				Name("resp").
		// 				FuncSignatureField(),
		// 		},
		// 		nil,
		// 	),
		// 	NewBlockStmt(),
		// ),
	)
}
