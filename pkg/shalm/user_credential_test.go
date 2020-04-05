package shalm

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var _ = Describe("userCredentials", func() {

	Context("Read", func() {
		It("reads username and password from k8s", func() {
			var user struct {
				Username []byte `json:"username"`
				Password []byte `json:"password"`
			}
			user.Username = []byte("username1")
			user.Password = []byte("password1")
			k8s := FakeK8s{
				GetStub: func(kind string, name string, k8s *K8sOptions) (*Object, error) {
					data, _ := json.Marshal(user)
					return &Object{
						Additional: map[string]json.RawMessage{
							"data": json.RawMessage(data),
						},
					}, nil
				},
			}
			u, _ := makeUserCredential(nil, nil, starlark.Tuple{starlark.String("test")}, nil)
			userCred := u.(*vault)
			err := userCred.read(&k8s)
			Expect(err).NotTo(HaveOccurred())
			Expect(attValue(userCred, "username")).To(Equal(string(user.Username)))
			Expect(attValue(userCred, "password")).To(Equal(string(user.Password)))

		})

		It("creates new random username and password if user_credential doesn't exist", func() {
			k8s := FakeK8s{
				GetStub: func(kind string, name string, k8s *K8sOptions) (*Object, error) {
					return nil, errors.New("NotFound")
				},
				IsNotExistStub: func(err error) bool {
					return true
				},
			}
			u, _ := makeUserCredential(nil, nil, starlark.Tuple{starlark.String("test")}, nil)
			userCred := u.(*vault)
			err := userCred.read(&k8s)
			Expect(err).NotTo(HaveOccurred())
			Expect(attValue(userCred, "username")).To(HaveLen(24))
			Expect(attValue(userCred, "password")).To(HaveLen(24))
		})

	})

})

func attValue(v *vault, name string) string {
	val, e := v.Attr(name)
	if e != nil {
		panic(e)
	}
	return val.(starlark.String).GoString()
}
