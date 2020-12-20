package starutils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Converter", func() {

	It("converts from go to starlark and back", func() {
		data := make(map[string]interface{})
		data["String"] = "test"
		data["ByteArray"] = []byte("test")
		data["Array"] = []string{"test1", "test2"}
		data["Bool"] = true
		data["Float"] = 0.1
		data["Map"] = map[string]string{"test1": "test1", "test2": "test2"}
		starlarkValue := ToStarlark(data)
		goValue := ToGo(starlarkValue).(map[string]interface{})
		Expect(goValue["String"]).To(Equal("test"))
		Expect(goValue["ByteArray"]).To(Equal("test"))
		Expect(goValue["Array"]).To(ConsistOf("test1", "test2"))
		Expect(goValue["Bool"]).To(BeEquivalentTo(true))
		Expect(goValue["Float"]).To(BeEquivalentTo(0.1))
		Expect(goValue["Map"]).To(HaveKeyWithValue("test1", "test1"))
		Expect(goValue["Map"]).To(HaveKeyWithValue("test2", "test2"))
	})

})
