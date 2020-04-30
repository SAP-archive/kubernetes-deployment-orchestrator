package cmd

import (
	"bytes"
	"path"

	"github.com/blang/semver"
	"github.com/wonderix/shalm/pkg/shalm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete Chart", func() {

	It("produces the correct output", func() {
		writer := bytes.Buffer{}
		k := &FakeK8s{
			DeleteStub: func(i shalm.ObjectStream, options *shalm.K8sOptions) error {
				i.Encode()(&writer)
				return nil
			},
		}
		k.ForSubChartStub = func(s string, app string, version semver.Version, children int) shalm.K8s {
			return k
		}

		err := delete(path.Join(example, "cf"), k)
		Expect(err).ToNot(HaveOccurred())
		output := writer.String()
		Expect(output).To(ContainSubstring("CREATE OR REPLACE USER 'uaa'"))
		Expect(k.DeleteCallCount()).To(Equal(3))
	})
})
