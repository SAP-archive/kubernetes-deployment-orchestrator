package v1alpha2

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAPIV1alpha2(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "api v1alpha1 suite")
}
