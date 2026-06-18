// internal/cli/root.go
package cli

import "github.com/spf13/cobra"

const defaultProfile = "default"

type options struct {
	profile string
}

func NewRootCommand() *cobra.Command {
	opts := &options{profile: defaultProfile}

	cmd := &cobra.Command{
		Use:           "wenet",
		Short:         "Deploy projects to WENet",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVar(&opts.profile, "profile", defaultProfile, "credentials profile to use")

	cmd.AddCommand(newLoginCommand(opts))
	cmd.AddCommand(newLogoutCommand(opts))
	cmd.AddCommand(newPackageCommand())
	cmd.AddCommand(newPushCommand(opts))
	cmd.AddCommand(newDeployCommand(opts))

	return cmd
}
