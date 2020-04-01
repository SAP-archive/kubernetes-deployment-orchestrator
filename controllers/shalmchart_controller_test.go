package controllers

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o client_test.go sigs.k8s.io/controller-runtime/pkg/client.Client
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./fake_k8s_test.go ../pkg/shalm K8s

import (
	"bytes"
	"context"
	"io/ioutil"
	"path"
	"path/filepath"
	goruntime "runtime"
	"time"

	"github.com/blang/semver"
	"github.com/wonderix/shalm/pkg/shalm"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	shalmv1a2 "github.com/wonderix/shalm/api/v1alpha2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	_, b, _, _ = goruntime.Caller(0)
	basepath   = filepath.Dir(b)
	root       = path.Join(filepath.Dir(b), "..")
	example    = path.Join(root, "charts", "example", "simple")
)

var _ = Describe("ShalmChartReconciler", func() {

	chartTgz, _ := ioutil.ReadFile(path.Join(example, "mariadb-6.12.2.tgz"))

	It("applies shalm chart correct", func() {

		buffer := &bytes.Buffer{}
		k8s := &FakeK8s{
			ApplyStub: func(cb shalm.ObjectStream, options *shalm.K8sOptions) error {
				return cb.Encode()(buffer)
			},
		}
		k8s.ForSubChartStub = func(s string, app string, version semver.Version) shalm.K8s {
			return k8s
		}
		chart := shalmv1a2.ShalmChart{
			Spec: shalmv1a2.ChartSpec{
				ChartTgz: chartTgz,
			},
		}

		clnt := &FakeClient{
			GetStub: func(ctx context.Context, name types.NamespacedName, object runtime.Object) error {
				switch object := object.(type) {
				case *shalmv1a2.ShalmChart:
					chart.DeepCopyInto(object)
					return nil
				}
				return apierrors.NewNotFound(schema.GroupResource{}, name.String())
			},
			UpdateStub: func(ctx context.Context, object runtime.Object, options ...client.UpdateOption) error {
				switch object := object.(type) {
				case *shalmv1a2.ShalmChart:
					object.DeepCopyInto(&chart)
					return nil
				}
				return apierrors.NewNotFound(schema.GroupResource{}, "???")
			},
		}
		clnt.StatusStub = func() client.StatusWriter { return clnt }
		var k8sConfigs shalm.K8sConfigs
		repo, err := shalm.NewRepo()
		Expect(err).NotTo(HaveOccurred())
		reconciler := ShalmChartReconciler{
			Client: clnt,
			Log:    ctrl.Log.WithName("reconciler"),
			Scheme: nil,
			Repo:   repo,
			K8s: func(configs ...shalm.K8sConfig) (shalm.K8s, error) {
				for _, config := range configs {
					config(&k8sConfigs)
				}
				return k8s, nil
			},
		}
		_, err = reconciler.Reconcile(ctrl.Request{})
		k8sConfigs.Progress(100)
		Expect(err).NotTo(HaveOccurred())
		Expect(buffer.String()).To(ContainSubstring(`"serviceName":"mariadb-master"`))
		Expect(chart.ObjectMeta.Finalizers).To(ContainElement("controller.shalm.wonderix.github.com"))
		Expect(chart.Status.LastOp.Progress).To(Equal(100))
		Expect(k8s.ApplyCallCount()).To(Equal(1))
	})
	It("deletes shalm chart correct", func() {

		buffer := &bytes.Buffer{}
		k8s := &FakeK8s{
			DeleteStub: func(cb shalm.ObjectStream, options *shalm.K8sOptions) error {
				return cb.Encode()(buffer)
			},
		}
		k8s.ForSubChartStub = func(s string, app string, version semver.Version) shalm.K8s {
			return k8s
		}
		chart := shalmv1a2.ShalmChart{
			ObjectMeta: v1.ObjectMeta{
				Finalizers: []string{"controller.shalm.wonderix.github.com"},
				DeletionTimestamp: &v1.
					Time{Time: time.Now()},
			},
			Spec: shalmv1a2.ChartSpec{
				ChartTgz: chartTgz,
			},
		}

		clnt := &FakeClient{
			GetStub: func(ctx context.Context, name types.NamespacedName, object runtime.Object) error {
				switch object := object.(type) {
				case *shalmv1a2.ShalmChart:
					chart.DeepCopyInto(object)
					return nil
				}
				return apierrors.NewNotFound(schema.GroupResource{}, name.String())
			},
			UpdateStub: func(ctx context.Context, object runtime.Object, options ...client.UpdateOption) error {
				switch object := object.(type) {
				case *shalmv1a2.ShalmChart:
					object.DeepCopyInto(&chart)
					return nil
				}
				return apierrors.NewNotFound(schema.GroupResource{}, "???")
			},
		}
		clnt.StatusStub = func() client.StatusWriter { return clnt }
		repo, err := shalm.NewRepo()
		Expect(err).NotTo(HaveOccurred())
		reconciler := ShalmChartReconciler{
			Client: clnt,
			Log:    ctrl.Log.WithName("reconciler"),
			Scheme: nil,
			Repo:   repo,
			K8s: func(configs ...shalm.K8sConfig) (shalm.K8s, error) {
				return k8s, nil
			},
		}
		_, err = reconciler.Reconcile(ctrl.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(chart.ObjectMeta.Finalizers).NotTo(ContainElement("controller.shalm.wonderix.github.com"))
		Expect(k8s.DeleteCallCount()).To(Equal(1))
		Expect(buffer.String()).To(ContainSubstring(`"serviceName":"mariadb-master"`))
	})
})
