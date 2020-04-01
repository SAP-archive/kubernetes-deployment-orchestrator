package shalm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("kwargs parser", func() {

	It("converts from and to starlark", func() {
		kwArgs := map[string]interface{}{
			"int":    4,
			"string": "hello world",
			"bool":   false,
		}
		kwArgs2 := kwargsToGo(kwargsToStarlark(kwArgs))
		Expect(kwArgs2).To(HaveKeyWithValue("string", "hello world"))
		Expect(kwArgs2).To(HaveKeyWithValue("bool", false))
		Expect(kwArgs2).To(HaveKeyWithValue("int", int64(4)))
	})

})
