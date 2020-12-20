package k8s

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("k8s", func() {

	Context("ObjectStream", func() {
		It("filters", func() {
			stream := ObjectStream(func(w ObjectConsumer) error {
				w(&Object{Kind: "test"})
				return w(&Object{Kind: "xxx"})
			})
			test := stream.Filter(func(obj *Object) bool { return obj.Kind == "test" })
			xxx := stream.Filter(func(obj *Object) bool { return obj.Kind == "xxx" })
			{
				count := 0
				kind := ""
				test(func(obj *Object) error { count++; kind = obj.Kind; return nil })
				Expect(count).To(Equal(1))
				Expect(kind).To(Equal("test"))
			}
			{
				count := 0
				kind := ""
				xxx(func(obj *Object) error { count++; kind = obj.Kind; return nil })
				Expect(count).To(Equal(1))
				Expect(kind).To(Equal("xxx"))
			}
		})
		It("buffers", func() {
			stream := ObjectStream(func(w ObjectConsumer) error {
				w(&Object{Kind: "test"})
				return w(&Object{Kind: "xxx"})
			})
			grouped := stream.GroupBy(func(o *Object) string { return o.Kind })
			count := 0
			grouped("test")(func(obj *Object) error { count++; return nil })
			Expect(count).To(Equal(1))
			count = 0
			grouped("xxx")(func(obj *Object) error { count++; return nil })
			Expect(count).To(Equal(1))
			count = 0
			grouped("unkonw")(func(obj *Object) error { count++; return nil })
			Expect(count).To(Equal(0))
		})

	})
	It("Sorts in correct order", func() {
		ordinal := 0
		for _, kind := range []string{"namespace",
			"NetworkPolicy",
			"ResourceQuota",
			"LimitRange",
			"PodSecurityPolicy",
			"PodDisruptionBudget",
			"Secret",
			"ConfigMap",
			"StorageClass",
			"PersistentVolume",
			"PersistentVolumeClaim",
			"ServiceAccount",
			"CustomResourceDefinition",
			"ClusterRole",
			"ClusterRoleList",
			"ClusterRoleBinding",
			"ClusterRoleBindingList",
			"Role",
			"RoleList",
			"RoleBinding",
			"RoleBindingList",
			"Service",
			"DaemonSet",
			"Pod",
			"ReplicationController",
			"ReplicaSet",
			"Deployment",
			"HorizontalPodAutoscaler",
			"StatefulSet",
			"Job",
			"CronJob",
			"Ingress",
			"APIService"} {
			obj := Object{Kind: kind}
			ord := obj.kindOrdinal()
			Expect(ord).To(BeNumerically(">", ordinal))
			ordinal = ord
		}
	})

	It("doesn't set default namespace non namepspaced objects", func() {
		for _, kind := range []string{"namespace", "ResourceQuota", "CustomResourceDefinition", "ClusterRole",
			"ClusterRoleList", "ClusterRoleBinding", "ClusterRoleBindingList", "APIService"} {
			obj := Object{Kind: kind}
			obj.setDefaultNamespace("test")
			Expect(obj.MetaData.Namespace).To(Equal(""))
		}
	})

})
