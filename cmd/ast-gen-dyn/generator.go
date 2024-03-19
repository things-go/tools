package main

import (
	"log"

	"golang.org/x/tools/go/packages"
)

type Generator struct {
	InputPattern []string
	OutputDir    string

	InterfacePatterns []string
}

func (g *Generator) Parser() error {
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
	pkgs, err := packages.Load(cfg, g.InputPattern...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages matching %v", len(pkgs), g.InputPattern)
	}
	pkg := pkgs[0]
	for _, file := range pkg.Syntax {
		gfile := NewGeneratorFile(g, file)
		err = gfile.Inspect()
		if err != nil {
			return err
		}
	}
	return nil
}
