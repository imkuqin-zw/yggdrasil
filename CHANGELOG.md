# 1.1.8 (2023/06/21)

### Features

- refactor: 修改reason直接使用google.rpc.Code
- perf: 优化客户端创建流的重试逻辑
- docs: CHANGELOG.md

# 1.1.7 (2023/04/26)

### Features

- feat: 添加通过governor接口修改配置文件
- docs: changelog

### Bug Fixes

- fix: 修改动态更改监听未应用的问题

# 1.1.6 (2023/04/21)

### Features

- style: 更新changelog

### Bug Fixes

- fix: 修改日志linux下的格式

# 1.1.4 (2023/04/21)

### Features

- feat: value和values中scan添加映射yaml标签
- feat: 环境变量配置，添加数组分割和设置分割符功能
- CHANGELOG.md

### Bug Fixes

- fix: 添加remote logger level key
- fix: 修复config并发崩溃的问题
- fix: 修复remote日志未跟随日志初始化的问题
- fix: 修复日志error等级解析失败，默认日志打印格式兼容linux

# 1.1.3 (2023/04/13)

### Features

- feat: map source 读取yaml标签

### Bug Fixes

- fix: 修复多个service时server启动阻塞问题
- fix: 修改rpc生成器引用未使用和没有stream接口时stream引用错误的问题

# 1.1.2 (2023/03/14)

### Features

- style: 更新changelog
- refactor: 修改init和serve流程，使得可以无需启动server使用框架的其它功能
- style: 修改changelog

### Bug Fixes

- fix: 修复Run未初始化server
- fix: 修改例子中protobuf编译结果和server启动方式
- fix: 修复服务启动异常是程序无法退出问题

# 1.1.0 (2023/03/13)

### Features

- feat: 完善example
- feat: 添加注册实例metadata常量serverKink
- feat: 将service注册放入yggdrasil入口的选项中
- refactor: 重构server启动函数的逻辑
- refactor: refactoring production code
- style: add change log

### Bug Fixes

- fix: cloneMap函数中gob注册map[string]string{}类型，并修改CoverInterfaceToStringMap名字为CoverInterfaceMapToStringMap
- fix: 修复grpc transport关闭时不能正常退出的问题
- fix: 修复xnet工具获取IP地址时返回非ipv4的地址
- fix: 修复中间件初始化函数取错名字和空指针问题
- fix: 修复配置key中含有tag时配置api失效的问题
- fix: 修改服务发现和负载的逻辑问题
- fix: 修复server为空时恐慌和注册时服务未初始化的问题
- fix: 修改rpc生成器FullMethod错误bug

# 0.1.3 (2022/08/12)

### Features

- refactor: change framework log format
- feat: log add context filed
- feat: log support add fields and log interface add w function

# 0.1.2 (2022/08/12)

### Features

- style: copyright
- style: add copyright
- style: LICENSE
- style: add copyright
- perf: optimization polaris resolver
- feat: server info add get host function
- feat: grpc md support header and trailer
- feat: add before start hook
- feat: add BeforeStart hook
- perf: add reason error help function
- feat: add error code filed to the Reason
- feat: server listen info add to config and add function to get endpoints

# 0.1.1 (2022/08/12)

### Features

- refactor: modify polaris-go dependence to third_party
- feat: actively stop function
- perf: optimize polaris balancer log
- feat: config key tag support -
- refactor: change grpc log init mode

### Bug Fixes

- fix: config value cannot read []string
- fix(polaris): polaris config_source load config error

# 0.1.0 (2022/08/12)

### Features

- feat: initial version source code
- Initial commit
