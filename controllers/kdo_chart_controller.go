/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"reflect"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/k14s/starlark-go/starlark"
	"github.com/pkg/errors"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/k8s"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/kdo"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"

	kdov1a2 "github.com/sap/kubernetes-deployment-orchestrator/api/v1alpha2"
)

var (
	myFinalizerName   = "controller.kdo.sap.github.com"
	applyStatus       = "apply"
	applyErrorStatus  = "apply-error"
	deleteStatus      = "delete"
	deleteErrorStatus = "delete-error"
)

// KdoChartReconciler reconciles a KdoChart object
type KdoChartReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Repo     kdo.Repo
	K8s      func(configs ...k8s.Config) (k8s.K8s, error)
	Load     func(thread *starlark.Thread, module string) (dict starlark.StringDict, err error)
	Recorder record.EventRecorder
}

type kdoChartPredicate struct {
}

// +kubebuilder:rbac:groups=kdo.sap.github.com,resources=kdocharts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kdo.sap.github.com,resources=kdocharts/status,verbs=get;update;patch

// Reconcile -
func (r *KdoChartReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	result := ctrl.Result{}
	ctx := context.Background()
	_ = r.Log.WithValues("kdochart", req.NamespacedName)

	var kdoChart kdov1a2.KdoChart
	err := r.Client.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, &kdoChart)
	if err != nil {
		return result, client.IgnoreNotFound(err)
	}
	if kdoChart.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(kdoChart.ObjectMeta.Finalizers, myFinalizerName) {
			kdoChart.ObjectMeta.Finalizers = append(kdoChart.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(context.Background(), &kdoChart); err != nil {
				return result, errors.Wrapf(err, "error updating KdoChart %s", req.NamespacedName.String())
			}
			kdoChart.Status.LastOp = kdov1a2.Operation{Type: applyStatus, Progress: 0}
			if err := r.Status().Update(context.Background(), &kdoChart); err != nil {
				return result, errors.Wrapf(err, "error updating status of KdoChart %s", req.NamespacedName.String())
			}
		}
		if err := r.apply(&kdoChart.Spec, func(progress int) {
			kdoChart.Status.LastOp = kdov1a2.Operation{Type: applyStatus, Progress: progress}
			r.Status().Update(context.Background(), &kdoChart)
		}); err != nil {
			err = errors.Wrapf(err, "error applying KdoChart %s", req.NamespacedName.String())
			r.Recorder.Event(&kdoChart, corev1.EventTypeWarning, "ApplyError", err.Error())
			kdoChart.Status.LastOp = kdov1a2.Operation{Type: applyErrorStatus, Progress: 0}
			r.Status().Update(context.Background(), &kdoChart)
			return result, err
		}
		return result, err
	}
	if len(kdoChart.ObjectMeta.Finalizers) == 1 && containsString(kdoChart.ObjectMeta.Finalizers, myFinalizerName) {
		kdoChart.Status.LastOp = kdov1a2.Operation{Type: deleteStatus, Progress: 0}
		if err := r.Status().Update(context.Background(), &kdoChart); err != nil {
			return result, errors.Wrapf(err, "error updating status of KdoChart %s", req.NamespacedName.String())
		}
		if err := r.delete(&kdoChart.Spec, func(progress int) {
			kdoChart.Status.LastOp = kdov1a2.Operation{Type: deleteStatus, Progress: progress}
			r.Status().Update(context.Background(), &kdoChart)
		}); err != nil {
			err = errors.Wrapf(err, "error deleting KdoChart %s", req.NamespacedName.String())
			r.Recorder.Event(&kdoChart, corev1.EventTypeWarning, "DeleteError", err.Error())
			kdoChart.Status.LastOp = kdov1a2.Operation{Type: deleteErrorStatus, Progress: 0}
			r.Status().Update(context.Background(), &kdoChart)
			return result, err
		}

		kdoChart.ObjectMeta.Finalizers = removeString(kdoChart.ObjectMeta.Finalizers, myFinalizerName)
		if err := r.Update(context.Background(), &kdoChart); err != nil {
			return result, errors.Wrapf(err, "error updating KdoChart %s", req.NamespacedName.String())
		}
	}

	return result, err

}

func (r *KdoChartReconciler) apply(spec *kdov1a2.ChartSpec, progressCb k8s.ProgressSubscription) error {
	var tool k8s.Tool
	if err := tool.Set(spec.Tool); err != nil {
		return err
	}
	k8s, err := r.K8s(k8s.WithKubeConfigContent(spec.KubeConfig),
		k8s.WithProgressSubscription(progressCb),
		k8s.WithTool(tool))
	if err != nil {
		return err
	}
	thread := &starlark.Thread{Name: "main", Load: r.Load}
	chart, err := r.Repo.GetFromSpec(thread, spec)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	return chart.Apply(thread, k8s.WithContext(ctx))
}

func (r *KdoChartReconciler) delete(spec *kdov1a2.ChartSpec, progressCb k8s.ProgressSubscription) error {
	var tool k8s.Tool
	if err := tool.Set(spec.Tool); err != nil {
		return err
	}
	k8s, err := r.K8s(k8s.WithKubeConfigContent(spec.KubeConfig),
		k8s.WithProgressSubscription(progressCb),
		k8s.WithTool(tool))
	if err != nil {
		return err
	}
	thread := &starlark.Thread{Name: "main", Load: r.Load}
	chart, err := r.Repo.GetFromSpec(thread, spec)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	return chart.Delete(thread, k8s.WithContext(ctx), &kdo.DeleteOptions{})
}

// SetupWithManager -
func (r *KdoChartReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kdov1a2.KdoChart{}).
		WithEventFilter(&kdoChartPredicate{}).
		WithOptions(options).
		Complete(r)
}

// Create -
func (r *kdoChartPredicate) Create(event.CreateEvent) bool {
	return true
}

// Delete -
func (r *kdoChartPredicate) Delete(event.DeleteEvent) bool {
	return false
}

// Update -
func (r *kdoChartPredicate) Update(ev event.UpdateEvent) bool {
	old := ev.ObjectOld.(*kdov1a2.KdoChart)
	new := ev.ObjectNew.(*kdov1a2.KdoChart)
	if !reflect.DeepEqual(old.Spec, new.Spec) {
		return true
	}
	if !containsString(new.ObjectMeta.Finalizers, myFinalizerName) {
		return true
	}
	if !new.ObjectMeta.DeletionTimestamp.IsZero() {
		return true
	}
	return false
}

// Generic -
func (r *kdoChartPredicate) Generic(event.GenericEvent) bool {
	return false
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
