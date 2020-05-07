package shalm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("k8s client", func() {

	It("can read kubernetes service", func() {
		config, err := configKube("")
		if err != nil {
			Skip("no connection to k8s")
		}
		client, err := newK8sClient(config)
		if err != nil {
			Skip("no connection to k8s")
		}
		namespace := "default"
		obj, err := client.Get().
			Namespace(&namespace).
			Resource("services").
			Name("kubernetes").
			Do().
			Get()

		Expect(err).NotTo(HaveOccurred())
		Expect(obj.Kind).To(Equal("Service"))
	})

	It("can read kubernetes deployment", func() {
		// NOTICE: deployment is not in the core k8s api group
		config, err := configKube("")
		if err != nil {
			Skip("no connection to k8s")
		}
		client, err := newK8sClient(config)
		if err != nil {
			Skip("no connection to k8s")
		}
		namespace := "kube-system"
		obj, err := client.Get().
			Namespace(&namespace).
			Resource("deployments").
			Name("coredns").
			Do().
			Get()

		Expect(err).NotTo(HaveOccurred())
		Expect(obj.Kind).To(Equal("Deployment"))
	})

})
