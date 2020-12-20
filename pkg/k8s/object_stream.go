package k8s

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"

	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// ObjectWriter -
type ObjectWriter = func(obj *Object) error

// ObjectStream -
type ObjectStream func(w ObjectWriter) error

// CancelObjectStream -
type CancelObjectStream struct{}

func (c CancelObjectStream) Error() string {
	return "object stream canceled"
}

var _ error = (*CancelObjectStream)(nil)

// Map -
func (o ObjectStream) Map(f func(obj *Object) *Object) ObjectStream {
	return func(w ObjectWriter) error {
		return o(func(obj *Object) error {
			return w(f(obj))
		})
	}
}

// Sort -
func (o ObjectStream) Sort(f func(o1 *Object, o2 *Object) int, reverse bool) ObjectStream {
	return func(w ObjectWriter) error {
		var objs []*Object
		err := o(func(obj *Object) error {
			objs = append(objs, obj)
			return nil
		})
		if err != nil {
			return err
		}
		if reverse {
			sort.Slice(objs, func(i, j int) bool {
				return f(objs[i], objs[j]) >= 0
			})
		} else {
			sort.Slice(objs, func(i, j int) bool {
				return f(objs[i], objs[j]) < 0
			})
		}
		for _, obj := range objs {
			err = w(obj)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// Encode -
func (o ObjectStream) Encode() Stream {
	return func(w io.Writer) error {
		enc := json.NewEncoder(w)
		return o(func(obj *Object) error {
			w.Write([]byte("\n---\n"))
			return enc.Encode(obj)
		})
	}
}

// GroupBy -
func (o ObjectStream) GroupBy(group func(o *Object) string) func(key string) ObjectStream {
	var objs []*Object
	err := o(func(obj *Object) error {
		objs = append(objs, obj)
		return nil
	})
	return func(key string) ObjectStream {
		if err != nil {
			return func(w ObjectWriter) error {
				return err
			}
		}
		return func(w ObjectWriter) error {
			for _, obj := range objs {
				if group(obj) == key {
					if err := w(obj); err != nil {
						return err
					}
				}
			}
			return nil
		}
	}
}

// Filter -
func (o ObjectStream) Filter(filter func(obj *Object) bool) ObjectStream {
	return func(w ObjectWriter) error {
		return o(func(obj *Object) error {
			if filter(obj) {
				return w(obj)
			}
			return nil
		})
	}
}

// Decode -
func Decode(in Stream) ObjectStream {
	return func(w ObjectWriter) error {
		buffer := &bytes.Buffer{}
		err := in(buffer)
		if err != nil {
			return err
		}
		if buffer.Len() == 0 {
			return nil
		}
		dec := yaml.NewYAMLToJSONDecoder(buffer)
		for {
			var doc Object
			if err = dec.Decode(&doc); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
			if !(doc.Kind == "" && len(doc.Additional) == 0 && doc.MetaData.Name == "") {
				err := w(&doc)
				if err != nil {
					return err
				}
			}
		}
	}
}

func concat(streams ...ObjectStream) ObjectStream {
	return func(w ObjectWriter) error {
		for _, s := range streams {
			if err := s(w); err != nil {
				return err
			}
		}
		return nil
	}
}

var encoder = k8sjson.NewSerializerWithOptions(k8sjson.DefaultMetaFactory, nil, nil, k8sjson.SerializerOptions{})

func objectStreamOf(objs ...runtime.Object) ObjectStream {
	return func(w ObjectWriter) error {
		for _, obj := range objs {
			buffer := &bytes.Buffer{}
			if err := encoder.Encode(obj, buffer); err != nil {
				return err
			}
			o := Object{}
			if err := json.Unmarshal(buffer.Bytes(), &o); err != nil {
				return err
			}
			if err := w(&o); err != nil {
				return err
			}
		}
		return nil
	}
}
