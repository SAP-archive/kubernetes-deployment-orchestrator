package shalm

import (
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

type testBackend struct {
	prefix string
}

var _ JewelBackend = (*testBackend)(nil)

func (u *testBackend) Name() string {
	return "jewel"
}

func (u *testBackend) Keys() map[string]string {
	return map[string]string{
		"field": "field.json",
	}
}

func (u *testBackend) Apply(m map[string][]byte) (map[string][]byte, error) {
	return map[string][]byte{
		"field.json": []byte(fmt.Sprintf("%s-%d", u.prefix, time.Now().Unix())),
	}, nil
}

func makeMyJewel(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c := &testBackend{}
	var name string
	err := starlark.UnpackArgs("myjewel", args, kwargs, "name", &name, "prefix", &c.prefix)
	if err != nil {
		return nil, err
	}
	return NewJewel(c, name)
}

var _ = Describe("generic jewel", func() {

	Context("generic jewel value", func() {
		v, err := makeMyJewel(nil, nil, starlark.Tuple{starlark.String("name"), starlark.String("prefix")}, nil)
		jewel := v.(*jewel)

		It("behaves like starlark value", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(jewel.String()).To(ContainSubstring("jewel(name = name)"))
			Expect(func() { jewel.Hash() }).Should(Panic())
			Expect(jewel.Type()).To(Equal("jewel"))
			Expect(jewel.Truth()).To(BeEquivalentTo(true))

			By("attribute name", func() {
				value, err := jewel.Attr("name")
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal(starlark.String("name")))
				Expect(jewel.AttrNames()).To(ContainElement("name"))

			})

			By("attribute field", func() {
				value, err := jewel.Attr("field")
				Expect(err).NotTo(HaveOccurred())
				Expect(value.(starlark.String).GoString()).To(ContainSubstring("prefix"))
				Expect(jewel.AttrNames()).To(ContainElement("field"))

			})

		})

		It("reads values from k8s", func() {
			k8s := &FakeK8s{
				GetStub: func(kind string, name string, k8s *K8sOptions) (*Object, error) {
					return nil, errors.New("NotFound")
				},
				IsNotExistStub: func(err error) bool {
					return true
				},
			}
			err := jewel.read(&vaultK8s{k8s: k8s})
			Expect(err).NotTo(HaveOccurred())

		})
	})

})
