package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

type StructMeta struct {
	PackageName string
	Name        string
}

type Meta struct {
	PackageName string
	FuncName    string
	Struct      []*StructMeta
}

type Gen struct {
	Output   string
	FuncName string
}

func (g *Gen) Gen() error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, 0)
	if err != nil {
		return err
	}
	ss := make([]*StructMeta, 0, 512)
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			sms, err := parseStruct(file)
			if err != nil {
				continue
			}
			ss = append(ss, sms...)
		}
	}
	if len(ss) == 0 {
		return nil
	}
	file, err := os.Create(args.Output)
	if err != nil {
		return err
	}
	return tpl.Execute(
		file,
		&Meta{
			PackageName: ss[0].PackageName,
			FuncName:    g.FuncName,
			Struct:      ss,
		},
	)
}

func parseStruct(file *ast.File) ([]*StructMeta, error) {
	sms := make([]*StructMeta, 0, 128)
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
			_, ok = typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			// not support generic type
			if typeSpec.TypeParams != nil {
				continue
			}
			sms = append(sms, &StructMeta{
				PackageName: file.Name.Name,
				Name:        typeSpec.Name.String(),
			})
		}
	}
	return sms, nil
}
