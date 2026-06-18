// internal/cli/logout.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wenet-ec/wenet-cli/internal/config"
)

func newLogoutCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove a saved WENet API token",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}

			file, err := config.LoadFile(path)
			if err != nil {
				return err
			}
			if _, ok := file.Profiles[opts.profile]; !ok {
				return fmt.Errorf("profile %q not found", opts.profile)
			}
			delete(file.Profiles, opts.profile)

			if err := config.SaveFile(path, file); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed profile %q from %s\n", opts.profile, path)
			return nil
		},
	}
}
