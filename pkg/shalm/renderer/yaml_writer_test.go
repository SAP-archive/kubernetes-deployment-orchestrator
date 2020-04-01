package renderer

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("YamlWriter", func() {

	It("normalizes document separators", func() {
		buf := &bytes.Buffer{}
		writer := &YamlWriter{Writer: buf}
		writer.Write([]byte("\n\n---\ntest\n"))
		Expect(buf.String()).To(Equal("---\ntest\n"))
	})

	It("adds document separators", func() {
		buf := &bytes.Buffer{}
		writer := &YamlWriter{Writer: buf}
		writer.Write([]byte("test\n"))
		Expect(buf.String()).To(Equal("---\ntest\n"))
	})

	It("doesn't add separator to empty doc", func() {
		buf := &bytes.Buffer{}
		writer := &YamlWriter{Writer: buf}
		writer.Write([]byte("    "))
		Expect(buf.String()).To(Equal(""))
	})

})
