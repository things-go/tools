package command

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

var validMode = []string{"gin", "resty", "both"}

type ginOpt struct {
	Pattern []string
	GinGenOption
}
type RootCmd struct {
	cmd *cobra.Command
	ginOpt
	level string
}

func NewRootCmd() *RootCmd {
	root := &RootCmd{}
	cmd := &cobra.Command{
		Use:       "astgen-dyn",
		Short:     "generate http server and client tools",
		Long:      "generate http server and client tools",
		Version:   BuildVersion(),
		ValidArgs: []string{},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !slices.Contains(validMode, root.Mode) {
				return fmt.Errorf("[--mode|-m] flag must be one of %v", validMode)
			}

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
		SilenceUsage:  false,
		SilenceErrors: false,
		Args:          cobra.NoArgs,
	}
	cobra.OnInitialize(func() {
		textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   false,
			Level:       level(root.level),
			ReplaceAttr: nil,
		})
		slog.SetDefault(slog.New(textHandler))
	})

	cmd.PersistentFlags().StringVarP(&root.level, "level", "l", "info", "log level(debug,info,warn,error)")
	cmd.Flags().StringSliceVarP(&root.Pattern, "pattern", "p", []string{"."}, "the list of files or a directory.")
	cmd.Flags().StringSliceVarP(&root.Interface, "interface", "i", nil, "the list of interface names; must be set")
	cmd.Flags().StringVarP(&root.Mode, "mode", "m", "both", "set generate mode, one of (gin,resty,both)")
	cmd.Flags().BoolVar(&root.AllowDeleteBody, "allow_delete_body", false, "allow delete body")
	cmd.Flags().BoolVar(&root.AllowEmptyPatchBody, "allow_empty_patch_body", false, "allow empty patch body")
	cmd.Flags().BoolVar(&root.UseEncoding, "use_encoding", false, "use the framework encoding")
	cmd.MarkFlagRequired("interface")
	cmd.RegisterFlagCompletionFunc("mode", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validMode, cobra.ShellCompDirectiveDefault
	})
	root.cmd = cmd
	return root
}

// Execute adds all child commands to the root command and sets flags appropriately.
func (r *RootCmd) Execute() error {
	return r.cmd.Execute()
}

func level(s string) slog.Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
