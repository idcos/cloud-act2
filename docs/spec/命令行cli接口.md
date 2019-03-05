# 命令行cli接口

v 0.1.1



## 目标

act2ctl的操作，需要能够对自身进行管理，并能够对远程机器进行操作，然后获取到结果，显示在终端上。



## 分类

从目标上看，总的来说，可以分为下面两类：

- 自身，管理act2相关的。
- 机器，管理底层机器操作相关的。



## 术语

`proxy`：代理服务器，工程通过代理服务器来管理机器
`idc`: 逻辑机房，在`act2ctl`中，用来管理一批机器的对象，通常是`act2 proxy`可以管理的机器的集合
`entityId`: 系统上的唯一编码，每个系统有一个值，取值方法如下：

* windows: `wmic csproduct get uuid`
* linux： `cat /sys/class/dmi/id/product_uuid`
* aix: `uname -f`



## 命令行统一接口描述

和王苏商量后，定义了下面的脚本执行方法， 命令行格式类似如下：

```
命令 子命令 目标地址 脚本文件路径 '脚本参数'
```

- 命令: 即 act2ctl 命令
- 子命令: 有 run, file, conf 等
- 目标地址: 'target_id', 可以用 ip 或者 设备的 entityId ，地址之间用英文逗号分隔
  - 文件名中的后缀，是有特殊业务规则:
    - .sls，使用状态应用
    - 其他，使用脚本执行或者文件下发

- 脚本参数：
  参数可选，如果有多个参数，必须使用单引号括起来



## 通常参数描述

```
-c,idc: 机房信息 参数后面跟随idc的名称
-output=json|yaml|default: 输出结果用现在的默认输出，json输出或者yaml输出
-nocolor: 输出的时候，没有高亮，默认是高亮的
-p,password: 密码
-P,port: 端口
-i,ip: ip地址
-s,status: 状态，running|done
-v,verbose: 调试输出，输出运行中的更多的信息
-H,hostfile, 主机文件列表，以行分割，其值可以是ip地址
-a,arg: 参数列表，用来传递给脚本的参数信息
-C,command: 命令行
-t,type: 类型，支持salt,ssh,puppet，默认为salt
-T,timeout: 超时时间，每台机器上执行的超时时间
-f,file: 执行的脚本文件
```



## 自身

- 对act2 master的配置，并确认act2ctl是否可以和act2 master连通
    * `act2ctl config act2_cluster`
      - 配置远程的服务器的信息，`act2_cluster`为act2 master的集群地址，放在 `~/.act2.yaml` 文件中
      - 文件格式采用yaml支持
```
act2ctl:
    cluster: "http://192.168.1.20"
    auth: "basic"
    username: ""
    password: ""
```


- 对act2 master下管理的idc列表的展现，知道管理了哪些idc，并确认master是否可以idc连通
    * `act2ctl idc list`

      - 列出当前act2下的所有的`idc`，及连通状态
      - 连通状态，idc下任意一个proxy机器是可以访问的，及状态是success，全部不可以访问，则为fail
      - 输出为:  
  ```
  proxy              idc           status 
  192.168.1.20       杭州          success
  192.168.1.22       北京          fail
  ```
​		用1个`\t`进行数据隔开

- 对`act2 master`的内部正在执行的作业情况进行输出（并输出状态）(先不做)(0.3)
    * `act2ctl job list -c idc [-s status]`

      - 列出某个idc下的作业执行情况，默认是所有状态，默认逆排序，最新50条数据
      - `-s status`，列出status状态： `running | done`
      - 输出为：
      ```
      job_id    status        start_time                  elapse
      xxxx      running       2018-08-11 10:21:33           
      yyyy      done          2018-08-11 10:20:33           10s
      ```



## 机器

- 列出idc下面的机器

  - `act2ctl host list -c idc,idc2`
    - `-c idc,idc2`，可以接多个idc
    - 只列出puppet或者salt下的机器列表
    - 输出结果:

```
ip				  entityId		 idc		type
192.168.1.20		xxxx	     杭州		   salt
192.168.1.21		yyy	         杭州        puppet
```




- 可以对idc下的一批机器下发脚本并执行和命令行远程执行

    - 执行脚本
        - `act2ctl run ip_list|-H hostfile [-t salt] -f /tmp/run.sh [-a arg] [--nocolor] [-ouput=json|yaml] [-T timeout]` 
            - `-t salt` 支持`salt，puppet，ssh`，默认为`salt`
            - `-f /tmp/run.sh` `/tmp/run.sh`可以为相对路径，后缀名如果为`sls`的时候，需要通过salt来下发，注意和`-t`的参数必须保持兼容
            - `ip_list`为ip列表，用逗号分隔，与-h参数互斥
            - `-H hostfile`，主机文件，当ip信息比较多的时候，使用主机文件，每行一个ip地址
            - `-a arg` 参数列表信息，注意，最好使用单引号括起来，注意bash的转义
            - `--nocolor` 没有高亮的结果输出
            - `-output`，输出结果格式 `yaml|json|default`，json格式和yaml格式的输出见最底部
            - `-T timeout`，超时时间，单位秒，默认30秒
    - 执行命令行
        - `act2ctl run ip_list|-H hostfile [-t salt] -C 'df -h' [--nocolor] [-ouput=json|yaml] [-T timeout]` 
            - `-C 'df -h'` 命令行参数执行，注意和默认的salt调用脚本的执行互斥，`-a`和`-C`是互斥项

- 可以对idc下的一批机器下发文件

    * `act2ctl file ip_list|-H hostfile -t salt /tmp/run.sh|http://10.0.0.1/run.sh /tmp/file`
      * 第一个为源，在`act2ctl`服务器上，或者可以用`http`地址
      * 第二个为目标路径
      * 输出结果

- 可以对一批机器下发目录（先不做）(0.3)
  * `act2ctl file ip_list|-H hostfile -t salt -r /tmp/ /tmp/`
    * 如果目标机器上目录不存在，则自动创建子目录
    * 输出结果

- 可以在idc的机器之间进行文件传输（先不做）(0.3)

    * `act2ctl file --trans salt://192.168.11:/tmp/run.sh ssh://user:password@192.168.1.20:/tmp/run.sh`
      * 表示通过`salt`协议，从`192.168.1.11`机器上复制文件`/tmp/run.sh`到目标机器`192.168.1.20`上的`/tmp/run.sh`，复制到目标机器的方式为`ssh`方式

- 支持通过job id获取执行结果（先不做）(0.3)
    - `act2ctl job get job_id`
    - 输出结果为
```
[1] 20:00:11(3s)	[success]		192.168.1.1 
Filesystem               Size  Used Avail Use% Mounted on
/dev/mapper/centos-root   17G  2.4G   15G  15% /
devtmpfs                 908M     0  908M   0% /dev
tmpfs                    920M   16K  920M   1% /dev/shm
tmpfs                    920M   66M  854M   8% /run
tmpfs                    920M     0  920M   0% /sys/fs/cgroup
/dev/sda1               1014M  142M  873M  14% /boot
tmpfs                    184M     0  184M   0% /run/user/0

```



### json格式

```json
[
    {
    	"start_time": "20:00:11",
        "elapse": "3s",
        "status": "success",
        "ip": "192.168.1.1",
        "stdout": "Filesystem               Size  Used Avail Use% Mounted on\n/dev/mapper/centos-root   17G  2.4G   15G  15% /\ndevtmpfs                 908M     0  908M   0% /dev\ntmpfs 920M   16K  920M   1% /dev/shm\ntmpfs                    920M   66M  854M   8% /run\ntmpfs                    920M     0  920M   0% /sys/fs/cgroup\n/dev/sda1 1014M  142M  873M  14% /boot\ntmpfs                    184M     0  184M   0% /run/user/0"
	},
]

```



### yaml格式

```yaml
- 
 start_time: "20:00:11"
 elapse: 3s
 status: success
 ip: "192.168.1.1"
 stdout: "Filesystem               Size  Used Avail Use% Mounted on\n/dev/mapper/centos-root   17G  2.4G   15G  15% /\ndevtmpfs                 908M     0  908M   0% /dev\ntmpfs 920M   16K  920M   1% /dev/shm\ntmpfs                    920M   66M  854M   8% /run\ntmpfs                    920M     0  920M   0% /sys/fs/cgroup\n/dev/sda1 1014M  142M  873M  14% /boot\ntmpfs                    184M     0  184M   0% /run/user/0"
```











