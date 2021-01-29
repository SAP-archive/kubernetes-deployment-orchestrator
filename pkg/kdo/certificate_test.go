package kdo

import (
	"github.com/k14s/starlark-go/starlark"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sap/kubernetes-deployment-orchestrator/pkg/k8s"
)

var _ = Describe("certificates", func() {

	Context("Self signed certifcate", func() {
		sca, _ := makeCertificate(nil, nil, starlark.Tuple{starlark.String("name")}, []starlark.Tuple{{starlark.String("is_ca"), starlark.Bool(true)}})
		ca := sca.(*jewel)

		It("behaves like starlark value", func() {
			Expect(ca.String()).To(ContainSubstring("name = name"))
			Expect(func() { ca.Hash() }).Should(Panic())
			Expect(ca.Type()).To(Equal("certificate"))
			Expect(ca.Truth()).To(BeEquivalentTo(true))

			By("attribute name", func() {
				value, err := ca.Attr("name")
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal(starlark.String("name")))
				Expect(ca.AttrNames()).To(ContainElement("name"))

			})

			By("attribute certificate", func() {
				value, err := ca.Attr("certificate")
				Expect(err).NotTo(HaveOccurred())
				Expect(value.(starlark.String).GoString()).To(ContainSubstring("BEGIN CERTIFICATE"))
				Expect(ca.AttrNames()).To(ContainElement("certificate"))

			})

			By("attribute private_key", func() {
				value, err := ca.Attr("private_key")
				Expect(err).NotTo(HaveOccurred())
				Expect(value.(starlark.String).GoString()).To(ContainSubstring("BEGIN RSA PRIVATE KEY"))
				Expect(ca.AttrNames()).To(ContainElement("private_key"))

			})

		})

		It("reads values from k8s", func() {
			k8s := k8s.NewK8sInMemoryEmpty()
			err := ca.read(&vaultK8s{k8s: k8s})
			Expect(err).NotTo(HaveOccurred())

		})
		Context("signs certificates", func() {
			It("works", func() {
				domains := starlark.NewList([]starlark.Value{starlark.String("example.com")})
				scertificate, err := makeCertificate(nil, nil, starlark.Tuple{starlark.String("name")},
					[]starlark.Tuple{
						{starlark.String("signer"), ca},
						{starlark.String("domains"), domains},
					})
				Expect(err).NotTo(HaveOccurred())
				certificate := scertificate.(*jewel)
				lca, err := certificate.Attr("ca")
				Expect(err).NotTo(HaveOccurred())
				gca, err := ca.Attr("certificate")
				Expect(err).NotTo(HaveOccurred())
				Expect(lca).To(Equal(gca))
			})
		})
	})

})
