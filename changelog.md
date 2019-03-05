## 0.7

- 优化 timeout执行任务超时等待
- 优化 callback回掉时每次起goroutine
- 支持 通过IP的白名单访问API机制
- 优化 文件下发支持二进制格式
- 支持 在act2ctl中添加一个删除proxy的命令
- 支持 支持循环下发给proxy的策略
- 修复 通道为ssh时的文件下发，参数位置会被修改


## 0.6

- 支持 列出idc下面proxy和机器的展现
- 支持 查看任务列表和任务执行状态信息
- 支持 ssh的密钥的远程访问
- 修复 目前存在上报的时候数据重复
- 优化 http等log的日志打印
- 修复 act2ctl idc host -l返回异常信息，json unmarshal异常
- 修复 当minion关闭后proxy中会不断轮询salt-master
- 修复 sls脚本执行完后一直会轮询为
- 修复 cpu过高加载问题
- 修复 执行100台ssh命令，proxy会挂掉




## 0.5

- act2ctl 支持默认参数
- 支持 windows 和 linux 下的输出结果编码控制
- 添加 oom 启动的控制
- 优化 idc 输出列表的对齐控制
- fix 输出结果异常时无 ip 显示问题
- 优化输出结果的 title 的顺序调整
- fix act2ctl 列表下参数顺序的变化
- act2ctl 的版本参数携带分支信息
- 优化系统 enttiy_id 的获取方式
- 添加 puppet 下的系统 oslang.rb 语言信息获取获取脚本



## 0.4

- puppet 下的 mco 协议对接和实现
- 修改 act2-proxy 回调 act2-master 的回调地址组装由 act2-proxy 处理
- aix 下的 entityId 获取实现
- 修改 act2ctl 执行参数结构变化


## 0.3

- 修订脚本执行和文件下发
- 支持 openssh 通道的脚本和文件
- 文件下发反向代理实现
- 使用系统 entityId 作为设备唯一编码处理，改进 ip 使用方案
- 支持 act2ctl 的命令行执行脚本和文件下发


## 0.2

- 执行结果回调 conf
- 均分下发执行策略
- 通过 uuid 或者 ip 进行执行下发
- salt-api 获取主机执行结果
- proxy 代理脚本执行
- golang 的 debug 信息添加
- 获取任务执行结果


## 0.1

初始化发布


