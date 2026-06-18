// internal/cli/deploy.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wenet-ec/wenet-cli/internal/api"
	"github.com/wenet-ec/wenet-cli/internal/config"
)

func newDeployCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Build, upload, and trigger a rollout",
		RunE: func(cmd *cobra.Command, args []string) error {
			cred, err := config.ResolveCredential(opts.profile)
			if err != nil {
				return err
			}
			client := api.NewClient(cred.Server, cred.Token)
			return fmt.Errorf("deploy is not wired yet; rollout creation needs the public API endpoint contract for %s", client.BaseURL())
		},
	}
}
