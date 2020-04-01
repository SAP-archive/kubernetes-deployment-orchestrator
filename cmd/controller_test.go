package cmd

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller", func() {

	It("test if controller starts successfully", func() {
		if os.Getenv("KUBECONFIG") == "" {
			Skip("No KUBECONFIG set")
		}
		stopCh := make(chan struct{}, 1)
		stopCh <- struct{}{}
		err := controller(stopCh)
		Expect(err).ToNot(HaveOccurred())
	})
})
