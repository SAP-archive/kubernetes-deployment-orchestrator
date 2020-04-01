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

var _ VaultBackend = (*testBackend)(nil)

func (u *testBackend) Name() string {
	return "vault"
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

func makeMyVault(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	c := &testBackend{}
	var name string
	err := starlark.UnpackArgs("myvault", args, kwargs, "name", &name, "prefix", &c.prefix)
	if err != nil {
		return nil, err
	}
	return NewVault(c, name)
}

var _ = Describe("generic vault", func() {

	Context("generic vault value", func() {
		v, err := makeMyVault(nil, nil, starlark.Tuple{starlark.String("name"), starlark.String("prefix")}, nil)
		vault := v.(*vault)

		It("behaves like starlark value", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(vault.String()).To(ContainSubstring("vault(name = name)"))
			Expect(func() { vault.Hash() }).Should(Panic())
			Expect(vault.Type()).To(Equal("vault"))
			Expect(vault.Truth()).To(BeEquivalentTo(false))

			By("attribute name", func() {
				value, err := vault.Attr("name")
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal(starlark.String("name")))
				Expect(vault.AttrNames()).To(ContainElement("name"))

			})

			By("attribute field", func() {
				value, err := vault.Attr("field")
				Expect(err).NotTo(HaveOccurred())
				Expect(value.(starlark.String).GoString()).To(ContainSubstring("prefix"))
				Expect(vault.AttrNames()).To(ContainElement("field"))

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
			err := vault.read(k8s)
			Expect(err).NotTo(HaveOccurred())

		})
	})

})
