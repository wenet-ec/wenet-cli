// internal/cli/deploy.go
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wenet-ec/wenet-cli/internal/api"
	"github.com/wenet-ec/wenet-cli/internal/config"
	"github.com/wenet-ec/wenet-cli/internal/project"
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
			secretScopeID, err := resolveSecretScope(client, pkg.Project, cfg.SecretScope)
			if err != nil {
				return err
			}
			rollout, err := client.CreateRollout(api.RolloutInput{
				PackageID:       pkg.ID,
				SecretScopeID:   secretScopeID,
				DownloadBaseDir: cfg.DownloadBaseDir,
				Cleanup:         *cfg.Cleanup,
				Targeting:       targetingPayload(cfg),
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created rollout %s for package %s:%s (%d deployments)\n", rollout.ID, cfg.Project, pkg.Tag, rollout.DeploymentCount)
			return nil
		},
	}
	addSourceFlags(cmd, &sourceOpts)
	return cmd
}

func resolveSecretScope(client *api.Client, projectID string, name string) (string, error) {
	if name == "" {
		return "", nil
	}
	scope, err := client.FindSecretScope(projectID, name)
	if err != nil {
		return "", err
	}
	if scope == nil {
		return "", fmt.Errorf("secret scope %q not found for project", name)
	}
	return scope.ID, nil
}

func targetingPayload(cfg *project.Config) map[string]any {
	switch {
	case cfg.All:
		return map[string]any{"type": "all"}
	case len(cfg.NodeIDs) > 0:
		return map[string]any{"type": "nodes", "node_ids": cfg.NodeIDs}
	case len(cfg.ClusterNames) > 0:
		return map[string]any{"type": "clusters", "cluster_names": cfg.ClusterNames}
	case len(cfg.Tags) > 0:
		return map[string]any{"type": "tags", "tag_names": cfg.Tags}
	default:
		// Config.Validate() rejects configs that reach here; this is unreachable
		// in normal operation but makes the exhaustiveness explicit.
		panic("targetingPayload: no targeting form set; call Config.Validate() before this")
	}
}
