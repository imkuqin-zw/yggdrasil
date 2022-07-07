package config

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xmap"
	"github.com/mitchellh/mapstructure"
)

var regx, _ = regexp.Compile(`{([\w.-]+)}`)

type values struct {
	keyDelimiter string
	val          map[string]interface{}
}

func newValues(keyDelimiter string, val map[string]interface{}) *values {
	if val == nil {
		val = map[string]interface{}{}
	}
	return &values{keyDelimiter: keyDelimiter, val: val}
}

func (vs *values) get(key string) interface{} {
	dd, ok := vs.val[key]
	if ok {
		return dd
	}
	return xmap.DeepSearchInMap(vs.val, vs.genPath(key)...)
}

func (vs *values) genPath(key string) []string {
	matches := make([]string, 0)
	key = regx.ReplaceAllStringFunc(key, func(s string) string {
		matches = append(matches, s[1:len(s)-1])
		return "{}"
	})
	paths := strings.Split(key, vs.keyDelimiter)
	j := 0
	for i, item := range paths {
		if item == "{}" {
			paths[i] = matches[j]
			j++
		}
	}
	return paths
}

func (vs *values) Get(key string) types.ConfigValue {
	if key == "" {
		return &value{val: vs.val}
	}
	return newValue(vs.get(key))
}

func (vs *values) Del(key string) error {
	paths := strings.Split(key, vs.keyDelimiter)
	tmp := vs.val
	var ok bool
	for _, path := range paths[:len(paths)-1] {
		tmp, ok = tmp[path].(map[string]interface{})
		if !ok {
			return nil
		}
	}
	delete(tmp, key)
	return nil
}

func (vs *values) Set(key string, val interface{}) error {
	paths := strings.Split(key, vs.keyDelimiter)
	tmp := vs.val
	var ok bool
	for _, path := range paths[:len(paths)-1] {
		tmp, ok = tmp[path].(map[string]interface{})
		if !ok {
			return nil
		}
	}
	tmp[key] = val
	return nil
}

func (vs *values) Map() map[string]interface{} {
	return vs.val
}

func (vs *values) Scan(v interface{}) error {
	config := mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     v,
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return err
	}
	return decoder.Decode(vs.val)
}

func (vs *values) Bytes() []byte {
	if vs.val != nil {
		data, _ := json.Marshal(vs.val)
		return data
	}
	return []byte{}
}

func (vs *values) deepSearchInMap(val map[string]interface{}, key, delimiter string) interface{} {
	if v, ok := val[key]; ok {

		return v
	}
	keys := strings.SplitN(key, delimiter, 2)
	tmp, ok := val[keys[0]].(map[string]interface{})
	if !ok {
		return nil
	}
	return vs.deepSearchInMap(tmp, keys[1], delimiter)
}
