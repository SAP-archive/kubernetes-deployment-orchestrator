package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/k14s/starlark-go/starlark"

	"github.com/k14s/ytt/pkg/yttlibrary"
	"github.com/k14s/ytt/pkg/yttlibrary/overlay"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/extensions"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/k8s"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/kdo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var repoConfigFile string
var repoConfigFileDefault string
var rootOsbConfig = extensions.OsbConfig{}

func init() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	repoConfigFileDefault = path.Join(homedir, ".kdo", "config")
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(packageCmd)
	rootCmd.AddCommand(controllerCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.PersistentFlags().StringVar(&repoConfigFile, "config", repoConfigFileDefault, "kdo configuration file (e.g. credentials)")
}

func repo() (kdo.Repo, error) {
	if repoConfigFile == repoConfigFileDefault {
		if _, err := os.Stat(repoConfigFile); err != nil {
			return kdo.NewRepo()
		}
	}
	return kdo.NewRepo(kdo.WithConfigFile(repoConfigFile))
}

var rootCmd = &cobra.Command{
	Use:   "kdo",
	Short: "Kubernete deployment orchestrator brings the starlark language to helm charts",
	Long:  ``,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Invalid command %s", args[0])
	},
}

var rootExecuteOptions = ExecuteOptions{load: defaultLoad}

// ExecuteOptions -
type ExecuteOptions struct {
	load func(thread *starlark.Thread, module string) (dict starlark.StringDict, err error)
}

func defaultLoad(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	switch module {
	case "@ytt:base64":
		return yttlibrary.Base64API, nil
	case "@ytt:json":
		return yttlibrary.JSONAPI, nil
	case "@ytt:md5":
		return yttlibrary.MD5API, nil
	case "@ytt:regexp":
		return yttlibrary.RegexpAPI, nil
	case "@ytt:sha256":
		return yttlibrary.SHA256API, nil
	case "@ytt:url":
		return yttlibrary.URLAPI, nil
	case "@ytt:yaml":
		return yttlibrary.YAMLAPI, nil
	case "@ytt:overlay":
		return overlay.API, nil
	case "@ytt:struct":
		return yttlibrary.StructAPI, nil
	case "@ytt:module":
		return yttlibrary.ModuleAPI, nil
	case "@kdo:bcrypt":
		return BcryptAPI, nil
	case "@kdo:osb":
		return extensions.OsbAPI(rootOsbConfig), nil
	}
	return nil, fmt.Errorf("Unknown module '%s'", module)
}

// ExecuteOption -
type ExecuteOption func(e *ExecuteOptions)

// WithModules add new modules to kdo, which can be loaded using load inside starlark scripts
func WithModules(load func(thread *starlark.Thread, module string) (starlark.StringDict, error)) ExecuteOption {
	return func(e *ExecuteOptions) {
		oldLoad := e.load
		e.load = func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
			dict, err := load(thread, module)
			if err != nil {
				return oldLoad(thread, module)
			}
			return dict, nil
		}
	}
}

// WithFlags adds global command line options
func WithFlags(flags func(flagsSet *pflag.FlagSet)) ExecuteOption {
	return func(e *ExecuteOptions) {
		flags(rootCmd.PersistentFlags())
	}
}

// WithApplyFlags adds global command line options to apply, delete and test
func WithApplyFlags(flags func(flagsSet *pflag.FlagSet)) ExecuteOption {
	return func(e *ExecuteOptions) {
		flags(applyCmd.Flags())
		flags(deleteCmd.Flags())
	}
}

// WithTestFlags adds global command line options to apply, delete and test
func WithTestFlags(flags func(flagsSet *pflag.FlagSet)) ExecuteOption {
	return func(e *ExecuteOptions) {
		flags(testCmd.Flags())
	}
}

// WithK8s overrides constructor for k8s
func WithK8s(k func(configs ...k8s.Config) (k8s.K8s, error)) ExecuteOption {
	return func(e *ExecuteOptions) {
		newK8s = k
	}
}

// WithTestK8s overrides constructor for k8s
func WithTestK8s(k func(configs ...k8s.Config) (k8s.K8s, error)) ExecuteOption {
	return func(e *ExecuteOptions) {
		testK8s = k
	}
}

// Execute executes the root command.
func Execute(executeOptions ...ExecuteOption) error {
	for _, eo := range executeOptions {
		eo(&rootExecuteOptions)
	}
	return rootCmd.Execute()
}

func unwrapEvalError(err error) error {
	if err == nil {
		return nil
	}
	evalError, ok := err.(*starlark.EvalError)
	if ok {
		return errors.New(evalError.Backtrace())
	}
	return err
}

func exit(err error) {
	if err != nil {
		fmt.Println(unwrapEvalError(err).Error())
		os.Exit(1)
	}
	os.Exit(0)
}
