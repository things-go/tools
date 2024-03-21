package command

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type ginOpt struct {
	Pattern []string
	GinGenOption
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
			srcDir := root.Pattern[0]
			fileInfo, err := os.Stat(srcDir)
			if err != nil {
				return err
			}
			if !fileInfo.IsDir() {
				srcDir = filepath.Dir(srcDir)
			}
			g := GinGen{
				Pattern:      root.Pattern,
				OutputDir:    srcDir,
				GinGenOption: root.GinGenOption,
				Processed:    map[string]struct{}{},
			}
			err = g.Generate()
			if err != nil {
				slog.Error("Error generating", slog.Any("error", err))
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&root.Pattern, "pattern", "p", []string{"."}, "the list of files or a directory.")
	cmd.Flags().StringSliceVarP(&root.Interface, "interface", "i", nil, "the list of interface names; must be set")
	cmd.Flags().BoolVar(&root.AllowDeleteBody, "allow_delete_body", false, "allow delete body")
	cmd.Flags().BoolVar(&root.AllowEmptyPatchBody, "allow_empty_patch_body", false, "allow empty patch body")
	cmd.Flags().BoolVar(&root.UseEncoding, "use_encoding", false, "use the framework encoding")
	cmd.MarkFlagRequired("interface")

	root.cmd = cmd
	return root
}
