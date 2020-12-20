package k8s

import (
	"bytes"
	"fmt"
	"io"

	"github.com/k14s/starlark-go/starlark"
	"github.com/pkg/errors"
	"github.com/wonderix/shalm/pkg/shalm/renderer"
	"github.com/wonderix/shalm/pkg/starutils"
)

// Stream -
type Stream = func(io.Writer) error

type streamValue struct {
	Stream
}

var _ starlark.Value = (*streamValue)(nil)
var _ starutils.GoConvertible = (*streamValue)(nil)

// ErrorStream -
func ErrorStream(err error) Stream {
	return func(writer io.Writer) error {
		return err
	}
}

// ObjectErrorStream -
func ObjectErrorStream(err error) ObjectStream {
	return func(writer ObjectConsumer) error {
		return err
	}
}

// NewStreamValue -
func NewStreamValue(s Stream) starlark.Value {
	return &streamValue{Stream: s}
}

// ToStream -
func ToStream(v starlark.Value, err error) Stream {
	if err != nil {
		return ErrorStream(err)
	}
	switch v := v.(type) {
	case *streamValue:
		return v.Stream
	case starlark.String:
		return func(writer io.Writer) error {
			_, err := writer.Write([]byte(v.GoString()))
			return err
		}
	}
	return ErrorStream(errors.New("Invalid return code from template"))
}

// YamlConcat -
func YamlConcat(streams ...Stream) Stream {
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
func (c *streamValue) String() string {
	buf := &bytes.Buffer{}
	err := c.Stream(buf)
	if err != nil {
		return err.Error()
	}
	return buf.String()
}

// Type -
func (c *streamValue) Type() string { return "stream" }

// Freeze -
func (c *streamValue) Freeze() {}

// Truth -
func (c *streamValue) Truth() starlark.Bool { return true }

// Hash -
func (c *streamValue) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (c *streamValue) Attr(name string) (starlark.Value, error) {
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("stream has no .%s attribute", name))
}

// AttrNames -
func (c *streamValue) AttrNames() []string {
	return []string{}
}

// starutils.ToGo -

func (c *streamValue) ToGo() interface{} {
	return nil
}
