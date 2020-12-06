package cmd

import (
	"fmt"
	"os"

	"github.com/wonderix/shalm/pkg/shalm"

	"github.com/k14s/starlark-go/starlark"

	"github.com/spf13/cobra"
)

var templateChartArgs = shalm.ChartOptions{}
var templateK8sArgs = shalm.K8sConfigs{}

var templateCmd = &cobra.Command{
	Use:   "template [chart]",
	Short: "template shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		k8s, err := k8s(applyK8sArgs.Merge(), shalm.WithProgressSubscription(func(progress int) {
			fmt.Printf("Progress  %d%%\n", progress)
		}))
		if err != nil {
			exit(err)
		}
		exit(template(args[0], k8s)(os.Stdout))
	},
}

func template(url string, k shalm.K8s) shalm.Stream {

	thread := &starlark.Thread{Name: "main", Load: rootExecuteOptions.load}
	repo, err := repo()
	if err != nil {
		return shalm.ErrorStream(err)
	}
	c, err := repo.Get(thread, url, templateChartArgs.Merge())

	if err != nil {
		return shalm.ErrorStream(err)
	}
	return c.Template(thread, k)

}

func init() {
	templateChartArgs.AddFlags(templateCmd.Flags())
	templateK8sArgs.AddFlags(templateCmd.Flags())
}
