package shalm

import (
	"errors"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stream", func() {

	Context("stream", func() {
		It("behaves like starlark value", func() {
			s := &stream{Stream: func(w io.Writer) error { _, err := w.Write([]byte("test")); return err }}
			Expect(s.String()).To(ContainSubstring("test"))
			Expect(s.Type()).To(Equal("stream"))
			Expect(func() { s.Hash() }).Should(Panic())
			Expect(s.Truth()).To(BeEquivalentTo(true))
			_, err := s.Attr("test")
			Expect(err).To(HaveOccurred())
			Expect(s.AttrNames()).To(BeEmpty())
		})
	})
	Context("ErrorStream", func() {
		It("fails", func() {
			s := ErrorStream(errors.New("test"))
			Expect(s(nil)).To(MatchError("test"))
		})
	})
	Context("ObjectErrorStream", func() {
		It("fails", func() {
			s := ObjectErrorStream(errors.New("test"))
			Expect(s(nil)).To(MatchError("test"))
		})
	})

})
