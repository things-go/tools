package main

import (
	"os"

	"github.com/things-go/tools/cmd/ast-gen-dyn/internal/command"
)

func main() {
	err := command.NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
