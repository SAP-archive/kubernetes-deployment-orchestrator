package cmd

import (
	"fmt"

	"github.com/sap/kubernetes-deployment-orchestrator/pkg/kdo"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", kdo.Version())
		fmt.Printf("Docker tag: %s\n", kdo.DockerTag())
	},
}
