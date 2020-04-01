package shalm

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	. "github.com/wonderix/shalm/pkg/shalm/test"
)

var _ = Describe("Dir Cache", func() {

	It("dir cache works", func() {
		dir := NewTestDir()
		defer dir.Remove()
		cache := NewDirCache(dir.Root())
		tag := "tag1"
		load := func(name string, dir func() (string, error), etag string) (string, error) {
			if etag != tag {
				target, err := dir()
				if err != nil {
					return "", err
				}
				err = ioutil.WriteFile(path.Join(target, name), []byte(tag), 0644)
				return tag, err
			}
			return tag, nil
		}
		opener := cache.WrapDir(load)
		By("Caches initial artifact", func() {
			cached, err := opener("hello.txt")
			Expect(err).NotTo(HaveOccurred())
			_, err = os.Stat(cached)
			Expect(err).NotTo(HaveOccurred())
			content, err := ioutil.ReadFile(path.Join(cached, "hello.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(tag))
		})

		By("Doesn't touch unchanged files", func() {
			cached1, err := opener("hello.txt")
			Expect(err).NotTo(HaveOccurred())
			stat1, err := os.Stat(cached1)
			Expect(err).NotTo(HaveOccurred())

			cached2, err := opener("hello.txt")
			Expect(err).NotTo(HaveOccurred())
			stat2, err := os.Stat(cached2)
			Expect(err).NotTo(HaveOccurred())
			Expect(cached1).To(Equal(cached2))
			Expect(stat1.ModTime()).To(Equal(stat2.ModTime()))
		})

		By("Refreshes files if tag changes", func() {
			tag = "tag2"
			cached, err := opener("hello.txt")
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadFile(path.Join(cached, "hello.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(tag))
		})

		By("Returns error if load fails", func() {
			opener := cache.WrapDir(func(name string, dir func() (string, error), etag string) (string, error) {
				return "", errors.New("error")
			})
			_, err := opener("hello.txt")
			Expect(err).To(HaveOccurred())
		})

		By("Returns error if load doesn't call dir", func() {
			opener := cache.WrapDir(func(name string, dir func() (string, error), etag string) (string, error) {
				return "", nil
			})
			_, err := opener("hello.txt")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("No content written for hello.txt"))
		})

	})

	It("reader cache works", func() {
		dir := NewTestDir()
		defer dir.Remove()
		cache := NewDirCache(dir.Root())
		tag := "tag1"
		load := func(name string, etag string) (io.ReadCloser, string, error) {
			if etag != tag {
				return ioutil.NopCloser(bytes.NewBuffer([]byte(tag))), tag, nil
			}
			return nil, tag, nil
		}
		opener := cache.WrapReader(load)
		By("Caches reader", func() {
			reader, err := opener("hello.txt")
			Expect(err).NotTo(HaveOccurred())
			content := make([]byte, 32)
			c, err := reader.Read(content)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content[:c])).To(Equal(tag))
		})

	})

})
