
## 命令行统一接口描述

和王苏商量后，定义了下面的脚本执行方法， 命令行格式类似如下：


```
命令 子命令 下发列表 脚本文件路径 '脚本参数'
```

- 命令: 即 act2ctl 命令
- 子命令: 有 run, file, conf 等
- 下发列表: 'target_id', 可以用 ip 或者 设备的 uuid ，地址之间用英文逗号分隔
- 脚本文件路径: salt://usr/yunji/srv.sh

	* 支持的协议有：salt, http, ssh

		- http 协议是指从 http 的目标服务器去获取文件
		- salt 协议是：表示 proxy 需要使用 salt 的方式下发文件到 minion 中
		- ssh 协议是：先从 act2master 推送脚本内容到 act2proxy，然后 act2proxy 拷贝文件到目标机器，目标机器上执行内容（目标机器的路径，默认是 / tmp 路径）

	* 要求路径必须是绝对路径, 文件名中的后缀，是有特殊业务规则:
		- .sls，使用状态应用
		- 其他，使用脚本执行或者文件下发


- 脚本参数：
	参数可选，如果有多个参数，必须使用单引号括起来



## 脚本执行

act2ctl run 'target_id' salt://usr/yunji/srv.sh 'xxxx xxx xxx xxx'

相同参数解释见上面



## 脚本文件下发

act2ctl file 'target_id, target_id' salt://srv.sh /tmp/start

相同参数解释见上面


目标路径：绝对路径，必须保证目标路径可写！


## 配置重新加载

act2 conf reload