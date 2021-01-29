package kdo

import (
	"github.com/k14s/starlark-go/starlark"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

		maintainer := starlark.NewDict(1)
		maintainer.SetKey(starlark.String("name"), starlark.String("name"))
		err = cc.SetField("maintainers", starlark.NewList([]starlark.Value{maintainer}))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.Maintainers).To(HaveLen(1))
		Expect(cc.Maintainers[0]).To(HaveKeyWithValue("name", "name"))

		err = cc.SetField("sources", starlark.NewList([]starlark.Value{starlark.String("source")}))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.Sources).To(HaveLen(1))
		Expect(cc.Sources[0]).To(Equal("source"))

		err = cc.SetField("home", starlark.String("home"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.Home).To(Equal("home"))

		err = cc.SetField("icon", starlark.String("icon"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.Icon).To(Equal("icon"))

		err = cc.SetField("kube_version", starlark.String("kube_version"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.KubeVersion).To(Equal("kube_version"))

		err = cc.SetField("engine", starlark.String("engine"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.Engine).To(Equal("engine"))

		err = cc.SetField("app_version", starlark.String("app_version"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cc.AppVersion).To(Equal("app_version"))

		err = cc.SetField("invalid", starlark.String("invalid"))
		Expect(err).To(HaveOccurred())
	})

})
