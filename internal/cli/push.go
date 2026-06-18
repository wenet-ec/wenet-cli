// internal/cli/push.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wenet-ec/wenet-cli/internal/api"
	"github.com/wenet-ec/wenet-cli/internal/config"
)

func newPushCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Build and upload a package",
		RunE: func(cmd *cobra.Command, args []string) error {
			cred, err := config.ResolveCredential(opts.profile)
			if err != nil {
				return err
			}
			client := api.NewClient(cred.Server, cred.Token)
			return fmt.Errorf("push is not wired yet; package upload needs the public API endpoint contract for %s", client.BaseURL())
		},
	}
}
