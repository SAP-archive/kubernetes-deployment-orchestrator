package cmd

import (
	"fmt"
	"os"

	"github.com/wonderix/shalm/pkg/k8s"
	"github.com/wonderix/shalm/pkg/shalm"

	"github.com/k14s/starlark-go/starlark"

	"github.com/spf13/cobra"
)

var templateChartArgs = shalm.ChartOptions{}
var templateK8sArgs = k8s.Configs{}

var templateCmd = &cobra.Command{
	Use:   "template [chart]",
	Short: "template shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		k8s, err := newK8s(applyK8sArgs.Merge(), k8s.WithProgressSubscription(func(progress int) {
			fmt.Printf("Progress  %d%%\n", progress)
		}))
		if err != nil {
			exit(err)
		}
		exit(template(args[0], k8s)(os.Stdout))
	},
}

func template(url string, k k8s.K8s) k8s.Stream {

	thread := &starlark.Thread{Name: "main", Load: rootExecuteOptions.load}
	repo, err := repo()
	if err != nil {
		return k8s.ErrorStream(err)
	}
	c, err := repo.Get(thread, url, templateChartArgs.Merge())

	if err != nil {
		return k8s.ErrorStream(err)
	}
	return c.Template(thread, k)

}

func init() {
	templateChartArgs.AddFlags(templateCmd.Flags())
	templateK8sArgs.AddFlags(templateCmd.Flags())
}
