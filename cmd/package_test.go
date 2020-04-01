package cmd

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Package Chart", func() {

	It("produces the correct output", func() {
		filename := "cf-11.6.3.tgz"
		defer func() {
			os.Remove(filename)
		}()
		err := pkg(path.Join(example, "cf"))
		Expect(err).ToNot(HaveOccurred())
		_, err = os.Stat(filename)
		Expect(err).ToNot(HaveOccurred())
	})
})
