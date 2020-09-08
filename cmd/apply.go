package cmd

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/wonderix/shalm/pkg/shalm"

	"github.com/spf13/cobra"
)

var applyChartArgs = shalm.ChartOptions{}
var applyK8sArgs = shalm.K8sConfigs{}

var k8s = func(configs ...shalm.K8sConfig) (shalm.K8s, error) {
	return shalm.NewK8s(configs...)
}

var applyCmd = &cobra.Command{
	Use:   "apply [chart]",
	Short: "apply shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		k8s, err := k8s(applyK8sArgs.Merge(), shalm.WithProgressSubscription(func(progress int) {
			fmt.Printf("Progress  %d%%\n", progress)
		}))
		if err != nil {
			exit(err)
		}
		exit(apply(args[0], k8s, applyChartArgs.Merge()))
	},
}

func apply(url string, k shalm.K8s, opts ...shalm.ChartOption) error {
	repo, err := repo()
	if err != nil {
		return err
	}
	thread := &starlark.Thread{Name: "main", Load: rootExecuteOptions.load}
	c, err := repo.Get(thread, url, opts...)
	if err != nil {
		return err
	}
	return c.Apply(thread, k)
}

func init() {
	applyChartArgs.AddFlags(applyCmd.Flags())
	applyK8sArgs.AddFlags(applyCmd.Flags())
}
