package pkg

import (
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
)

var (
	instanceInfo = &InstanceInfo{}
	once         sync.Once
)

func InitInstanceInfo() {
	once.Do(func() {
		instanceInfo = &InstanceInfo{
			name:      config.GetString("yggdrasil.application.name", ""),
			region:    config.GetString("yggdrasil.application.region", ""),
			zone:      config.GetString("yggdrasil.application.zone", ""),
			campus:    config.GetString("yggdrasil.application.campus", ""),
			namespace: config.GetString("yggdrasil.application.namespace", "default"),
			version:   config.GetString("yggdrasil.application.version", "v1"),
			metadata:  config.GetStringMap("yggdrasil.application.metadata", map[string]string{}),
		}
	})
}

// 命名空间
func Namespace() string {
	return instanceInfo.Namespace()
}

// 服务名
func Name() string {
	return instanceInfo.Name()
}

// 版本号
func Version() string {
	return instanceInfo.Version()
}

// 地区
func Region() string {
	return instanceInfo.Region()
}

// 地域
func Zone() string {
	return instanceInfo.Zone()
}

// 园区
func Campus() string {
	return instanceInfo.Campus()
}

// 元数据
func Metadata() map[string]string {
	return instanceInfo.Metadata()
}

// 元数据
func AddMetadata(key, val string) bool {
	return instanceInfo.AddMetadata(key, val)
}

type InstanceInfo struct {
	namespace string
	name      string
	version   string
	region    string
	zone      string
	campus    string
	mu        sync.RWMutex
	metadata  map[string]string
}

var _ types.InstanceInfo = (*InstanceInfo)(nil)

func (i *InstanceInfo) Namespace() string {
	return i.namespace
}

func (i *InstanceInfo) Name() string {
	return i.name
}

func (i *InstanceInfo) Version() string {
	return i.version
}

func (i *InstanceInfo) Region() string {
	return i.region
}

func (i *InstanceInfo) Zone() string {
	return i.zone
}

func (i *InstanceInfo) Campus() string {
	return i.campus
}

func (i *InstanceInfo) Metadata() map[string]string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	md := make(map[string]string)
	for k, v := range i.metadata {
		md[k] = v
	}
	return md
}

func (i *InstanceInfo) AddMetadata(key, val string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.metadata[key]; ok {
		return false
	}
	i.metadata[key] = val
	return true
}
