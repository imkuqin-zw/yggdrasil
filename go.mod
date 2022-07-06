module github.com/imkuqin-zw/yggdrasil

go 1.13

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1
	github.com/polarismesh/polaris-go v1.1.0
	github.com/prometheus/client_golang v1.12.1
	github.com/smartystreets/assertions v0.0.0-20190116191733-b6c0e53d7304 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.2
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/exporters/jaeger v1.7.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	go.uber.org/atomic v1.7.0
	go.uber.org/multierr v1.6.0
	go.uber.org/zap v1.21.0
	google.golang.org/genproto v0.0.0-20201214200347-8c77b98c765d
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/polarismesh/polaris-go v1.1.0 => github.com/imkuqin-zw/yggdrasil-polaris v1.1.2
