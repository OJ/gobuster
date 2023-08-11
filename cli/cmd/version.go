package cmd

import (
	"fmt"

	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

// nolint:gochecknoglobals
var cmdVersion *cobra.Command

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Println(libgobuster.VERSION)
	return nil
}

// nolint:gochecknoinits
func init() {
	cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "shows the current version",
		RunE:  runVersion,
	}

	rootCmd.AddCommand(cmdVersion)
}
