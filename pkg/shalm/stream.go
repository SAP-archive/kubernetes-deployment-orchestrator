package shalm

import (
	"bytes"
	"fmt"
	"io"

	"github.com/k14s/starlark-go/starlark"
	"github.com/pkg/errors"
	"github.com/wonderix/shalm/pkg/shalm/renderer"
)

// Stream -
type Stream = func(io.Writer) error

type stream struct {
	Stream
}

var _ starlark.Value = (*stream)(nil)

// ErrorStream -
func ErrorStream(err error) Stream {
	return func(writer io.Writer) error {
		return err
	}
}

// ObjectErrorStream -
func ObjectErrorStream(err error) ObjectStream {
	return func(writer ObjectWriter) error {
		return err
	}
}

func toStream(v starlark.Value, err error) Stream {
	if err != nil {
		return ErrorStream(err)
	}
	switch v := v.(type) {
	case *stream:
		return v.Stream
	case starlark.String:
		return func(writer io.Writer) error {
			_, err := writer.Write([]byte(v.GoString()))
			return err
		}
	}
	return ErrorStream(errors.New("Invalid return code from template"))
}

func yamlConcat(streams ...Stream) Stream {
	return func(in io.Writer) error {

		for _, s := range streams {
			writer := &renderer.YamlWriter{Writer: in}
			err := s(writer)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

type writeCounter struct {
	counter int
	writer  io.Writer
}

func (w *writeCounter) Write(data []byte) (int, error) {
	w.counter++
	return w.writer.Write(data)
}

// String -
func (c *stream) String() string {
	buf := &bytes.Buffer{}
	err := c.Stream(buf)
	if err != nil {
		return err.Error()
	}
	return buf.String()
}

// Type -
func (c *stream) Type() string { return "stream" }

// Freeze -
func (c *stream) Freeze() {}

// Truth -
func (c *stream) Truth() starlark.Bool { return true }

// Hash -
func (c *stream) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (c *stream) Attr(name string) (starlark.Value, error) {
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("stream has no .%s attribute", name))
}

// AttrNames -
func (c *stream) AttrNames() []string {
	return []string{}
}
