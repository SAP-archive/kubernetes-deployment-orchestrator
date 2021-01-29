package kdo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Injected Files", func() {

	It("behaves like starlark value", func() {
		f := &injectedFiles{files: []string{"hello"}}
		Expect(f.String()).To(ContainSubstring("hello"))
		Expect(f.Type()).To(Equal("injected_files"))
		Expect(func() { f.Hash() }).Should(Panic())
		Expect(f.Truth()).To(BeEquivalentTo(false))
		_, err := f.Attr("test")
		Expect(err).To(HaveOccurred())
		Expect(f.AttrNames()).To(BeEmpty())
	})

})
