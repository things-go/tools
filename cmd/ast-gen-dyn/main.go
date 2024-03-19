package main

import (
	"os"
)

func main() {
	err := NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
