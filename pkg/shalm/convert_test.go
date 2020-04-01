package shalm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
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
		starlarkValue := toStarlark(data)
		goValue := toGo(starlarkValue).(map[string]interface{})
		Expect(goValue["String"]).To(Equal("test"))
		Expect(goValue["ByteArray"]).To(Equal("test"))
		Expect(goValue["Array"]).To(ConsistOf("test1", "test2"))
		Expect(goValue["Bool"]).To(BeEquivalentTo(true))
		Expect(goValue["Float"]).To(BeEquivalentTo(0.1))
		Expect(goValue["Map"]).To(HaveKeyWithValue("test1", "test1"))
		Expect(goValue["Map"]).To(HaveKeyWithValue("test2", "test2"))
	})

	It("merges starlark simple values", func() {
		Expect(merge(nil, starlark.True)).To(Equal(starlark.True))
		Expect(merge(starlark.True, nil)).To(Equal(starlark.True))
		Expect(merge(starlark.False, starlark.True)).To(Equal(starlark.True))
		Expect(merge(starlark.None, starlark.True)).To(Equal(starlark.True))
		Expect(merge(starlark.False, starlark.None)).To(Equal(starlark.False))
		Expect(merge(starlark.MakeInt(1), starlark.MakeInt(2))).To(Equal(starlark.MakeInt(2)))
		Expect(merge(starlark.String("v"), starlark.String("o"))).To(Equal(starlark.String("o")))
	})

	It("merges starlark tuples", func() {
		Expect(merge(starlark.Float(1.0), starlark.Float(2.0))).To(Equal(starlark.Float(2.0)))
		Expect(merge(starlark.Tuple([]starlark.Value{starlark.String("v")}), starlark.Tuple([]starlark.Value{starlark.String("o")}))).To(Equal(starlark.Tuple([]starlark.Value{starlark.String("o")})))
		Expect(merge(starlark.Tuple([]starlark.Value{starlark.String("v1"), starlark.String("v1")}), starlark.Tuple([]starlark.Value{starlark.String("o1")}))).
			To(Equal(starlark.Tuple([]starlark.Value{starlark.String("o1"), starlark.String("v1")})))
		Expect(merge(starlark.Tuple([]starlark.Value{starlark.String("v1")}), starlark.Tuple([]starlark.Value{starlark.String("o1"), starlark.String("o2")}))).
			To(Equal(starlark.Tuple([]starlark.Value{starlark.String("o1"), starlark.String("o2")})))
	})

	It("merges starlark list", func() {
		Expect(merge(starlark.NewList([]starlark.Value{starlark.String("v")}), starlark.NewList([]starlark.Value{starlark.String("o")}))).To(Equal(starlark.NewList([]starlark.Value{starlark.String("o")})))
	})

	It("merges starlark maps", func() {
		v := starlark.NewDict(0)
		v.SetKey(starlark.String("k1"), starlark.String("v1"))
		o := starlark.NewDict(0)
		o.SetKey(starlark.String("k1"), starlark.String("o1"))
		Expect(merge(v, o)).To(Equal(o))
		o.SetKey(starlark.String("k2"), starlark.String("o2"))
		Expect(merge(v, o)).To(Equal(o))

		v.SetKey(starlark.String("k3"), starlark.String("v3"))
		element, found, err := merge(v, o).(starlark.IterableMapping).Get(starlark.String("k3"))
		Expect(found).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
		Expect(element).To(Equal(starlark.String("v3")))
	})

	It("merges string dicts", func() {
		v := starlark.StringDict{}
		v["k1"] = starlark.String("v1")
		o := starlark.NewDict(0)
		o.SetKey(starlark.String("k1"), starlark.String("o1"))
		Expect(mergeStringDict(v, o)).To(HaveKeyWithValue("k1", starlark.String("o1")))
		o.SetKey(starlark.String("k2"), starlark.String("o2"))
		Expect(mergeStringDict(v, o)).To(HaveKeyWithValue("k2", starlark.String("o2")))

		v["k3"] = starlark.String("v3")
		Expect(mergeStringDict(v, o)).To(HaveKeyWithValue("k3", starlark.String("v3")))
	})

})
