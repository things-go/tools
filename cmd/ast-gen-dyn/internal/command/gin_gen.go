package command

import (
	"fmt"
	"go/ast"
	"go/token"
	"log/slog"
	"path"
	"slices"

	"github.com/things-go/tools/cmd/ast-gen-dyn/internal/astdyn"
	"github.com/thinkgos/astgo"
	"golang.org/x/tools/go/packages"
)

type GinGenOption struct {
	Interface []string
	astdyn.Option
}

type GinGen struct {
	Pattern   []string
	OutputDir string
	GinGenOption
	Processed map[string]struct{}
}

func (g *GinGen) Generate() error {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports |
			packages.NeedSyntax,
		Tests: false,
		// BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
		// Logf: g.logf,
	}
	pkgs, err := packages.Load(cfg, g.Pattern...)
	if err != nil {
		return err
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("%d packages matching %v", len(pkgs), g.Pattern)
	}
	pkg := pkgs[0]
	for _, file := range pkg.Syntax {
		if err = g.InspectFile(file); err != nil {
			return err
		}
	}
	remainingInterfaceName := make([]string, 0, len(g.Interface))
	for _, v := range g.Interface {
		if _, ok := g.Processed[v]; !ok {
			remainingInterfaceName = append(remainingInterfaceName, v)
		}
	}
	if len(remainingInterfaceName) > 0 {
		slog.Warn("There are some unprocessed interface", slog.Any("interface", remainingInterfaceName))
	}
	return nil
}

func (g *GinGen) InspectFile(file *ast.File) error {
	imports := astgo.NewImportMgr()
	for _, imp := range file.Imports {
		if imp.Name == nil {
			imports.AddNamedImport("", astgo.ImportPath(imp))
		} else {
			imports.AddNamedImport(imp.Name.Name, astgo.ImportPath(imp))
		}
	}
	disposeName := make(map[string]struct{}, len(g.Interface))

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			_, ok = typeSpec.Type.(*ast.InterfaceType)
			if !ok || !slices.Contains(g.Interface, typeSpec.Name.Name) {
				continue
			}
			disposeName[typeSpec.Name.Name] = struct{}{}
			ifc, err := astdyn.ParserInterface(typeSpec, genDecl.Doc, &g.Option)
			if err != nil {
				return err
			}
			g.Processed[typeSpec.Name.Name] = struct{}{}
			f := &astgo.File{
				Filename: path.Join(g.OutputDir, astgo.SnakeCase(ifc.Name)+".gin.gen.go"),
				Doc:      nil,
				Name:     file.Name.Name,
				Imports:  astgo.NewImportMgr(),
				GenDecl:  nil,
				Decls:    nil,
			}
			imports.IterNamedImport(func(name, path string) bool {
				f.Imports.AddNamedImport(name, path)
				return true
			})

			// imports
			f.Imports.AddNamedImport("context", "context")
			f.Imports.AddNamedImport("errors", "errors")
			f.Imports.AddNamedImport("gin", "github.com/gin-gonic/gin")
			f.Imports.AddNamedImport("http", "github.com/things-go/dyn/transport/http")
			// generic declaration
			f.GenDecl = append(f.GenDecl,
				astgo.NewGenDeclReferenceVariable("errors", "New",
					astgo.WithDocComment(astgo.NewCommentGroup(true, "// Reference imports to suppress errors if they are not otherwise used.")),
				),
				astgo.NewGenDeclReferenceVariable("context", "TODO"),
				astgo.NewGenDeclReferenceVariable("gin", "New"),
			)
			// func declaration
			f.Decls = append(f.Decls, ifc.GenInterfaceGenDecl_GinHttpServer())
			f.Decls = append(f.Decls, ifc.GenFuncDecl_GinRegisterHttpServer())
			f.Decls = append(f.Decls, ifc.GenDecl_GinHandler()...)
			err = f.WriteFile()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
