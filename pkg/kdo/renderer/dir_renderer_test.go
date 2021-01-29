package renderer

import (
	"bytes"
	"io"
	"os"

	. "github.com/sap/kubernetes-deployment-orchestrator/pkg/kdo/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DirRender", func() {

	fileRenderer := func(filename string) func(io.Writer) error {
		return func(writer io.Writer) error {
			f, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(writer, f)
			return err
		}
	}

	Context("renders chart", func() {
		It("renders multipe files", func() {
			var err error
			dir := NewTestDir()
			defer dir.Remove()
			dir.WriteFile("test1.yaml", []byte("test: test1"), 0644)
			dir.WriteFile("test2.yml", []byte("test: test2"), 0644)

			writer := &bytes.Buffer{}
			err = DirRender("", DirSpec{dir.Root(), fileRenderer})(writer)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("---\ntest: test1\n---\ntest: test2\n"))
		})
		It("repects glob patterns", func() {
			var err error
			dir := NewTestDir()
			defer dir.Remove()
			dir.WriteFile("test1.yaml", []byte("test: test1"), 0644)
			dir.WriteFile("test3.yaml", []byte("test: test2"), 0644)
			writer := &bytes.Buffer{}
			err = DirRender("*[1-2].yaml", DirSpec{dir.Root(), fileRenderer})(writer)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("---\ntest: test1\n"))
		})

	})
})
