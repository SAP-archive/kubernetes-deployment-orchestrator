package shalm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"go.starlark.net/starlark"

	"github.com/blang/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/wonderix/shalm/pkg/shalm/test"
)

var _ = Describe("Chart", func() {

	Context("initialize", func() {

		It("reads Chart.yaml", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := NewTestDir()
			defer dir.Remove()
			repo, _ := NewRepo()
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			c, err := newChart(thread, repo, dir.Root())
			Expect(err).NotTo(HaveOccurred())
			Expect(c.GetName()).To(Equal("mariadb"))
		})
		It("reads Chart.yaml 'v' prefix in version", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := NewTestDir()
			defer dir.Remove()
			repo, _ := NewRepo()
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: v6.12.2\n"), 0644)
			c, err := newChart(thread, repo, dir.Root())
			Expect(err).NotTo(HaveOccurred())
			Expect(c.GetName()).To(Equal("mariadb"))
			Expect(c.GetVersion()).To(Equal(semver.Version{Major: 6, Minor: 12, Patch: 2}))
		})

		It("reads values.yaml", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := NewTestDir()
			defer dir.Remove()
			repo, _ := NewRepo()
			dir.WriteFile("values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
			c, err := newChart(thread, repo, dir.Root())
			Expect(err).NotTo(HaveOccurred())
			attr, err := c.Attr("replicas")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.(starlark.String).GoString()).To(Equal("1"))
			attr, err = c.Attr("timeout")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.(starlark.String).GoString()).To(Equal("30s"))
		})

	})
	Context("Chart.start", func() {
		var dir TestDir
		var c ChartValue
		thread := &starlark.Thread{Name: "main"}
		var k8s *FakeK8s
		BeforeEach(func() {
			k8s = &FakeK8s{}
			k8s.ForConfigStub = func(context string) (K8s, error) {
				return k8s, nil
			}
			k8s.ForSubChartStub = func(namespace, app string, version semver.Version) K8s {
				return k8s
			}
			dir = NewTestDir()
			repo, _ := NewRepo()
			dir.WriteFile("values.yaml", []byte("timeout: \"30s\"\n"), 0644)
			dir.WriteFile("values_2.yaml", []byte("backup: true\n"), 0644)
			dir.WriteFile("Chart.star", []byte(`
def init(self):
	self.timeout = "60s"
	self.load_yaml("values_2.yaml")
def method(self):
	return self.namespace
def apply(self):
	self.k8s.for_config('test')
	return self.__apply()
def template(self):
	return '{ "Kind" : "hello" }'
def delete(self):
	return self.__delete()
`),
				0644)
			var err error
			c, err = newChart(thread, repo, dir.Root(), WithK8s(k8s))
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			dir.Remove()
		})
		It("evalutes constructor", func() {
			{
				attr, err := c.Attr("timeout")
				Expect(err).NotTo(HaveOccurred())
				Expect(attr.(starlark.String).GoString()).To(Equal("60s"))
			}
			{
				attr, err := c.Attr("backup")
				Expect(err).NotTo(HaveOccurred())
				Expect(bool(attr.(starlark.Bool))).To(Equal(true))
			}
		})
		It("binds method to self", func() {
			attr, err := c.Attr("method")
			Expect(err).NotTo(HaveOccurred())
			value, err := starlark.Call(thread, attr.(starlark.Callable), nil, nil)
			Expect(value.(starlark.String).GoString()).To(Equal("default"))
		})
		It("overrides apply", func() {
			attr, err := c.Attr("apply")
			Expect(err).NotTo(HaveOccurred())
			_, err = starlark.Call(thread, attr.(starlark.Callable), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8s.ApplyCallCount()).To(Equal(1))
		})
		It("overrides delete", func() {
			attr, err := c.Attr("delete")
			Expect(err).NotTo(HaveOccurred())
			_, err = starlark.Call(thread, attr.(starlark.Callable), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8s.DeleteCallCount()).To(Equal(1))
		})
		It("overrides template", func() {
			attr, err := c.Attr("template")
			Expect(err).NotTo(HaveOccurred())
			s := toStream(starlark.Call(thread, attr.(starlark.Callable), nil, nil))
			buf := &bytes.Buffer{}
			err = s(buf)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println(buf.String())
			Expect(buf.String()).To(Equal("{ \"Kind\" : \"hello\" }"))
		})
	})
	Context("Render embedded ytt template", func() {
		var dir TestDir
		var c ChartValue
		thread := &starlark.Thread{Name: "main"}
		writer := bytes.Buffer{}
		k := &FakeK8s{
			ApplyStub: func(i ObjectStream, options *K8sOptions) error {
				return i.Encode()(&writer)
			},
		}
		k.ForSubChartStub = func(s string, app string, version semver.Version) K8s {
			return k
		}
		BeforeEach(func() {
			dir = NewTestDir()
			repo, _ := NewRepo()
			dir.MkdirAll("ytt", 0755)
			dir.WriteFile("Chart.star", []byte(`
def init(self):
	self.timeout = "60s"
def template(self,glob=''):
	return self.eytt("ytt")
`),
				0644)
			dir.WriteFile("ytt/test.yml", []byte("#@ if True:\ntest: #@ self.timeout\n#@ end\n"), 0644)
			var err error
			c, err = newChart(thread, repo, dir.Root(), WithSkipChart(true), WithK8s(k))
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			dir.Remove()
		})
		It("applies embedded ytt", func() {
			err := c.Apply(thread)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.ApplyCallCount()).To(Equal(1))
			Expect(writer.String()).To(Equal("\n---\n{\"test\":\"60s\"}\n"))
		})
	})
	Context("Render ytt template", func() {
		var dir TestDir
		var c ChartValue
		thread := &starlark.Thread{Name: "main"}
		writer := bytes.Buffer{}
		k := &FakeK8s{
			ApplyStub: func(i ObjectStream, options *K8sOptions) error {
				return i.Encode()(&writer)
			},
		}
		k.ForSubChartStub = func(s string, app string, version semver.Version) K8s {
			return k
		}
		BeforeEach(func() {
			dir = NewTestDir()
			repo, _ := NewRepo()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("Chart.star", []byte(`
def init(self):
	self.timeout = "60s"
def template(self,glob=''):
	return self.ytt(self.helm("templates"))
`),
				0644)
			dir.WriteFile("templates/test.yaml", []byte("#@ timeout = {{ .Values.timeout | quote }}\ntest: #@ timeout\n"), 0644)
			var err error
			c, err = newChart(thread, repo, dir.Root(), WithSkipChart(true), WithK8s(k))
			Expect(err).NotTo(HaveOccurred())

		})
		AfterEach(func() {
			dir.Remove()
		})
		It("applies ytt", func() {
			_, err := os.Stat("/usr/local/bin/ytt")
			if err != nil {
				Skip("ytt is not installed")
			}
			err = c.Apply(thread)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.ApplyCallCount()).To(Equal(1))
			Expect(writer.String()).To(Equal("\n---\n{\"test\":\"60s\"}\n"))
		})
	})
	Context("Render template in ytt-templates", func() {
		var dir TestDir
		var c ChartValue
		thread := &starlark.Thread{Name: "main"}
		writer := bytes.Buffer{}
		k := &FakeK8s{
			ApplyStub: func(i ObjectStream, options *K8sOptions) error {
				return i.Encode()(&writer)
			},
		}
		k.ForSubChartStub = func(s string, app string, version semver.Version) K8s {
			return k
		}
		BeforeEach(func() {
			dir = NewTestDir()
			repo, _ := NewRepo()
			dir.MkdirAll("ytt-templates", 0755)
			dir.WriteFile("Chart.star", []byte(`
def init(self):
	self.timeout = "60s"
`),
				0644)
			dir.WriteFile("ytt-templates/test.yaml", []byte("test: #@ self.timeout\n"), 0644)
			var err error
			c, err = newChart(thread, repo, dir.Root(), WithSkipChart(true), WithK8s(k))
			Expect(err).NotTo(HaveOccurred())

		})
		AfterEach(func() {
			dir.Remove()
		})
		It("applies ytt", func() {
			err := c.Apply(thread)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.ApplyCallCount()).To(Equal(1))
			Expect(writer.String()).To(Equal("\n---\n{\"test\":\"60s\"}\n"))
		})
	})
	Context("methods", func() {
		var dir TestDir
		var c ChartValue
		thread := &starlark.Thread{Name: "main"}
		var writer *bytes.Buffer
		BeforeEach(func() {
			dir = NewTestDir()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			repo, _ := NewRepo()
			var err error
			writer = &bytes.Buffer{}
			k := &FakeK8s{
				ApplyStub: func(i ObjectStream, options *K8sOptions) error {
					return i.Encode()(writer)
				},
				DeleteStub: func(i ObjectStream, options *K8sOptions) error {
					i.Encode()(writer)
					return nil
				},
			}
			k.ForSubChartStub = func(s string, app string, version semver.Version) K8s {
				return k
			}
			c, err = newChart(thread, repo, dir.Root(), WithNamespace("namespace"), WithSkipChart(true), WithK8s(k))
			Expect(err).NotTo(HaveOccurred())
			Expect(c.GetName()).To(Equal("mariadb"))

		})
		AfterEach(func() {
			dir.Remove()
		})
		It("templates a chart", func() {
			defer dir.Remove()
			Expect(c.GetName()).To(Equal("mariadb"))
			buf := &bytes.Buffer{}
			err := c.Template(thread)(buf)
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("---\nnamespace: namespace\n"))
		})

		It("applies a chart", func() {
			Expect(c.GetName()).To(Equal("mariadb"))
			err := c.Apply(thread)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.String()).To(Equal("\n---\n{\"namespace\":\"namespace\"}\n"))
		})

		It("deletes a chart", func() {
			Expect(c.GetName()).To(Equal("mariadb"))
			err := c.Delete(thread)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.String()).To(Equal("\n---\n{\"namespace\":\"namespace\"}\n"))
		})

		It("packages a chart", func() {
			writer := &bytes.Buffer{}
			err := c.Package(writer, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(bytes.HasPrefix(writer.Bytes(), []byte{0x1F, 0x8B, 0x08})).To(BeTrue())
		})

		It("packages a helm chart", func() {
			writer := &bytes.Buffer{}
			err := c.Package(writer, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(bytes.HasPrefix(writer.Bytes(), []byte{0x1F, 0x8B, 0x08})).To(BeTrue())
		})

		It("applies subcharts", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := NewTestDir()
			defer dir.Remove()
			repo, _ := NewRepo()
			dir.MkdirAll("chart1/templates", 0755)
			dir.MkdirAll("chart2/templates", 0755)
			dir.WriteFile("chart1/Chart.star", []byte("def init(self):\n  self.chart2 = chart(\"../chart2\",namespace=\"chart2\")\n"), 0644)

			dir.WriteFile("chart2/templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			dir.WriteFile("chart2/Chart.yaml", []byte("name: test\nversion: 1.0.0\n"), 0644)
			writer := bytes.Buffer{}
			k := &FakeK8s{
				DeleteStub: func(i ObjectStream, options *K8sOptions) error {
					i.Encode()(&writer)
					return nil
				},
			}
			k.ForSubChartStub = func(s string, app string, version semver.Version) K8s {
				return k
			}
			c, err := newChart(thread, repo, dir.Join("chart1"), WithSkipChart(true), WithK8s(k))
			Expect(err).NotTo(HaveOccurred())
			err = c.Delete(thread)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.DeleteCallCount()).To(Equal(2))
			Expect(writer.String()).To(Equal("\n---\n{\"namespace\":\"chart2\"}\n"))
		})

	})
	It("behaves like starlark value", func() {
		thread := &starlark.Thread{Name: "main"}
		dir := NewTestDir()
		defer dir.Remove()
		repo, _ := NewRepo()
		dir.WriteFile("values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
		c, err := newChart(thread, repo, dir.Root())
		Expect(err).NotTo(HaveOccurred())
		Expect(c.String()).To(ContainSubstring("replicas = \"1\""))
		Expect(c.Hash()).NotTo(Equal(uint32(0)))
		Expect(c.Truth()).To(BeEquivalentTo(true))
		Expect(c.Type()).To(Equal("chart"))
		value, err := c.Attr("name")
		Expect(err).NotTo(HaveOccurred())
		Expect(value.(starlark.String).GoString()).To(ContainSubstring("shalm"))
		value, err = c.Attr("namespace")
		Expect(err).NotTo(HaveOccurred())
		Expect(value.(starlark.String).GoString()).To(Equal("default"))
		value, err = c.Attr("apply")
		Expect(err).NotTo(HaveOccurred())
		Expect(value.(starlark.Callable).Name()).To(ContainSubstring("apply at"))
		value, err = c.Attr("unknown")
		Expect(err).To(HaveOccurred())
	})

	It("applies a credentials ", func() {
		thread := &starlark.Thread{Name: "main"}
		dir := NewTestDir()
		defer dir.Remove()
		repo, _ := NewRepo()
		dir.WriteFile("Chart.star", []byte("def init(self):\n  self.cred = user_credential(\"test\")\n"), 0644)
		var obj Object
		k := &FakeK8s{
			ApplyStub: func(i ObjectStream, options *K8sOptions) error {
				return i(func(o *Object) error { obj = *o; return nil })
			},
			GetStub: func(kind string, name string, k8s *K8sOptions) (*Object, error) {
				return nil, errors.New("NotFound")
			},
			IsNotExistStub: func(err error) bool {
				return true
			},
		}
		k.ForSubChartStub = func(s string, app string, version semver.Version) K8s {
			return k
		}
		c, err := newChart(thread, repo, dir.Root(), WithSkipChart(true), WithK8s(k))
		Expect(err).NotTo(HaveOccurred())
		err = c.Apply(thread)
		Expect(err).NotTo(HaveOccurred())
		Expect(obj.APIVersion).To(Equal("v1"))
		Expect(obj.Kind).To(Equal("Secret"))
		Expect(obj.MetaData.Name).To(Equal("test"))
		var user struct {
			Username []byte `json:"username"`
			Password []byte `json:"password"`
		}
		json.Unmarshal(obj.Additional["data"], &user)
		Expect(user.Username).To(HaveLen(24))
		Expect(user.Password).To(HaveLen(24))
	})

	It("merges values ", func() {
		thread := &starlark.Thread{Name: "main"}
		dir := NewTestDir()
		defer dir.Remove()
		repo, _ := NewRepo()
		dir.WriteFile("Chart.star", []byte("def init(self):\n  self.timeout=50\n"), 0644)
		c, err := newChart(thread, repo, dir.Root())
		Expect(err).NotTo(HaveOccurred())
		c.mergeValues(map[string]interface{}{"timeout": 60, "string": "test"})
		Expect(c.values["timeout"]).To(Equal(starlark.MakeInt(60)))
		Expect(c.values["string"]).To(Equal(starlark.String("test")))
	})

	Context("Exchange templates between charts", func() {
		var dir TestDir
		var c ChartValue
		thread := &starlark.Thread{Name: "main"}
		BeforeEach(func() {
			dir = NewTestDir()
			repo, _ := NewRepo()
			dir.MkdirAll("chart1/ytt", 0755)
			dir.WriteFile("chart1/ytt/values.yaml", []byte("#@ load(\"@ytt:data\", \"data\")\nconfig: #@ data.values.sys_domain\n"), 0644)
			dir.WriteFile("chart1/Chart.star", []byte(`
def init(self,config=None):
	self.config = config
def template(self,glob=""):
   return self.ytt("ytt",self.config)
`), 0644)
			dir.MkdirAll("chart2/config", 0755)
			dir.WriteFile("chart2/config/values.yaml", []byte("#@data/values\n---\nsys_domain: {{ .Values.domain }}\n"), 0644)
			dir.WriteFile("chart2/Chart.star", []byte(`
def init(self):
    self.domain = "example.com"
    self.chart1 = chart("../chart1",config=self.helm("config"))
`), 0644)
			var err error
			c, err = newChart(thread, repo, dir.Join("chart2"))
			Expect(err).NotTo(HaveOccurred())
		})
		It("templates correct", func() {
			_, err := os.Stat("/usr/local/bin/ytt")
			if err != nil {
				Skip("ytt is not installed")
			}
			s := c.Template(thread)
			out := &bytes.Buffer{}
			err = s(out)
			Expect(err).NotTo(HaveOccurred())
			Expect(out.String()).To(ContainSubstring("config: example.com"))
		})
		AfterEach(func() {
			dir.Remove()
		})

	})
})
