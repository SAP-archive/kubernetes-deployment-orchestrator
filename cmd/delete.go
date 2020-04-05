package cmd

import (
	"fmt"

	"github.com/wonderix/shalm/pkg/shalm"
	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var deleteChartArgs = shalm.ChartOptions{}
var deleteK8sArgs = shalm.K8sConfigs{}

var deleteCmd = &cobra.Command{
	Use:   "delete [chart]",
	Short: "delete shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		k8s, err := k8s(deleteK8sArgs.Merge(), shalm.WithProgressSubscription(func(progress int) {
			fmt.Printf("Progress  %d%%\n", progress)
		}))
		if err != nil {
			exit(err)
		}
		exit(delete(args[0], k8s, deleteChartArgs.Merge()))
	},
}

func delete(url string, k shalm.K8s, opts ...shalm.ChartOption) error {
	repo, err := repo()
	if err != nil {
		return err
	}
	opts = append(opts, shalm.WithK8s(k))
	thread := &starlark.Thread{Name: "main", Load: rootExecuteOptions.load}
	c, err := repo.Get(thread, url, opts...)
	if err != nil {
		return err
	}
	return c.Delete(thread)
}

func init() {
	deleteChartArgs.AddFlags(deleteCmd.Flags())
	deleteK8sArgs.AddFlags(deleteCmd.Flags())
}
