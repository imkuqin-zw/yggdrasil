package types

type InstanceInfo interface {
	// 命名空间
	Namespace() string
	// 服务名
	Name() string
	// 版本号
	Version() string
	// 地区
	Region() string
	// 地域
	Zone() string
	// 园区
	Campus() string
	// 元信息
	Metadata() map[string]string
	// 添加元信息（忽略已存在的KEY）
	AddMetadata(key, val string) bool
}
