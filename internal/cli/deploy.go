// internal/cli/deploy.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wenet-ec/wenet-cli/internal/api"
	"github.com/wenet-ec/wenet-cli/internal/config"
	"github.com/wenet-ec/wenet-cli/internal/source"
)

func newDeployCommand(opts *options) *cobra.Command {
	sourceOpts := source.Options{}
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build, upload, and trigger a rollout",
		RunE: func(cmd *cobra.Command, args []string) error {
			cred, err := config.ResolveCredential(opts.profile)
			if err != nil {
				return err
			}
			sourceOpts = sourceOpts.MergeEnv(source.FromEnv())
			if err := sourceOpts.Validate(); err != nil {
				return err
			}
			client := api.NewClient(cred.Server, cred.Token)
			if sourceOpts.IsRemote() {
				return fmt.Errorf("deploy is not wired yet; repo import and rollout creation need the public API endpoint contract for %s", client.BaseURL())
			}
			if sourceOpts.IsPackageFile() {
				return fmt.Errorf("deploy is not wired yet; package file upload and rollout creation need the public API endpoint contract for %s", client.BaseURL())
			}
			return fmt.Errorf("deploy is not wired yet; rollout creation needs the public API endpoint contract for %s", client.BaseURL())
		},
	}
	addSourceFlags(cmd, &sourceOpts)
	return cmd
}
