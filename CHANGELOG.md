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
