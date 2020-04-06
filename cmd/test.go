package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"testing"

	"github.com/fatih/color"

	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"

	"github.com/wonderix/shalm/pkg/shalm"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"

	"github.com/spf13/cobra"
)

var once sync.Once

const maxDepth = 10

const namespace = "default"

var testK8s = func(configs ...shalm.K8sConfig) (shalm.K8s, error) {
	return shalm.NewK8sInMemory(namespace), nil
}

var testCmd = &cobra.Command{
	Use:   "test [chart]",
	Short: "test shalm charts",
	Long:  `test shalm charts using starlark`,
	Run: func(cmd *cobra.Command, args []string) {
		exit(test(args, shalm.NewK8sInMemory("test")))
	},
}

func env(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
	var name string
	err := starlark.UnpackArgs("env", args, kwargs, "name", &name)
	if err != nil {
		return starlark.None, err
	}
	return starlark.String(os.Getenv(name)), nil
}

func test(files []string, k shalm.K8s) error {
	t := &testing.T{}
	repo, _ := repo()
	thread := &starlark.Thread{Name: "main", Load: rootExecuteOptions.load}
	starlarktest.SetReporter(thread, t)
	var lastErr error
	testColor := color.New(color.Bold)
	testGreen := color.New(color.FgGreen, color.Bold)
	testRed := color.New(color.FgRed, color.Bold)

	for _, file := range files {
		k8s, err := testK8s()
		if err != nil {
			return err
		}
		predeclared := starlark.StringDict{
			"env":    starlark.NewBuiltin("env", env),
			"chart":  starlark.NewBuiltin("chart", shalm.NewChartFunction(repo, path.Dir(file), shalm.WithNamespace(namespace))),
			"k8s":    shalm.NewK8sValue(k8s),
			"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
			"assert": &starlarkstruct.Module{
				Name: "assert",
				Members: starlark.StringDict{
					"fail": starlark.NewBuiltin("fail", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
						var msg string
						err := starlark.UnpackPositionalArgs("fail", args, kwargs, 1, &msg)
						if err != nil {
							return starlark.None, err
						}
						return starlark.None, errors.New(msg)
					}),
					"true": starlark.NewBuiltin("fail", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
						var cond bool
						var msg string
						err := starlark.UnpackPositionalArgs("true", args, kwargs, 2, &cond, &msg)
						if err != nil {
							return starlark.None, err
						}
						if cond {
							return starlark.None, nil
						}
						return starlark.None, errors.New(msg)
					}),
					"eq": starlark.NewBuiltin("eq", assertBinaryFunction("eq", func(v1 starlark.Value, v2 starlark.Value) (err error) {
						ok, err := starlark.CompareDepth(syntax.EQL, v1, v2, maxDepth)
						if !ok {
							err = fmt.Errorf("Values %s != %s", v1.String(), v2.String())
						}
						return
					})),
					"neq": starlark.NewBuiltin("neq", assertBinaryFunction("eq", func(v1 starlark.Value, v2 starlark.Value) (err error) {
						ok, err := starlark.CompareDepth(syntax.NEQ, v1, v2, maxDepth)
						if !ok {
							err = fmt.Errorf("Values %s == %s", v1.String(), v2.String())
						}
						return
					})),
				},
			},
		}
		testColor.Printf("Running test in %s", file)
		if _, err := starlark.ExecFile(thread, file, nil, predeclared); err != nil {
			if err, ok := err.(*starlark.EvalError); ok {
				lastErr = errors.New(err.Backtrace())
			}
			lastErr = errors.New("Test failed")
			testRed.Println("    ERROR")
			fmt.Println(unwrapEvalError(err).Error())
		} else {
			testGreen.Println("    OK")
		}
	}
	return lastErr
}

func assertBinaryFunction(name string, test func(starlark.Value, starlark.Value) error) func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		var v1 starlark.Value
		var v2 starlark.Value
		err := starlark.UnpackPositionalArgs("eq", args, kwargs, 2, &v1, &v2)
		if err != nil {
			return starlark.None, err
		}
		return starlark.None, test(v1, v2)
	}

}
