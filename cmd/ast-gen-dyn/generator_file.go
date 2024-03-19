package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"slices"

	"github.com/thinkgos/astgo"
	"golang.org/x/tools/imports"
)

type GeneratorFile struct {
	*Generator

	file    *ast.File
	imports *astgo.ImportMgr
}

func NewGeneratorFile(g *Generator, file *ast.File) *GeneratorFile {
	imports := astgo.NewImportMgr()
	for _, imp := range file.Imports {
		if imp.Name == nil {
			imports.AddNamedImport("", astgo.ImportPath(imp))
		} else {
			imports.AddNamedImport(imp.Name.Name, astgo.ImportPath(imp))
		}
	}
	return &GeneratorFile{
		Generator: g,
		file:      file,
		imports:   imports,
	}
}

func (g *GeneratorFile) Inspect() error {
	for _, decl := range g.file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.InterfaceType); ok && slices.Contains(g.InterfacePatterns, typeSpec.Name.Name) {
						err := g.GenerateGinFromInterfaceType(typeSpec, genDecl.Doc)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func (g *GeneratorFile) GenerateGinFromInterfaceType(typeSpec *ast.TypeSpec, genDeclDoc *ast.CommentGroup) error {
	file := &File{
		Name:      g.file.Name.Name,
		ImportMgr: astgo.NewImportMgr(),
		valuesDecls: []ast.Decl{
			astgo.NewGenDeclReferenceVariable(
				"errors", "New",
				astgo.WithDocComment(astgo.NewCommentGroup(true, "// Reference imports to suppress errors if they are not otherwise used.")),
			),
			astgo.NewGenDeclReferenceVariable("context", "TODO"),
			astgo.NewGenDeclReferenceVariable("gin", "New"),
		},
	}
	g.imports.IterNamedImport(func(name, path string) bool {
		file.ImportMgr.AddNamedImport(name, path)
		return true
	})
	file.ImportMgr.AddNamedImport("context", "context")
	file.ImportMgr.AddNamedImport("errors", "errors")
	file.ImportMgr.AddNamedImport("gin", "github.com/gin-gonic/gin")

	impl, err := createHttpServerImpl(typeSpec, genDeclDoc)
	if err != nil {
		return err
	}
	// declaration interface
	file.decls = append(file.decls, impl.declInterface)
	file.decls = append(file.decls, impl.declRegisterFunc)
	file.decls = append(file.decls, impl.handlers...)
	return file.Gen()
}

type File struct {
	Doc         *ast.CommentGroup
	Name        string           // 包名
	ImportMgr   *astgo.ImportMgr // import mgr
	valuesDecls []ast.Decl       // Var or Const
	decls       []ast.Decl
}

func (f *File) Gen() error {
	fset := token.NewFileSet()
	file := &ast.File{
		Doc:  f.Doc,
		Name: astgo.NewIdent(f.Name),
	}
	//* imports
	importSpecs := make([]ast.Spec, 0, f.ImportMgr.Len())
	for _, imp := range f.ImportMgr.Imports() {
		importSpecs = append(importSpecs, imp)
	}
	file.Decls = append(file.Decls, astgo.NewGenDeclImport(importSpecs...).Build())
	//* var or const
	file.Decls = append(file.Decls, f.valuesDecls...)
	//* function
	file.Decls = append(file.Decls, f.decls...)
	buf := &bytes.Buffer{}
	err := format.Node(buf, fset, file)
	if err != nil {
		return err
	}
	// data := buf.Bytes()
	data, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}

func httpServerInterfaceName(name string) string {
	return name + "HTTPServer"
}

func httpServerRegisterName(name string) string {
	return "Register" + name + "HTTPServer"
}

type httpServerImpl struct {
	declInterface    *ast.GenDecl
	declRegisterFunc ast.Decl
	handlers         []ast.Decl
}

func createHttpServerImpl(typeSpec *ast.TypeSpec, genDeclDoc *ast.CommentGroup) (*httpServerImpl, error) {
	ifcType := typeSpec.Type.(*ast.InterfaceType)
	ifcTypeName := typeSpec.Name.Name
	ifcTypeNewName := httpServerInterfaceName(ifcTypeName)

	ifcTypeNewMethods := make([]*ast.Field, 0, len(ifcType.Methods.List))
	registerStmt := make([]ast.Stmt, 0, len(ifcType.Methods.List))
	decls := make([]ast.Decl, 0, len(ifcType.Methods.List))
	for _, method := range ifcType.Methods.List {
		methodName := astgo.IdentName(method.Names)
		methodType := method.Type.(*ast.FuncType)
		if len(methodType.Params.List) > 1 ||
			methodType.Results == nil || len(methodType.Results.List) > 1 {
			return nil, fmt.Errorf("Definition of the interface(%v)'s method(%v) param and result must be only one.", ifcTypeName, methodName)
		}
		//* http server interface
		param := methodType.Params.List[0]
		result := methodType.Results.List[0]
		ifcTypeNewMethods = append(ifcTypeNewMethods,
			astgo.
				NewField(
					astgo.NewFuncType().
						Param(
							astgo.NewField(Expr_Context()).FuncField(),
							astgo.NewField(astgo.CloneExpr(param.Type)).FuncField(),
						).
						Result(
							astgo.NewField(astgo.CloneExpr(result.Type)).FuncField(),
							astgo.NewField(astgo.IdentError()).FuncField(),
						).
						Build(),
				).
				Name(methodName).
				Doc(astgo.CloneCommentGroup(true, method.Doc)).
				Comment(astgo.CloneCommentGroup(false, method.Comment)).
				FuncField(),
		)

		stmt, decl := createHttpServerMethodHandler(ifcTypeName, methodName, methodType)
		registerStmt = append(registerStmt, stmt)
		decls = append(decls, decl)
	}
	declInterface := astgo.NewGenDeclType(astgo.NewTypeSpec(ifcTypeNewName, astgo.NewInterfaceType(ifcTypeNewMethods...))).
		Doc(astgo.CloneCommentGroup(true, genDeclDoc)).
		Build()

	// func Register{{Service}}HTTPServer(g *gin.RouterGroup, srv {{Service}}HTTPServer) {
	//      r := g.Group("")
	//      {
	//          r.POST("{{Path}}", _{{Service}}_{{Method}}{{Num}}_HTTP_Handler(srv))
	//      }
	//  }
	declRegisterFunc := astgo.NewFuncDecl(
		httpServerRegisterName(ifcTypeName),
		astgo.NewFuncType().Param(
			astgo.NewField(astgo.NewChainExprIdent("gin").Selector("RouterGroup").Star().Build()).Name("g").FuncField(),
			astgo.NewField(astgo.NewIdent(ifcTypeNewName)).Name("srv").FuncField(),
		).Build(),
		astgo.NewBlockStmt(
			astgo.NewShortVarDeclStmt().
				Lhs(astgo.NewIdent("r")).
				Rhs(astgo.NewChainExprIdent("g").Selector("Group").Call(astgo.IdentEmptyString()).Build()).
				Build(),
			astgo.NewBlockStmt(registerStmt...),
		),
	)
	return &httpServerImpl{
		declInterface:    declInterface,
		declRegisterFunc: declRegisterFunc,
		handlers:         decls,
	}, nil
}

func createHttpServerMethodHandler(ifcTypeName, methodName string, methodType *ast.FuncType) (ast.Stmt, ast.Decl) {
	param := methodType.Params.List[0]
	result := methodType.Results.List[0]
	_ = result

	blockStmts := make([]ast.Stmt, 0, 256)
	{
		// carrier := http.FromCarrier(c.Request.Context())
		blockStmts = append(blockStmts,
			astgo.NewShortVarDeclStmt().
				Lhs(astgo.NewIdent("carrier")).
				Rhs(
					astgo.NewChainExprIdent("http").
						Selector("FromCarrier").
						Call(
							astgo.NewChainExprIdent("c").
								Selector("Request").
								Selector("Context").
								Call().
								Build(),
						).
						Build(),
				).
				Build(),
			// shouldBind := func(req *AddPoRequest) error {
			// 	if err := c.ShouldBind(req); err != nil {
			// 		return err
			// 	}
			// 	return carrier.Validate(c.Request.Context(), req)
			// }
			astgo.NewShortVarDeclStmt().
				Lhs(astgo.NewIdent("shouldBind")).
				Rhs(
					astgo.NewFuncLit(
						astgo.NewFuncType().
							Param(
								astgo.NewField(astgo.CloneExpr(param.Type)).Name("req").FuncField(),
							).
							Result(astgo.NewField(astgo.IdentError()).FuncField()).
							Build(),
						astgo.NewBlockStmt(
							astgo.NewIfStmt().
								Init(astgo.NewShortVarDeclStmt().Lhs(astgo.IdentErr()).Rhs(astgo.NewChainExprIdent("c").Selector("ShouldBind").Call(astgo.NewIdent("req")).Build()).Build()).
								Cond(astgo.NewChainExpr(astgo.IdentErr()).Binary(token.NEQ, astgo.IdentNil()).Build()).
								Body(astgo.NewBlockStmt(
									astgo.NewReturnStmt(astgo.IdentErr()),
								)).
								Build(),
							astgo.NewReturnStmt(
								astgo.NewChainExprIdent("carrier").
									Selector("Validate").
									Call(
										astgo.NewChainExprIdent("c").Selector("Request").Selector("Context").Build(),
										astgo.NewIdent("req"),
									).
									Build(),
							),
						),
					),
				).
				Build(),
			// var err error
			// var req AddPoRequest
			// var reply *emptypb.Empty
			astgo.NewDeclStmt(astgo.NewGenDeclVar(astgo.NewTypeSpec("err", astgo.IdentError())).Build()),
			astgo.NewDeclStmt(astgo.NewGenDeclVar(astgo.NewTypeSpec("req", astgo.UnStar(astgo.CloneExpr(param.Type)))).Build()),
			astgo.NewDeclStmt(astgo.NewGenDeclVar(astgo.NewTypeSpec("reply", astgo.CloneExpr(result.Type))).Build()),
			// if err = shouldBind(&req); err != nil {
			// 	carrier.Error(c, err)
			// 	return
			// }
			astgo.NewIfStmt().
				Init(
					astgo.NewAssignStmt().
						Lhs(astgo.IdentErr()).
						Rhs(astgo.NewChainExprIdent("shouldBind").Call(astgo.NewUnaryExpr(token.AND, astgo.NewIdent("srv"))).Build()).
						Build(),
				).
				Cond(astgo.NewBinaryExpr(token.NEQ, astgo.IdentErr(), astgo.IdentNil())).
				Body(astgo.NewBlockStmt(
					astgo.NewExprStmt(
						astgo.NewChainExprIdent("carrier").Selector("Error").
							Call(
								astgo.NewIdent("c"), astgo.IdentErr(),
							).
							Build(),
					),
					astgo.NewReturnStmt(),
				)).
				Build(),

			// reply, err = srv.PublishPo(c.Request.Context(), &req)
			astgo.NewAssignStmt().
				Lhs(astgo.NewIdent("reply"), astgo.IdentErr()).
				Rhs(
					astgo.NewChainExprIdent("srv").
						Selector(methodName).
						Call(
							astgo.NewChainExprIdent("c").Selector("Request").Selector("Context").Call().Build(),
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
					astgo.NewExprStmt(
						astgo.NewChainExprIdent("carrier").Selector("Error").Call(
							astgo.NewIdent("c"), astgo.IdentErr(),
						).Build(),
					),
					astgo.NewReturnStmt(),
				)).
				Build(),
			// carrier.Render(c, reply)
			astgo.NewExprStmt(
				astgo.NewChainExprIdent("carrier").Selector("Render").Call(
					astgo.NewIdent("c"), astgo.NewIdent("reply"),
				).Build(),
			),
		)
	}

	ifcTypeNewName := httpServerInterfaceName(ifcTypeName)
	handlerFuncName := fmt.Sprintf("_%s_%s%d_HTTP_Handler", ifcTypeName, methodName, 0)
	// func _{{ifc_name}}_{{method_name}}{{num}}_HTTP_Handler(srv {{ifc_name}}HTTPServer) gin.HandlerFunc {
	// 	...block stmt
	// }
	declHandler := astgo.NewFuncDecl(
		handlerFuncName,
		astgo.NewFuncType().
			Param(astgo.NewField(astgo.NewIdent(ifcTypeNewName)).Name("srv").FuncField()).
			Result(astgo.NewField(Expr_GinHandlerFunc()).FuncField()).
			Build(),
		astgo.NewBlockStmt(
			// return func(c *gin.Context) {
			// 	...block stmt
			// }
			astgo.NewReturnStmt(
				astgo.NewFuncLit(
					astgo.NewFuncType().
						Param(astgo.NewField(Expr_StarGinContext()).Name("c").FuncField()).
						Build(),
					astgo.NewBlockStmt(blockStmts...),
				),
			),
		),
	)
	// r.POST("/v1/hello", _Greeter_SayHello0_HTTP_Handler(srv))
	registerStmt := astgo.NewExprStmt(
		astgo.NewChainExprIdent("r").
			Selector("POST").
			Call(
				astgo.NewLitString(`"/v1/hello"`),
				astgo.NewCallExpr(astgo.NewIdent(handlerFuncName), astgo.NewIdent("srv")),
			).
			Build(),
	)
	return registerStmt, declHandler
}
