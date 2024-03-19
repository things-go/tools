package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

var args = &struct {
	Output   string
	FuncName string
}{}

func usage() {
	fmt.Fprintf(os.Stderr, "scan-struct is a tool for scan struct to model.\n")
	fmt.Fprintf(os.Stderr, "Usage: \n")
	fmt.Fprintf(os.Stderr, "\t scan-struct [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	fmt.Fprintf(os.Stderr, "\t -output <output filename>, default: {{source}}/struct_scan_list.gen.go\n")
	fmt.Fprintf(os.Stderr, "\t -func_name <function name>, default: ListStructScan\n")
}

func init() {
	flag.StringVar(&args.Output, "output", "struct_scan_list.gen.go", "Output filename, default: {{source}}/struct_scan_list.gen.go")
	flag.StringVar(&args.FuncName, "func_name", "ListStructScan", "function name, default: ListStructScan")

	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})
	slog.SetDefault(slog.New(textHandler))
}

func main() {
	flag.Usage = usage
	flag.Parse()
	g := Gen{
		Output:   args.Output,
		FuncName: args.FuncName,
	}
	err := g.Gen()
	if err != nil {
		slog.Debug("---> " + err.Error())
	}
}
