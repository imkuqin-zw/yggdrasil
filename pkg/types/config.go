package types

import (
	"io"
	"time"
)

type Config interface {
	ConfigValues
	Close() error
	LoadSource(...ConfigSource) error
	AddWatcher(string, func(ConfigWatchEvent)) error
	DelWatcher(string, func(ConfigWatchEvent)) error
}

type WatchEventType uint32

const (
	_ WatchEventType = iota
	WatchEventUpd
	WatchEventAdd
	WatchEventDel
)

type ConfigWatchEvent interface {
	Type() WatchEventType
	Value() ConfigValue
}

// Source is the source from which conf is loaded
type ConfigSource interface {
	Name() string
	Read() (ConfigSourceData, error)
	Changeable() bool
	Watch() (<-chan ConfigSourceData, error)
	io.Closer
}

type ConfigPriority uint8

const (
	ConfigPriorityFile ConfigPriority = iota
	ConfigPriorityEnv
	ConfigPriorityFlag
	ConfigPriorityCli
	ConfigPriorityRemote
	ConfigPriorityMemory
	ConfigPriorityMax
)

type ConfigSourceData interface {
	Priority() ConfigPriority
	Data() []byte
	Unmarshal(v interface{}) error
}

type ConfigValues interface {
	Get(key string) ConfigValue
	Set(key string, val interface{}) error
	Del(key string) error
	Map() map[string]interface{}
	Scan(v interface{}) error
	Bytes() []byte
}

type ConfigValue interface {
	Bool(def ...bool) bool
	Int(def ...int) int
	Int64(def ...int64) int64
	String(def ...string) string
	Float64(def ...float64) float64
	Duration(def ...time.Duration) time.Duration
	StringSlice(def ...[]string) []string
	StringMap(def ...map[string]string) map[string]string
	Map(def ...map[string]interface{}) map[string]interface{}
	Scan(val interface{}) error
	Bytes(def ...[]byte) []byte
}
