// internal/cli/source_flags.go
package cli

import (
	"github.com/spf13/cobra"
	"github.com/wenet-ec/wenet-cli/internal/source"
)

func addSourceFlags(cmd *cobra.Command, opts *source.Options) {
	cmd.Flags().StringVar(&opts.PackageFile, "package-file", "", "existing .tar.gz package archive to upload instead of building from cwd")
	cmd.Flags().StringVar(&opts.URL, "source-url", "", "HTTPS Git clone URL for platform-side repo import")
	cmd.Flags().StringVar(&opts.Ref, "source-ref", "", "branch, tag, or full commit SHA to import")
	cmd.Flags().StringVar(&opts.Token, "source-token", "", "PAT or deploy token for private repo import")
}
