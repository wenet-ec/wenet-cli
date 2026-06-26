// internal/cli/push.go
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wenet-ec/wenet-cli/internal/api"
	"github.com/wenet-ec/wenet-cli/internal/archive"
	"github.com/wenet-ec/wenet-cli/internal/config"
	"github.com/wenet-ec/wenet-cli/internal/project"
	"github.com/wenet-ec/wenet-cli/internal/source"
)

func newPushCommand(opts *options) *cobra.Command {
	sourceOpts := source.Options{}
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Build and upload a package",
		RunE: func(cmd *cobra.Command, args []string) error {
			cred, err := config.ResolveCredential(opts.profile)
			if err != nil {
				return err
			}
			sourceOpts = sourceOpts.MergeEnv(source.FromEnv())
			if err := sourceOpts.Validate(); err != nil {
				return err
			}
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			cfg, err := project.Load(root)
			if err != nil {
				return err
			}
			client := api.NewClient(cred.Server, cred.Token)
			pkg, err := pushPackage(client, root, cfg, sourceOpts)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Pushed package %s:%s (%s)\n", cfg.Project, pkg.Tag, pkg.ID)
			return nil
		},
	}
	addSourceFlags(cmd, &sourceOpts)
	return cmd
}

func pushPackage(client *api.Client, root string, cfg *project.Config, sourceOpts source.Options) (*api.Package, error) {
	projectRow, err := client.EnsureProject(cfg.Project)
	if err != nil {
		return nil, err
	}
	packageSource := api.PackageSource{
		SourceURL:   sourceOpts.URL,
		SourceRef:   sourceOpts.Ref,
		SourceToken: sourceOpts.Token,
	}
	if sourceOpts.IsPackageFile() {
		packageSource.FilePath = sourceOpts.PackageFile
	} else if !sourceOpts.IsRemote() {
		archiveResult, err := archive.Build(root, "")
		if err != nil {
			return nil, err
		}
		packageSource.FilePath = archiveResult.Path
	}
	return client.PushPackage(projectRow.ID, cfg.Tag, packageSource)
}
