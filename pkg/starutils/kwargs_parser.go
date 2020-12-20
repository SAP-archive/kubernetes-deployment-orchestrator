package starutils

import (
	"github.com/k14s/starlark-go/starlark"
)

// KwArgsParser -
type KwArgsParser struct {
	KwArgs []starlark.Tuple
	Args   map[string]func(starlark.Value)
}

// Arg-
func (k *KwArgsParser) Arg(name string, extractor func(starlark.Value)) {
	if k.Args == nil {
		k.Args = make(map[string]func(starlark.Value))
	}
	k.Args[name] = extractor
}

// Parse -
func (k *KwArgsParser) Parse() []starlark.Tuple {
	var result []starlark.Tuple
	for _, arg := range k.KwArgs {
		if arg.Len() == 2 {
			key, keyOK := arg.Index(0).(starlark.String)
			if keyOK {
				extractor, ok := k.Args[key.GoString()]
				if ok {
					extractor(arg.Index(1))
					continue
				}
			}
		}
		result = append(result, arg)
	}
	return result
}

// KwArgsToStringDict -
func KwArgsToStringDict(kwargs []starlark.Tuple) starlark.StringDict {
	result := starlark.StringDict{}
	for _, arg := range kwargs {
		if arg.Len() == 2 {
			key, keyOK := arg.Index(0).(starlark.String)
			if keyOK {
				result[key.GoString()] = arg.Index(1)
			}
		}
	}
	return result
}
