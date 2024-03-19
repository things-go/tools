package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type ginOpt struct {
	Input     []string
	Interface []string
}

type ginCmd struct {
	cmd *cobra.Command
	ginOpt
}

func newGinCmd() *ginCmd {
	root := &ginCmd{}
	cmd := &cobra.Command{
		Use:     "gin",
		Short:   "Generate gin http server from interface",
		Example: "ast-gen-dyn gin",
		RunE: func(cmd *cobra.Command, args []string) error {
			srcDir := root.Input[0]
			info, err := os.Stat(srcDir)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				srcDir = filepath.Dir(srcDir)
			}
			g := Generator{
				InputPattern:      root.Input,
				OutputDir:         srcDir,
				InterfacePatterns: root.Interface,
			}
			err = g.Parser()
			if err != nil {
				slog.Error("Error generating", slog.Any("error", err))
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&root.Input, "input", "i", []string{"."}, "")
	cmd.Flags().StringSliceVarP(&root.Interface, "interfaces", "I", nil, "")
	cmd.MarkFlagRequired("interfaces")
	root.cmd = cmd
	return root
}
