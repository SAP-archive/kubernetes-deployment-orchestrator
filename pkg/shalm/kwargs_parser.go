package shalm

import (
	"go.starlark.net/starlark"
)

type kwargsParser struct {
	kwargs []starlark.Tuple
	args   map[string]func(starlark.Value)
}

func (k *kwargsParser) Arg(name string, extractor func(starlark.Value)) {
	if k.args == nil {
		k.args = make(map[string]func(starlark.Value))
	}
	k.args[name] = extractor
}

func (k *kwargsParser) Parse() []starlark.Tuple {
	var result []starlark.Tuple
	for _, arg := range k.kwargs {
		if arg.Len() == 2 {
			key, keyOK := arg.Index(0).(starlark.String)
			if keyOK {
				extractor, ok := k.args[key.GoString()]
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

func kwargsToGo(kwargs []starlark.Tuple) map[string]interface{} {
	result := make(map[string]interface{})
	for _, arg := range kwargs {
		if arg.Len() == 2 {
			key, keyOK := arg.Index(0).(starlark.String)
			if keyOK {
				result[key.GoString()] = toGo(arg.Index(1))
			}
		}
	}
	return result
}

func kwargsToStarlark(kwargs map[string]interface{}) []starlark.Tuple {
	var result []starlark.Tuple
	for k, v := range kwargs {
		result = append(result, starlark.Tuple{starlark.String(k), toStarlark(v)})
	}
	return result
}

func kwargsToStringDict(kwargs []starlark.Tuple) starlark.StringDict {
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
