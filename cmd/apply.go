package cmd

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/k8s"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/kdo"

	"github.com/spf13/cobra"
)

var applyChartArgs = kdo.ChartOptions{}
var applyK8sArgs = k8s.Configs{}

var newK8s = func(configs ...k8s.Config) (k8s.K8s, error) {
	return k8s.NewK8s(configs...)
}

var applyCmd = &cobra.Command{
	Use:   "apply [chart]",
	Short: "apply kdo chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		k8s, err := newK8s(applyK8sArgs.Merge(), k8s.WithProgressSubscription(func(progress int) {
			fmt.Printf("Progress  %d%%\n", progress)
		}))
		if err != nil {
			exit(err)
		}
		exit(apply(args[0], k8s, applyChartArgs.Merge()))
	},
}

func apply(url string, k k8s.K8s, opts ...kdo.ChartOption) error {
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
	rootOsbConfig.AddFlags(applyCmd.Flags())
}
