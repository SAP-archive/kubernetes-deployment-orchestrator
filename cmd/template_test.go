package cmd

import (
	"bytes"
	"path"

	"github.com/wonderix/shalm/pkg/shalm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template Chart", func() {

	It("template the correct output", func() {
		writer := &bytes.Buffer{}
		err := template(path.Join(example, "cf"), shalm.NewK8sInMemoryEmpty())(writer)
		Expect(err).ToNot(HaveOccurred())
		output := writer.String()
		Expect(output).To(ContainSubstring("CREATE OR REPLACE USER 'uaa'"))
	})
})
