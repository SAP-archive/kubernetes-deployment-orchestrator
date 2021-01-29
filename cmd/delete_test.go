package cmd

import (
	"bytes"
	"path"

	semver "github.com/Masterminds/semver/v3"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/k8s"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/kdo"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete Chart", func() {

	It("produces the correct output", func() {
		Skip("unsupported")
		writer := bytes.Buffer{}
		k := &k8s.FakeK8s{
			DeleteStub: func(i k8s.ObjectStream, options *k8s.Options) error {
				i.Encode()(&writer)
				return nil
			},
		}
		k.ForSubChartStub = func(s string, app string, version *semver.Version, children int) k8s.K8s {
			return k
		}

		err := delete(path.Join(example, "cf"), k, &kdo.DeleteOptions{})
		Expect(err).ToNot(HaveOccurred())
		output := writer.String()
		Expect(output).To(ContainSubstring("CREATE OR REPLACE USER 'uaa'"))
		Expect(k.DeleteCallCount()).To(Equal(3))
	})
})
