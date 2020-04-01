package shalm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Converter", func() {

	It("configType convertion works", func() {
		var t configType
		var err error
		err = t.set("")
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(BeEquivalentTo(configTypeString))
		err = t.set("string")
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(BeEquivalentTo(configTypeString))
		err = t.set("bool")
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(BeEquivalentTo(configTypeBool))
		err = t.set("password")
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(BeEquivalentTo(configTypePassword))
		err = t.set("selection")
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(BeEquivalentTo(configTypeSelection))
		err = t.set("invalid")
		Expect(err).To(HaveOccurred())

	})

})
