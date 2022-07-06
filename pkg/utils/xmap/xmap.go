package xmap

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

func Marshal(obj interface{}) map[string]interface{} {
	ot := reflect.TypeOf(obj)
	ov := reflect.ValueOf(obj)

	var out = make(map[string]interface{}, ot.NumField())
	for i := 0; i < ot.NumField(); i++ {
		out[ot.Field(i).Name] = ov.Field(i).Interface()
	}
	return out
}

func Unmarshal(in interface{}, out interface{}) error {
	return errors.WithStack(mapstructure.Decode(in, out))
}

// MergeStringMap merge two map
func MergeStringMap(dest map[string]interface{}, src ...map[string]interface{}) {
	for _, item := range src {
		mergeStringMap(dest, item)
	}
}

func mergeStringMap(dest, src map[string]interface{}) {
	for sk, sv := range src {
		tv, ok := dest[sk]
		if !ok {
			// val不存在时，直接赋值
			dest[sk] = sv
			continue
		}

		svType := reflect.TypeOf(sv)
		tvType := reflect.TypeOf(tv)
		if svType != tvType {
			continue
		}

		switch ttv := tv.(type) {
		case map[interface{}]interface{}:
			tsv := sv.(map[interface{}]interface{})
			ssv := ToMapStringInterface(tsv)
			stv := ToMapStringInterface(ttv)
			mergeStringMap(stv, ssv)
			dest[sk] = stv
		case map[string]interface{}:
			mergeStringMap(ttv, sv.(map[string]interface{}))
			dest[sk] = ttv
		default:
			dest[sk] = sv
		}
	}
}

// ToMapStringInterface cast map[interface{}]interface{} to map[string]interface{}
func ToMapStringInterface(src map[interface{}]interface{}) map[string]interface{} {
	tgt := map[string]interface{}{}
	for k, v := range src {
		tgt[fmt.Sprintf("%v", k)] = v
	}
	return tgt
}

func CoverInterfaceToStringMap(src map[string]interface{}) {
	for k, v := range src {
		switch v := v.(type) {
		case map[interface{}]interface{}:
			src[k] = ToMapStringInterface(v)
			CoverInterfaceToStringMap(src[k].(map[string]interface{}))
		case map[string]interface{}:
			CoverInterfaceToStringMap(src[k].(map[string]interface{}))
		default:
		}
	}
}

// DeepSearchInMap deep search in map
func DeepSearchInMap(m map[string]interface{}, paths ...string) interface{} {
	tmp := make(map[string]interface{})
	for k, v := range m {
		tmp[k] = v
	}
	for i, k := range paths {
		v, ok := tmp[k]
		if !ok {
			return nil
		}
		tmp, ok = v.(map[string]interface{})
		if !ok {
			if i != len(paths)-1 {
				return nil
			} else {
				return v
			}
		}
	}
	return tmp
}

func CloneMap(src map[string]interface{}) (map[string]interface{}, error) {
	// https://gist.github.com/soroushjp/0ec92102641ddfc3ad5515ca76405f4d
	var buf bytes.Buffer
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(src)
	if err != nil {
		return nil, err
	}
	var copy map[string]interface{}
	err = dec.Decode(&copy)
	if err != nil {
		return nil, err
	}
	return copy, nil
}

func MergeKVMap(dest map[string]string, src ...map[string]string) {
	for _, item := range src {
		for k, v := range item {
			dest[k] = v
		}
	}
}
