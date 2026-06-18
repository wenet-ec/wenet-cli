// internal/cli/login.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wenet-ec/wenet-cli/internal/config"
)

func newLoginCommand(opts *options) *cobra.Command {
	var server string

	cmd := &cobra.Command{
		Use:   "login <token>",
		Short: "Persist a WENet API token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token := args[0]
			if token == "" {
				return fmt.Errorf("token is required")
			}

			path, err := config.DefaultPath()
			if err != nil {
				return err
			}

			file, err := config.LoadFile(path)
			if err != nil {
				return err
			}
			file.Profiles[opts.profile] = config.Profile{
				Server: server,
				Token:  token,
			}
			if err := config.SaveFile(path, file); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Saved profile %q to %s\n", opts.profile, path)
			return nil
		},
	}
	cmd.Flags().StringVar(&server, "server", config.DefaultServer, "WENet public API server")

	return cmd
}
