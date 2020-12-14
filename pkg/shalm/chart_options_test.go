package shalm

import (
	"fmt"
	"os"

	"github.com/k14s/starlark-go/starlark"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	. "github.com/wonderix/shalm/pkg/shalm/test"
)

var _ = Describe("Chart Options", func() {
	Context("ProxyMode", func() {
		It("set works correct", func() {
			var p ProxyMode
			Expect(p.Set("off")).NotTo(HaveOccurred())
			Expect(p).To(BeEquivalentTo(ProxyModeOff))
			Expect(p.Set("local")).NotTo(HaveOccurred())
			Expect(p).To(BeEquivalentTo(ProxyModeLocal))
			Expect(p.Set("remote")).NotTo(HaveOccurred())
			Expect(p).To(BeEquivalentTo(ProxyModeRemote))
			Expect(p.Set("")).NotTo(HaveOccurred())
			Expect(p.Set("invalid")).To(HaveOccurred())
		})

		It("string returns correct values", func() {
			Expect(ProxyMode(ProxyModeOff).String()).To(Equal("off"))
			Expect(ProxyMode(ProxyModeRemote).String()).To(Equal("remote"))
			Expect(ProxyMode(ProxyModeLocal).String()).To(Equal("local"))
		})

	})
	Context("KwArgsVar", func() {
		It("produces the correct output", func() {
			kwargs := KwArgsVar{}
			err := kwargs.Set("a=b=c")
			Expect(err).NotTo(HaveOccurred())
			Expect(kwargs).To(HaveLen(1))
			Expect(kwargs[0]).To(HaveLen(2))
			Expect(kwargs[0][0]).To(Equal(starlark.String("a")))
			Expect(kwargs[0][1]).To(Equal(starlark.String("b=c")))
		})

	})
	Context("KwArgsYamlVar", func() {
		It("produces the correct output", func() {
			dir := NewTestDir()
			dir.WriteFile("test.yaml", []byte("a: b\nc: d\n"), 0644)
			defer dir.Remove()
			kwargs := KwArgsVar{}
			kwargsYaml := kwArgsYamlVar{kwargs: &kwargs}
			err := kwargsYaml.Set(fmt.Sprintf("a=%s", dir.Join("test.yaml")))
			Expect(err).NotTo(HaveOccurred())
			Expect(kwargs).To(HaveLen(1))
			Expect(kwargs[0]).To(HaveLen(2))
			Expect(kwargs[0][0]).To(Equal(starlark.String("a")))
			c := kwargs[0][1].(*starlark.Dict)
			Expect(c.Len()).To(Equal(2))
		})

	})
	Context("kwArgsFileVar", func() {
		It("produces the correct output", func() {
			dir := NewTestDir()
			dir.WriteFile("test.txt", []byte("hello"), 0644)
			defer dir.Remove()
			kwargs := KwArgsVar{}
			kwargsFile := kwArgsFileVar{kwargs: &kwargs}
			err := kwargsFile.Set(fmt.Sprintf("a=%s", dir.Join("test.txt")))
			Expect(err).NotTo(HaveOccurred())
			Expect(kwargs).To(HaveLen(1))
			Expect(kwargs[0]).To(HaveLen(2))
			Expect(kwargs[0][0]).To(Equal(starlark.String("a")))
			Expect(kwargs[0][1]).To(Equal(starlark.String("hello")))
		})

	})
	Context("kwArgsEnvVar", func() {
		It("produces the correct output", func() {
			kwargs := KwArgsVar{}
			kwargsEnv := kwArgsEnvVar{kwargs: &kwargs}
			os.Setenv("SHALM_TEST", "XXXX")
			err := kwargsEnv.Set("test=SHALM_TEST")
			Expect(err).NotTo(HaveOccurred())
			Expect(kwargs).To(HaveLen(1))
			Expect(kwargs[0]).To(HaveLen(2))
			Expect(kwargs[0][0]).To(Equal(starlark.String("test")))
			Expect(kwargs[0][1]).To(Equal(starlark.String("XXXX")))
		})

	})
	Context("valuesFile", func() {
		It("produces the correct output", func() {
			dir := NewTestDir()
			dir.WriteFile("test.yaml", []byte("test: test\ntimeout: timeout"), 0644)
			defer dir.Remove()
			values := valuesFile{}
			err := values.Set(dir.Join("test.yaml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(HaveLen(2))
			Expect(values).To(HaveKeyWithValue("test", "test"), HaveKeyWithValue("timeout", "timeout"))
		})

	})
	Context("ChartOptions", func() {
		It("args are correct", func() {
			args := ChartOptions{}
			flagsSet := pflag.FlagSet{}
			args.AddFlags(&flagsSet)
			Expect(flagsSet.FlagUsages()).To(ContainSubstring(`--set kwargs             Set values (key=val).`))
			Expect(flagsSet.FlagUsages()).To(ContainSubstring(`-p, --proxy proxy-mode       Install helm chart using a combination of CR and operator. Possible values off, local and remote (default off)`))
			Expect(flagsSet.FlagUsages()).To(ContainSubstring(`-n, --namespace string       namespace for installation (default "default")`))
			Expect(flagsSet.FlagUsages()).To(ContainSubstring(`-s, --suffix string          Suffix which is used to build the chart name`))
		})
	})
})
