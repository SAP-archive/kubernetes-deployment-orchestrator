package cmd

import (
	"os"

	"github.com/wonderix/shalm/pkg/shalm"

	"go.starlark.net/starlark"

	"github.com/spf13/cobra"
)

var templateChartArgs = shalm.ChartOptions{}

var templateCmd = &cobra.Command{
	Use:   "template [chart]",
	Short: "template shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exit(template(args[0])(os.Stdout))
	},
}

func template(url string) shalm.Stream {
	thread := &starlark.Thread{Name: "main", Load: rootExecuteOptions.load}
	repo, err := repo()
	if err != nil {
		return shalm.ErrorStream(err)
	}
	c, err := repo.Get(thread, url, templateChartArgs.Merge())

	if err != nil {
		return shalm.ErrorStream(err)
	}
	return c.Template(thread)

}

func init() {
	templateChartArgs.AddFlags(templateCmd.Flags())
}
