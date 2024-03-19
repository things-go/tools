package main

import (
	"embed"
	"html/template"
)

//go:embed var.tpl
var Static embed.FS
var tpl = template.Must(template.New("components").
	ParseFS(Static, "var.tpl")).Lookup("var.tpl")
