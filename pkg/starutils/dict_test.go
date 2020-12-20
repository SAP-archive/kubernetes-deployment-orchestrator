package starutils

import (
	"github.com/k14s/starlark-go/starlark"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		d := WrapDict(starlark.NewDict(10))
		_, ok := d.(*dict)
		Expect(ok).To(BeTrue())

		d = WrapDict(starlark.String(""))
		_, ok = d.(starlark.String)
		Expect(ok).To(BeTrue())
	})

	It("unwraps correct", func() {
		d := UnwrapDict(WrapDict(starlark.NewDict(10)))
		_, ok := d.(*starlark.Dict)
		Expect(ok).To(BeTrue())

		d = UnwrapDict(starlark.String(""))
		_, ok = d.(starlark.String)
		Expect(ok).To(BeTrue())
	})

})
