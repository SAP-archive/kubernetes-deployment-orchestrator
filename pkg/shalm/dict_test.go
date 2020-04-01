package shalm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var _ = Describe("Dict", func() {

	It("SetField behaves like expected", func() {
		d := &dict{Dict: starlark.NewDict(10)}
		d.SetField("test", starlark.String("hello"))
		Expect(d.AttrNames()).To(ContainElement("test"))

		value, err := d.Attr("test")
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(starlark.String("hello")))

		_, err = d.Attr("undefined")
		Expect(err).To(HaveOccurred())

		value, err = d.Attr("keys")
		Expect(err).NotTo(HaveOccurred())
	})

	It("wraps correct", func() {
		d := wrapDict(starlark.NewDict(10))
		_, ok := d.(*dict)
		Expect(ok).To(BeTrue())

		d = wrapDict(starlark.String(""))
		_, ok = d.(starlark.String)
		Expect(ok).To(BeTrue())
	})

	It("unwraps correct", func() {
		d := unwrapDict(wrapDict(starlark.NewDict(10)))
		_, ok := d.(*starlark.Dict)
		Expect(ok).To(BeTrue())

		d = unwrapDict(starlark.String(""))
		_, ok = d.(starlark.String)
		Expect(ok).To(BeTrue())
	})

})
