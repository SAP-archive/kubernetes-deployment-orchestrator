package cmd

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/wonderix/shalm/pkg/k8s"
	"github.com/wonderix/shalm/pkg/shalm"

	"github.com/spf13/cobra"
)

var deleteChartArgs = shalm.ChartOptions{}
var deleteK8sArgs = k8s.K8sConfigs{}
var deleteOptions = shalm.DeleteOptions{}

var deleteCmd = &cobra.Command{
	Use:   "delete [chart]",
	Short: "delete shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		k8s, err := newK8s(deleteK8sArgs.Merge(), k8s.WithProgressSubscription(func(progress int) {
			fmt.Printf("Progress  %d%%\n", progress)
		}))
		if err != nil {
			exit(err)
		}
		exit(delete(args[0], k8s, &deleteOptions, deleteChartArgs.Merge()))
	},
}

func delete(url string, k k8s.K8s, deleteOpt *shalm.DeleteOptions, opts ...shalm.ChartOption) error {
	repo, err := repo()
	if err != nil {
		return err
	}
	thread := &starlark.Thread{Name: "main", Load: rootExecuteOptions.load}
	c, err := repo.Get(thread, url, opts...)
	if err != nil {
		return err
	}
	return c.Delete(thread, k, deleteOpt)
}

func init() {
	deleteChartArgs.AddFlags(deleteCmd.Flags())
	deleteK8sArgs.AddFlags(deleteCmd.Flags())
	rootOsbConfig.AddFlags(deleteCmd.Flags())
	deleteOptions.AddFlags(deleteCmd.Flags())
}
