package shalm

import (
	"bytes"

	"github.com/k14s/starlark-go/starlark"

	"github.com/blang/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/wonderix/shalm/pkg/shalm/test"
)

var _ = Describe("Chart Proxy", func() {

	Context("proxies apply and delete calls", func() {
		thread := &starlark.Thread{Name: "test"}
		repo, _ := NewRepo()
		var dir TestDir
		var chart ChartValue
		BeforeEach(func() {
			dir = NewTestDir()
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			dir.WriteFile("values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
			args := starlark.Tuple{starlark.String("hello")}
			kwargs := []starlark.Tuple{{starlark.String("key"), starlark.String("value")}}
			impl, err := newChart(thread, repo, dir.Root(), WithArgs(args), WithKwArgs(kwargs))
			Expect(err).NotTo(HaveOccurred())
			chart, err = newChartProxy(impl, "http://test.com", ProxyModeLocal, args, kwargs)
			Expect(err).NotTo(HaveOccurred())

		})
		AfterEach(func() {
			dir.Remove()

		})

		It("applies a ShalmChart to k8s", func() {
			buffer := &bytes.Buffer{}
			k := &FakeK8s{
				ApplyStub: func(cb ObjectStream, options *K8sOptions) error {
					return cb.Encode()(buffer)
				},
				ConfigContentStub: func() *string {
					result := "hello"
					return &result
				},
			}
			k.ForSubChartStub = func(s string, app string, version semver.Version, children int) K8s {
				return k
			}
			k.ForConfigStub = func(config string) (K8s, error) {
				return k, nil
			}
			err := chart.Apply(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.ApplyCallCount()).To(Equal(1))
			Expect(buffer.String()).To(ContainSubstring(`{"apiVersion":"wonderix.github.com/v1alpha2","kind":"ShalmChart","metadata":{"creationTimestamp":null,"name":"mariadb","namespace":"default"},"spec":{"values":{"replicas":"1","timeout":"30s"},"args":["hello"],"kwargs":{"key":"value"},"kubeconfig":"hello","namespace":"default","chart_url":"http://test.com","tool":"kubectl"},"status":{"lastOp":{"type":"","progress":0}}}`))
		})
		It("deletes a ShalmChart from k8s", func() {
			k := &FakeK8s{}
			k.ForSubChartStub = func(s string, app string, version semver.Version, children int) K8s {
				return k
			}
			err := chart.Delete(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.DeleteObjectCallCount()).To(Equal(1))
			kind, name, _ := k.DeleteObjectArgsForCall(0)
			Expect(kind).To(Equal("ShalmChart"))
			Expect(name).To(Equal("mariadb"))
		})
	})
	Context("isValidShalmURL", func() {
		It("works", func() {
			Expect(isValidShalmURL("test")).To(BeFalse())
			Expect(isValidShalmURL("/test/")).To(BeFalse())
			Expect(isValidShalmURL("http://test/")).To(BeTrue())
			Expect(isValidShalmURL("https://test/")).To(BeTrue())
			Expect(isValidShalmURL("ftp://test/")).To(BeFalse())
		})
	})

})
