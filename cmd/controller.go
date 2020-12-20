package cmd

import (
	shalmv1a2 "github.com/wonderix/shalm/api/v1alpha2"
	"github.com/wonderix/shalm/pkg/k8s"

	"github.com/pkg/errors"
	"github.com/wonderix/shalm/controllers"

	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	control "sigs.k8s.io/controller-runtime/pkg/controller"
)

var controllerK8sArgs = k8s.Configs{}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "run in controller mode",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		exit(controller(ctrl.SetupSignalHandler()))
	},
}
var (
	setupLog      = ctrl.Log.WithName("setup")
	reconcilerLog = ctrl.Log.WithName("reconciler")
	options       = control.Options{MaxConcurrentReconciles: 3}
)

func controller(stopCh <-chan struct{}) error {

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	err = shalmv1a2.AddToScheme(mgr.GetScheme())
	if err != nil {
		return errors.Wrap(err, "unable to add shalm scheme")
	}
	repo, err := repo()
	if err != nil {
		return err
	}

	reconciler := &controllers.ShalmChartReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    reconcilerLog,
		Repo:   repo,
		K8s: func(configs ...k8s.Config) (k8s.K8s, error) {
			configs = append([]k8s.Config{controllerK8sArgs.Merge()}, configs...)
			return k8s.NewK8s(configs...)
		},
		Load:     rootExecuteOptions.load,
		Recorder: mgr.GetEventRecorderFor("shalmchart-controller"),
	}
	err = reconciler.SetupWithManager(mgr, options)
	if err != nil {
		return errors.Wrap(err, "unable to create controller")
	}

	err = ctrl.NewWebhookManagedBy(mgr).
		For(&shalmv1a2.ShalmChart{}).
		Complete()
	if err != nil {
		return errors.Wrap(err, "unable to create webhook")
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(stopCh); err != nil {
		return errors.Wrap(err, "problem running manager")
	}
	return nil
}

func init() {
	controllerCmd.Flags().IntVar(&options.MaxConcurrentReconciles, "concurrent-reconciles", options.MaxConcurrentReconciles, "Number of concurrent reconciles")
	controllerK8sArgs.AddFlags(controllerCmd.Flags())
}
