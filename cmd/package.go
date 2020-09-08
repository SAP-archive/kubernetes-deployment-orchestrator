package cmd

import (
	"os"
	"path"

	"github.com/k14s/starlark-go/starlark"

	"github.com/spf13/cobra"
)

var helmFormat bool

var packageCmd = &cobra.Command{
	Use:   "package [chart]",
	Short: "package shalm chart",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exit(pkg(args[0]))
	},
}

func pkg(url string) error {
	repo, err := repo()
	if err != nil {
		return err
	}

	thread := &starlark.Thread{Name: "main", Load: rootExecuteOptions.load}
	c, err := repo.Get(thread, url)
	if err != nil {
		return err
	}
	out, err := os.Create(path.Base(c.GetName()) + "-" + c.GetVersion().String() + ".tgz")
	if err != nil {
		return err
	}
	defer out.Close()
	return c.Package(out, helmFormat)
}

func init() {
	packageCmd.Flags().BoolVar(&helmFormat, "helm", false, "package shalm chart as helm chart")
}
