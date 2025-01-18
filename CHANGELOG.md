<a name="unreleased"></a>
## [Unreleased]


<a name="contrib/polaris/v1.3.8"></a>
## [contrib/polaris/v1.3.8] - 2025-01-18

<a name="contrib/otelexporters/v1.3.8"></a>
## [contrib/otelexporters/v1.3.8] - 2025-01-18

<a name="v1.3.8"></a>
## [v1.3.8] - 2025-01-18
### Bug Fixes
- stats添加是否开启metrics配置
- 修改logger引用错误


<a name="contrib/gorm/v1.3.8"></a>
## [contrib/gorm/v1.3.8] - 2025-01-07
### Features
- **gorm:** gorm 添加 transaction


<a name="contrib/gorm/v1.3.7"></a>
## [contrib/gorm/v1.3.7] - 2024-09-23
### Features
- **gorm:** gorm 添加 sqlite driver


<a name="v1.3.7"></a>
## [v1.3.7] - 2024-09-02
### Bug Fixes
- 修复rest代码生成器未判断unaryInt为空的问题


<a name="v1.3.5"></a>
## [v1.3.5] - 2024-08-29
### Bug Fixes
- 修改未开启rest server时，注册rest handler 错误

### Features
- 添加管理内部服务功能


<a name="example/v1.3.4"></a>
## [example/v1.3.4] - 2024-08-22

<a name="v1.3.4"></a>
## [v1.3.4] - 2024-08-22
### Bug Fixes
- 修复genrest模板中查询参数解析的URL引用错误
- 修复rest代码生成器没有路径参数时import错误

### Features
- 增加处理原生http的能力


<a name="v1.3.3"></a>
## [v1.3.3] - 2024-07-28

<a name="contrib/polaris/v1.3.3"></a>
## [contrib/polaris/v1.3.3] - 2024-07-28

<a name="example/v1.3.3"></a>
## [example/v1.3.3] - 2024-07-28
### Bug Fixes
- 修复没有注入server时程序无法正常shutdown的问题
- 修复restful api server 正常退出时的错误的日志输出
- 修复stats配置为空是错误的日志输出

### Features
- 添加restful api


<a name="v1.3.1"></a>
## [v1.3.1] - 2024-05-02

<a name="contrib/otelexporters/v1.3.1"></a>
## [contrib/otelexporters/v1.3.1] - 2024-05-02

<a name="contrib/polaris/v1.3.1"></a>
## [contrib/polaris/v1.3.1] - 2024-05-02

<a name="contrib/gorm/v1.3.1"></a>
## [contrib/gorm/v1.3.1] - 2024-05-02

<a name="contrib/redis/v1.3.1"></a>
## [contrib/redis/v1.3.1] - 2024-05-02

<a name="example/v1.3.1"></a>
## [example/v1.3.1] - 2024-05-02
### Code Refactoring
- 重写logging引用
- 重构可观察性相关功能，添加stats中间扩展

### Features
- 添加monitor例子
- 使用workspace模式，contrib集成组件


<a name="v1.2.2"></a>
## [v1.2.2] - 2024-01-28
### Features
- 简单模式添加Trailer和header功能


<a name="v1.2.1"></a>
## [v1.2.1] - 2023-09-22
### Bug Fixes
- 修改status类型判断问题
- 添加公用方法并修改日志
- 修改status判断Reason错误，并优化代码

### Code Refactoring
- 重构日志库


<a name="v1.1.9"></a>
## [v1.1.9] - 2023-06-21
### Performance Improvements
- 客户端重试时忽略无可用实例情况，并减少创建picker


<a name="v1.1.8"></a>
## [v1.1.8] - 2023-06-21
### Code Refactoring
- 修改reason直接使用google.rpc.Code

### Performance Improvements
- 优化客户端创建流的重试逻辑


<a name="v1.1.7"></a>
## [v1.1.7] - 2023-04-26
### Bug Fixes
- 修改动态更改监听未应用的问题

### Features
- 添加通过governor接口修改配置文件


<a name="v1.1.6"></a>
## [v1.1.6] - 2023-04-21
### Bug Fixes
- 修改日志linux下的格式


<a name="v1.1.4"></a>
## [v1.1.4] - 2023-04-21
### Bug Fixes
- 添加remote logger level key
- 修复config并发崩溃的问题
- 修复remote日志未跟随日志初始化的问题
- 修复日志error等级解析失败，默认日志打印格式兼容linux

### Features
- value和values中scan添加映射yaml标签
- 环境变量配置，添加数组分割和设置分割符功能


<a name="v1.1.3"></a>
## [v1.1.3] - 2023-04-13
### Bug Fixes
- 修复多个service时server启动阻塞问题
- 修改rpc生成器引用未使用和没有stream接口时stream引用错误的问题

### Features
- map source 读取yaml标签


<a name="v1.1.2"></a>
## [v1.1.2] - 2023-03-14
### Bug Fixes
- 修复Run未初始化server
- 修改例子中protobuf编译结果和server启动方式
- 修复服务启动异常是程序无法退出问题

### Code Refactoring
- 修改init和serve流程，使得可以无需启动server使用框架的其它功能


<a name="v1.1.0"></a>
## [v1.1.0] - 2023-03-13
### Bug Fixes
- cloneMap函数中gob注册map[string]string{}类型，并修改CoverInterfaceToStringMap名字为CoverInterfaceMapToStringMap
- 修复grpc transport关闭时不能正常退出的问题
- 修复xnet工具获取IP地址时返回非ipv4的地址
- 修复中间件初始化函数取错名字和空指针问题
- 修复配置key中含有tag时配置api失效的问题
- 修改服务发现和负载的逻辑问题
- 修复server为空时恐慌和注册时服务未初始化的问题
- 修改rpc生成器FullMethod错误bug

### Code Refactoring
- 重构server启动函数的逻辑
- refactoring production code

### Features
- 完善example
- 添加注册实例metadata常量serverKink
- 将service注册放入yggdrasil入口的选项中


<a name="v0.1.3"></a>
## [v0.1.3] - 2022-08-02
### Code Refactoring
- change framework log format

### Features
- log add context filed
- log support add fields and log interface add w function


<a name="v0.1.2"></a>
## [v0.1.2] - 2022-07-29
### Features
- server info add get host function
- grpc md support header and trailer
- add before start hook
- add BeforeStart hook
- add error code filed to the Reason
- server listen info add to config and add function to get endpoints

### Performance Improvements
- optimization polaris resolver
- add reason error help function


<a name="v0.1.1"></a>
## [v0.1.1] - 2022-07-14
### Bug Fixes
- config value cannot read []string
- **polaris:** polaris config_source load config error

### Code Refactoring
- modify polaris-go dependence to third_party
- change grpc log init mode

### Features
- actively stop function
- config key tag support -

### Performance Improvements
- optimize polaris balancer log


<a name="v0.1.0"></a>
## v0.1.0 - 2022-07-06
### Features
- initial version source code


[Unreleased]: https://github.com/imkuqin-zw/yggdrasil/compare/contrib/polaris/v1.3.8...HEAD
[contrib/polaris/v1.3.8]: https://github.com/imkuqin-zw/yggdrasil/compare/contrib/otelexporters/v1.3.8...contrib/polaris/v1.3.8
[contrib/otelexporters/v1.3.8]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.3.8...contrib/otelexporters/v1.3.8
[v1.3.8]: https://github.com/imkuqin-zw/yggdrasil/compare/contrib/gorm/v1.3.8...v1.3.8
[contrib/gorm/v1.3.8]: https://github.com/imkuqin-zw/yggdrasil/compare/contrib/gorm/v1.3.7...contrib/gorm/v1.3.8
[contrib/gorm/v1.3.7]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.3.7...contrib/gorm/v1.3.7
[v1.3.7]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.3.5...v1.3.7
[v1.3.5]: https://github.com/imkuqin-zw/yggdrasil/compare/example/v1.3.4...v1.3.5
[example/v1.3.4]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.3.4...example/v1.3.4
[v1.3.4]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.3.3...v1.3.4
[v1.3.3]: https://github.com/imkuqin-zw/yggdrasil/compare/contrib/polaris/v1.3.3...v1.3.3
[contrib/polaris/v1.3.3]: https://github.com/imkuqin-zw/yggdrasil/compare/example/v1.3.3...contrib/polaris/v1.3.3
[example/v1.3.3]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.3.1...example/v1.3.3
[v1.3.1]: https://github.com/imkuqin-zw/yggdrasil/compare/contrib/otelexporters/v1.3.1...v1.3.1
[contrib/otelexporters/v1.3.1]: https://github.com/imkuqin-zw/yggdrasil/compare/contrib/polaris/v1.3.1...contrib/otelexporters/v1.3.1
[contrib/polaris/v1.3.1]: https://github.com/imkuqin-zw/yggdrasil/compare/contrib/gorm/v1.3.1...contrib/polaris/v1.3.1
[contrib/gorm/v1.3.1]: https://github.com/imkuqin-zw/yggdrasil/compare/contrib/redis/v1.3.1...contrib/gorm/v1.3.1
[contrib/redis/v1.3.1]: https://github.com/imkuqin-zw/yggdrasil/compare/example/v1.3.1...contrib/redis/v1.3.1
[example/v1.3.1]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.2.2...example/v1.3.1
[v1.2.2]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.2.1...v1.2.2
[v1.2.1]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.1.9...v1.2.1
[v1.1.9]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.1.8...v1.1.9
[v1.1.8]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.1.7...v1.1.8
[v1.1.7]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.1.6...v1.1.7
[v1.1.6]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.1.4...v1.1.6
[v1.1.4]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.1.3...v1.1.4
[v1.1.3]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.1.2...v1.1.3
[v1.1.2]: https://github.com/imkuqin-zw/yggdrasil/compare/v1.1.0...v1.1.2
[v1.1.0]: https://github.com/imkuqin-zw/yggdrasil/compare/v0.1.3...v1.1.0
[v0.1.3]: https://github.com/imkuqin-zw/yggdrasil/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/imkuqin-zw/yggdrasil/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/imkuqin-zw/yggdrasil/compare/v0.1.0...v0.1.1
