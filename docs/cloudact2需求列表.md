## cloudact2 需求列表


### 脚本执行

- bat，sh 脚本执行
- python 脚本执行
- sls 的状态应用

### 执行结果

- 监听执行状态
- 存入数据库


### API

- chi 框架构建
- chi 下的 controller 的编写
	- 命令行下的功能都需要
	- golang 的 debug 的 pprof
	- 状态信息暴露（给予 ha 功能使用）
- 认证机制


### 命令行

- 服务启动
- 服务配置重新加载
- 服务关闭
- 命令执行
- 获取任务执行结果
- 当前正在执行的任务列表
- 任务执行历史
- 任务停止
- 任务暂停
- 任务恢复
- 任务重试
- 任务重新执行
- 任务队列状态
- 列出 master 下的机器列表
- 查询机器所在的 master
- 查询机器是否可以 ping 通


### 其他

- 黑白名单功能
- 调用的第三方及功能权限控制
- HA 功能
	- salt
		- salt-master 的健康检查
		- salt-master 的异常切换
		- salt 的文件同步
	- act2
		- 由 nginx 来实现

### 辅助

- golang 内部状态的信息
- 配置文件定义及读取
- 日志路径及格式
- minion 主机与 master 关联维护
	- 通过 minion 查找 master 主机
	- 通过 master 查找 minion 机器列表
	- salt
		- minion 机器变更通知 act2
