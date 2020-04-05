package shalm

import (
	"math/rand"
	"strings"
	"time"

	"go.starlark.net/starlark"
	corev1 "k8s.io/api/core/v1"
)

type userCredentialBackend struct {
	usernameKey string
	passwordKey string
	username    string
	password    string
}

var _ VaultBackend = (*userCredentialBackend)(nil)

func (u *userCredentialBackend) Name() string {
	return "user_credential"
}

func (u *userCredentialBackend) Keys() map[string]string {
	return map[string]string{
		"username": u.usernameKey,
		"password": u.passwordKey,
	}
}

func (u *userCredentialBackend) Apply(m map[string][]byte) (map[string][]byte, error) {
	if u.username != "" {
		m[u.usernameKey] = []byte(u.username)
	} else if m[u.usernameKey] == nil {
		m[u.usernameKey] = []byte(createRandomString(24))
	}
	if u.password != "" {
		m[u.passwordKey] = []byte(u.password)
	} else if m[u.passwordKey] == nil {
		m[u.passwordKey] = []byte(createRandomString(24))
	}
	return m, nil
}

func makeUserCredential(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
	u := &userCredentialBackend{
		usernameKey: corev1.BasicAuthUsernameKey,
		passwordKey: corev1.BasicAuthPasswordKey,
	}
	var name string
	if err := starlark.UnpackArgs("user_credential", args, kwargs, "name", &name, "username?", &u.username, "password?", &u.password, "username_key?", u.usernameKey, "password_key?", &u.passwordKey); err != nil {
		return nil, err
	}

	return NewVault(u, name)
}

func createRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
