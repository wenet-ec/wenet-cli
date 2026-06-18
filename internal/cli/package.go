// internal/cli/package.go
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wenet-ec/wenet-cli/internal/archive"
)

func newPackageCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "package",
		Short: "Build a deployment archive locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}

			result, err := archive.Build(root, output)
			if err != nil {
				return err
			}

			rel, err := filepath.Rel(root, result.Path)
			if err != nil {
				rel = result.Path
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s (%d files)\n", rel, result.FileCount)
			return nil
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "archive output path")

	return cmd
}
