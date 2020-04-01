package renderer

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/wonderix/shalm/pkg/shalm/test"
	"go.starlark.net/starlark"
)

var _ = Describe("ytt", func() {

	It("template file is working", func() {
		dir := NewTestDir()
		defer dir.Remove()
		dir.WriteFile("ytt.yaml", []byte("test: #@ self\n"), 0644)
		out := &bytes.Buffer{}
		renderer := YttFileRenderer(starlark.String("hello"))
		err := renderer(dir.Join("ytt.yaml"))(out)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out.Bytes())).To(Equal("test: hello\n"))
	})

	It("template is working", func() {
		in := bytes.NewBuffer([]byte("test: #@ self\n"))
		out := &bytes.Buffer{}
		err := yttRender(starlark.String("hello"), in, "stdin", out)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out.Bytes())).To(Equal("test: hello\n"))
	})
	It("loading of data is working", func() {
		in := bytes.NewBuffer([]byte(`
#@ load("@ytt:data", "data")
test: #@ data.values
`))
		out := &bytes.Buffer{}
		err := yttRender(starlark.String("hello"), in, "stdin", out)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out.Bytes())).To(Equal("test: hello\n"))
	})
	It("returns errors", func() {
		in := bytes.NewBuffer([]byte(`
#@ load("@ytt:data", "data")
test: #@ data.values.test
`))
		out := &bytes.Buffer{}
		err := yttRender(starlark.String("hello"), in, "stdin", out)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("string has no .test field or method"))
	})
})
