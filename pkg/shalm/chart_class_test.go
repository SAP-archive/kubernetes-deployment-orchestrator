package shalm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var _ = Describe("chartClass", func() {

	It("behaves like starlark value", func() {
		cc := &chartClass{Name: "xxx", Version: "1.2.3"}
		Expect(cc.Validate()).NotTo(HaveOccurred())
		Expect(cc.String()).To(ContainSubstring("xxx"))
		Expect(cc.Type()).To(Equal("chart_class"))
		Expect(func() { cc.Hash() }).Should(Panic())
		Expect(cc.Truth()).To(BeEquivalentTo(true))
		Expect(cc.AttrNames()).To(ConsistOf("api_version", "name", "version", "description", "keywords", "home", "sources", "icon"))
		for _, attribute := range cc.AttrNames() {
			_, err := cc.Attr(attribute)
			Expect(err).NotTo(HaveOccurred())
		}
		_, err := cc.Attr("unknown")
		Expect(err).To(HaveOccurred())

		err = cc.SetField("api_version", starlark.String("api_version"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.APIVersion).To(Equal("api_version"))

		err = cc.SetField("name", starlark.String("name"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.Name).To(Equal("name"))

		err = cc.SetField("version", starlark.String("1.0.0"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.Version).To(Equal("1.0.0"))

		err = cc.SetField("description", starlark.String("description"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.Description).To(Equal("description"))

		err = cc.SetField("home", starlark.String("home"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.Home).To(Equal("home"))

		err = cc.SetField("icon", starlark.String("icon"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.Icon).To(Equal("icon"))

		err = cc.SetField("invalid", starlark.String("invalid"))
		Expect(err).To(HaveOccurred())
	})

})
