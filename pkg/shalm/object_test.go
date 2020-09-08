package shalm

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Object", func() {

	It("marshals", func() {
		obj := Object{
			Kind:       "Object",
			APIVersion: "v1",
			MetaData: MetaData{
				Name:        "name",
				Namespace:   "space",
				Labels:      map[string]string{"key": "value"},
				Annotations: map[string]string{"x": "y"},
				Additional: map[string]json.RawMessage{
					"finalizers": json.RawMessage([]byte(`["test"]`)),
				},
			},
			Additional: map[string]json.RawMessage{
				"spec": json.RawMessage([]byte(`{"x":"y"}`)),
			},
		}
		data, err := json.Marshal(obj)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(MatchJSON(`{
			"apiVersion": "v1",
			"kind": "Object",
			"metadata": {
			  "annotations": {
				"x": "y"
			  },
			  "finalizers": [
				"test"
			  ],
			  "labels": {
				"key": "value"
			  },
			  "name": "name",
			  "namespace": "space"
			},
			"spec": {
			  "x": "y"
			}
		  }`))
		var obj2 Object
		err = json.Unmarshal(data, &obj2)
		Expect(err).NotTo(HaveOccurred())
		Expect(obj2).To(Equal(obj))
	})

})
