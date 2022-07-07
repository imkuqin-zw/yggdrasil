package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
	"github.com/mitchellh/mapstructure"
)

type value struct {
	val interface{}
}

func newValue(val interface{}) types.ConfigValue {
	return &value{val: val}
}

func (m *value) Bool(def ...bool) bool {
	b, ok := m.val.(bool)
	if ok {
		return b
	}

	str, ok := m.val.(string)
	if !ok {
		if len(def) == 0 {
			return false
		}
		return def[0]
	}

	b, err := strconv.ParseBool(str)
	if err != nil {
		if len(def) == 0 {
			return false
		}
		return def[0]
	}

	return b
}

func (m *value) Int(def ...int) int {
	i, ok := m.val.(int)
	if ok {
		return i
	}

	str, ok := m.val.(string)
	if !ok {
		if len(def) == 0 {
			return 0
		}
		return def[0]
	}

	i, err := strconv.Atoi(str)
	if err != nil {
		if len(def) == 0 {
			return 0
		}
		return def[0]
	}

	return i
}

func (m *value) Int64(def ...int64) int64 {
	i, ok := m.val.(int64)
	if ok {
		return i
	}

	str, ok := m.val.(string)
	if !ok {
		if len(def) == 0 {
			return 0
		}
		return def[0]
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		if len(def) == 0 {
			return 0
		}
		return def[0]
	}

	return i
}

func (m *value) String(def ...string) string {
	if str, ok := m.val.(string); ok {
		return str
	}
	if len(def) == 0 {
		return ""
	}
	return def[0]
}

func (m *value) Float64(def ...float64) float64 {
	f, ok := m.val.(float64)
	if ok {
		return f
	}

	str, ok := m.val.(string)
	if !ok {
		if len(def) == 0 {
			return 0
		}
		return def[0]
	}

	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		if len(def) == 0 {
			return 0
		}
		return def[0]
	}

	return f
}

func (m *value) Duration(def ...time.Duration) time.Duration {
	v, ok := m.val.(time.Duration)
	if ok {
		return v
	}

	switch v := m.val.(type) {
	case time.Duration:
		return v
	case string:
		value, err := time.ParseDuration(v)
		if err != nil {
			if len(def) == 0 {
				return 0
			}
			return def[0]
		}
		return value
	default:
		if len(def) == 0 {
			return 0
		}
		return def[0]
	}
}

func (m *value) StringSlice(def ...[]string) []string {
	fmt.Printf("%T\n", m.val)
	switch sl := m.val.(type) {
	case []string:
		return sl
	case []interface{}:
		tmp := make([]string, len(sl))
		for i, item := range sl {
			tmp[i] = fmt.Sprintf("%v", item)
		}
		return tmp
	default:

	}
	sl, ok := m.val.([]string)
	if ok {
		return sl
	}
	if len(def) == 0 {
		return nil
	}
	return def[0]
}

func (m *value) StringMap(def ...map[string]string) map[string]string {
	res, ok := m.val.(map[string]string)
	if ok {
		return res
	}
	if len(def) == 0 {
		return map[string]string{}
	}
	return def[0]
}

func (m *value) Map(def ...map[string]interface{}) map[string]interface{} {
	res, ok := m.val.(map[string]interface{})
	if ok {
		return res
	}
	if len(def) == 0 {
		return map[string]interface{}{}
	}
	return def[0]
}

func (m *value) Scan(val interface{}) error {
	v, ok := m.val.(map[string]interface{})
	if !ok {
		return nil
	}
	config := mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     val,
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return err
	}
	return decoder.Decode(v)
}

func (m *value) Bytes(def ...[]byte) []byte {
	switch v := m.val.(type) {
	case []byte:
		return v
	case string:
		return xstrings.Str2bytes(v)
	default:
		if m.val != nil {
			if data, _ := json.Marshal(m.val); len(data) > 0 {
				return data
			}
		}
		if len(def) == 0 {
			return nil
		}
		return def[0]
	}
}
