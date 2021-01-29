package controllers

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o client_test.go sigs.k8s.io/controller-runtime/pkg/client.Client
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./fake_k8s_test.go ../pkg/kdo K8s

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	goruntime "runtime"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/k8s"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/kdo"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"

	kdov1a2 "github.com/sap/kubernetes-deployment-orchestrator/api/v1alpha2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var (
	_, b, _, _ = goruntime.Caller(0)
	basepath   = filepath.Dir(b)
	root       = path.Join(filepath.Dir(b), "..")
	example    = path.Join(root, "charts", "example", "simple")
)

var _ = Describe("KdoChartReconciler", func() {

	Context("KdoChartReconciler", func() {

		chartTgz, _ := ioutil.ReadFile(path.Join(example, "uaa-1.3.4.tgz"))

		var (
			chart      *kdov1a2.KdoChart
			buffer     *bytes.Buffer
			reconciler *KdoChartReconciler
			k          *k8s.FakeK8s
			k8sConfigs k8s.Configs
			recorder   *record.FakeRecorder
		)

		BeforeEach(func() {
			buffer = &bytes.Buffer{}
			repo, err := kdo.NewRepo()
			Expect(err).NotTo(HaveOccurred())

			k = &k8s.FakeK8s{
				ApplyStub: func(cb k8s.ObjectStream, options *k8s.Options) error {
					return cb.Encode()(buffer)
				},
				DeleteStub: func(cb k8s.ObjectStream, options *k8s.Options) error {
					return cb.Encode()(buffer)
				},
			}
			k.ForSubChartStub = func(s string, app string, version *semver.Version, children int) k8s.K8s {
				return k
			}
			k.WithContextStub = func(ctx context.Context) k8s.K8s {
				return k
			}
			k.GetStub = func(s string, s2 string, options *k8s.Options) (*k8s.Object, error) {
				return &k8s.Object{}, nil
			}

			clnt := &FakeClient{
				GetStub: func(ctx context.Context, name types.NamespacedName, object runtime.Object) error {
					switch object := object.(type) {
					case *kdov1a2.KdoChart:
						chart.DeepCopyInto(object)
						return nil
					}
					return apierrors.NewNotFound(schema.GroupResource{}, name.String())
				},
				UpdateStub: func(ctx context.Context, object runtime.Object, options ...client.UpdateOption) error {
					switch object := object.(type) {
					case *kdov1a2.KdoChart:
						object.DeepCopyInto(chart)
						return nil
					}
					return apierrors.NewNotFound(schema.GroupResource{}, "???")
				},
			}
			clnt.StatusStub = func() client.StatusWriter { return clnt }

			recorder = record.NewFakeRecorder(10)
			reconciler = &KdoChartReconciler{
				Client: clnt,
				Log:    ctrl.Log.WithName("reconciler"),
				Scheme: nil,
				Repo:   repo,
				K8s: func(configs ...k8s.Config) (k8s.K8s, error) {
					for _, config := range configs {
						config(&k8sConfigs)
					}
					return k, nil
				},
				Recorder: recorder,
			}
		})

		It("applies kdo chart correct", func() {

			chart = &kdov1a2.KdoChart{
				Spec: kdov1a2.ChartSpec{
					ChartTgz: chartTgz,
				},
			}

			_, err := reconciler.Reconcile(ctrl.Request{})
			k8sConfigs.Progress(100)
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(ContainSubstring(`"stringData":{"password":"test","username":"test"}`))
			Expect(chart.ObjectMeta.Finalizers).To(ContainElement("controller.kdo.sap.github.com"))
			Expect(chart.Status.LastOp.Progress).To(Equal(100))
			Expect(k.ApplyCallCount()).To(Equal(1))
		})

		It("handles error correct during apply", func() {
			chart = &kdov1a2.KdoChart{
				Spec: kdov1a2.ChartSpec{
					ChartTgz: chartTgz,
				},
			}
			k.ApplyStub = func(cb k8s.ObjectStream, options *k8s.Options) error {
				return fmt.Errorf("Apply error")
			}
			_, err := reconciler.Reconcile(ctrl.Request{})
			Expect(err).To(HaveOccurred())
			Expect(chart.Status.LastOp.Progress).To(Equal(0))
			Expect(chart.Status.LastOp.Type).To(Equal(applyErrorStatus))
			event := <-recorder.Events
			Expect(event).To(ContainSubstring("Apply error"))
		})

		It("deletes kdo chart correct", func() {

			chart = &kdov1a2.KdoChart{
				ObjectMeta: v1.ObjectMeta{
					Finalizers: []string{"controller.kdo.sap.github.com"},
					DeletionTimestamp: &v1.
						Time{Time: time.Now()},
				},
				Spec: kdov1a2.ChartSpec{
					ChartTgz: chartTgz,
				},
			}

			_, err := reconciler.Reconcile(ctrl.Request{})
			Expect(err).NotTo(HaveOccurred())
			Expect(chart.ObjectMeta.Finalizers).NotTo(ContainElement("controller.kdo.sap.github.com"))
			Expect(k.DeleteCallCount()).To(Equal(1))
			Expect(buffer.String()).To(ContainSubstring(`"stringData":{"password":"test","username":"test"}`))
		})
		It("handles error correct during delete", func() {
			chart = &kdov1a2.KdoChart{
				ObjectMeta: v1.ObjectMeta{
					Finalizers: []string{"controller.kdo.sap.github.com"},
					DeletionTimestamp: &v1.
						Time{Time: time.Now()},
				},
				Spec: kdov1a2.ChartSpec{
					ChartTgz: chartTgz,
				},
			}
			k.DeleteStub = func(cb k8s.ObjectStream, options *k8s.Options) error {
				return fmt.Errorf("delete error")
			}
			_, err := reconciler.Reconcile(ctrl.Request{})
			Expect(err).To(HaveOccurred())
			Expect(chart.Status.LastOp.Progress).To(Equal(0))
			Expect(chart.Status.LastOp.Type).To(Equal(deleteErrorStatus))
			event := <-recorder.Events
			Expect(event).To(ContainSubstring("delete error"))
		})
	})

	Context("Predicate", func() {
		It("works", func() {
			predicate := kdoChartPredicate{}
			chart := &kdov1a2.KdoChart{
				Spec: kdov1a2.ChartSpec{},
			}
			chart2 := &kdov1a2.KdoChart{
				Spec: kdov1a2.ChartSpec{
					Values: runtime.RawExtension{Raw: []byte("{}")},
				},
			}
			Expect(predicate.Create(event.CreateEvent{})).To(Equal(true))
			Expect(predicate.Delete(event.DeleteEvent{})).To(Equal(false))
			Expect(predicate.Generic(event.GenericEvent{})).To(Equal(false))
			Expect(predicate.Update(event.UpdateEvent{ObjectOld: chart, ObjectNew: chart})).To(Equal(true))
			Expect(predicate.Update(event.UpdateEvent{ObjectOld: chart, ObjectNew: chart2})).To(Equal(true))
		})

	})

})
