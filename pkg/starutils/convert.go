package starutils

import (
	"encoding/json"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
)

// GoConvertible -
type GoConvertible interface {
	ToGo() interface{}
}

// ToStarlark -
func ToStarlark(vi interface{}) starlark.Value {
	if vi == nil {
		return starlark.None
	}
	switch v := reflect.ValueOf(vi); v.Kind() {
	case reflect.String:
		return starlark.String(v.String())
	case reflect.Bool:
		return starlark.Bool(v.Bool())
	case reflect.Int:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Int16:
		return starlark.MakeInt64(v.Int())
	case reflect.Uint:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uint16:
		return starlark.MakeUint64(v.Uint())
	case reflect.Float32:
		return starlark.Float(v.Float())
	case reflect.Float64:
		return starlark.Float(v.Float())
	case reflect.Slice:
		if b, ok := vi.([]byte); ok {
			return starlark.String(string(b))
		}
		a := make([]starlark.Value, 0)
		for i := 0; i < v.Len(); i++ {
			a = append(a, ToStarlark(v.Index(i).Interface()))
		}
		return starlark.Tuple(a)
	case reflect.Ptr:
		return ToStarlark(v.Elem().Interface())
	case reflect.Map:
		d := starlark.NewDict(16)
		for _, key := range v.MapKeys() {
			strct := v.MapIndex(key)
			keyValue := ToStarlark(key.Interface())
			d.SetKey(keyValue, ToStarlark(strct.Interface()))
		}
		return d
	case reflect.Struct:
		ios, ok := vi.(intstr.IntOrString)
		if ok {
			switch ios.Type {
			case intstr.String:
				return starlark.String(ios.StrVal)
			case intstr.Int:
				return starlark.MakeInt(int(ios.IntVal))
			}
		} else {
			data, err := json.Marshal(vi)
			if err != nil {
				panic(err)
			}
			var m map[string]interface{}
			err = json.Unmarshal(data, &m)
			if err != nil {
				panic(err)
			}
			return ToStarlark(m)
		}
	}
	panic(fmt.Errorf("cannot convert %v to starlark", vi))
}

// ToGoMap -
func ToGoMap(v starlark.IterableMapping) map[string]interface{} {
	d := make(map[string]interface{})
	for _, t := range v.Items() {
		key, ok := t.Index(0).(starlark.String)
		if ok {
			value := ToGo(t.Index(1))
			if value != nil {
				d[key.GoString()] = value
			}
		}
	}
	return d
}

// ToGoStringList -
func ToGoStringList(v starlark.Value) []string {
	if v == nil {
		return nil
	}
	switch v := v.(type) {
	case starlark.Indexable: // Tuple, List
		a := make([]string, 0)
		for i := 0; i < starlark.Len(v); i++ {
			a = append(a, v.Index(i).(starlark.String).GoString())
		}
		return a
	default:
		panic(fmt.Errorf("cannot convert %s to go string list", v.Type()))
	}
}

// ToGo -
func ToGo(v starlark.Value) interface{} {
	if v == nil {
		return nil
	}
	switch v := v.(type) {
	case starlark.NoneType:
		return nil
	case starlark.Bool:
		return bool(v)
	case starlark.Int:
		i, _ := v.Int64()
		return i
	case starlark.Float:
		return float64(v)
	case starlark.String:
		return v.GoString()
	case starlark.Indexable: // Tuple, List
		a := make([]interface{}, 0)
		for i := 0; i < starlark.Len(v); i++ {
			a = append(a, ToGo(v.Index(i)))
		}
		return a
	case starlark.Callable:
		return nil
	case starlark.IterableMapping:
		return ToGoMap(v)
	case GoConvertible:
		return v.ToGo()
	case *starlarkstruct.Struct:
		d := starlark.StringDict{}
		v.ToStringDict(d)
		return StringDictToGo(d)
	default:
		panic(fmt.Errorf("cannot convert %s to go", v.Type()))
	}
}

// StringDictToGo -
func StringDictToGo(stringDict starlark.StringDict) map[string]interface{} {
	d := make(map[string]interface{})

	for k, v := range stringDict {
		value := ToGo(v)
		if value != nil {
			d[k] = value
		}
	}
	return d
}
