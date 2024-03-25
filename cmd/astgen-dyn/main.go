package main

import (
	"os"

	"github.com/things-go/tools/cmd/astgen-dyn/internal/command"
)

func main() {
	err := command.NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
