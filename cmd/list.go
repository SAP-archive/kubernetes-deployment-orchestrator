package cmd

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/wonderix/shalm/pkg/shalm"

	"github.com/spf13/cobra"
)

var listOptions = &shalm.RepoListOptions{}
var listK8sArgs = &shalm.K8sConfigs{}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list shalm charts",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		k8s, err := k8s(listK8sArgs.Merge())
		if err != nil {
			exit(err)
		}
		exit(list(k8s, listOptions))
	},
}

func list(k8s shalm.K8s, listOptions *shalm.RepoListOptions) error {
	repo, err := repo()
	if err != nil {
		return err
	}
	thread := &starlark.Thread{Name: "main", Load: rootExecuteOptions.load}
	charts, err := repo.List(thread, k8s, listOptions)
	if err != nil {
		return err
	}
	fmt.Printf("%-20s %-20s %-10s\n", "GENUS", "NAMESPACE", "VERSION")
	for _, c := range charts {
		fmt.Printf("%-20s %-20s %-10s\n", c.GetGenus(), c.GetNamespace(), c.GetVersion().String())
	}
	return nil
}

func init() {
	listOptions.AddFlags(listCmd.Flags())
	listK8sArgs.AddFlags(listCmd.Flags())
}
