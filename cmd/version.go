package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wonderix/shalm/pkg/shalm"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", shalm.Version())
		fmt.Printf("Docker tag: %s\n", shalm.DockerTag())
	},
}
