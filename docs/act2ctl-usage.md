## act2ctl简介

act2ctl是一个act2的命令行工具，支持linux和mac系统，目前版本0.6.0


### act2ctl命令概览

在终端下输入`act2ctl -h`命令，可以看到下面输出：

```bash
$ ./act2ctl -h
NAME:
   act2ctl - A new cli application

USAGE:
   act2ctl [global options] command [command options] [arguments...]

VERSION:

  date: 2018-11-29T09:39:11+08:00
  commit: 196e597809c99fabce7214272202801b19bce221
  branch: release/0.8.1

DESCRIPTION:
   cloud-act2 client controller tool

COMMANDS:
     config   act2ctl config -c cluster_addr
     file     act2ctl file 192.168.1.1,192.168.1.2|-H hostfile src target -t ssh
     group
     idc      act2ctl idc [-hl]
     record
     run      act2ctl run 192.168.1.1,192.168.1.2|-H hostfile -t salt -f /srv/tmp/test.sh|-C 'df -h' [-a arg] [--nocolor] [-output=json|yaml] [-T timeout] [-c idc] [-u username] [-p password] [-P port] [-o osType]
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -v, --verbose  -v to open the debug
   --help, -h     show help
   -V, --version  print only the version
```



## 配置act2ctl

### `config [options] <cluster_server>`

配置`act2`的服务器访问地址。`cluster_server`是`act2`的集群对外暴露的服务地址，相关配置信息会存储在`~/.act2.yaml`文件中

#### Options

- -c -- 设置cluster地址
- -a,--auth -- 设置auth的类型，默认是basic
- -u.--username -- auth下的用户名
- -p,--password -- 密码
- -i,--waitInterval -- 查询结果的等待时间
- -s,--salt --salt的版本信息设置，通常默认即可

#### 示例

```bash
$ ./act2ctl config -c http://192.168.1.17:6868
```



## idc操作

### `idc command [options]`


### Options

- -l -- 列出所有的idc

#### 示例

```bash
$ ./act2ctl idc -l
proxy                           idc     status
http://192.168.1.17:5555        杭州    running
```

## 代理操作

### `idc proxy [option]`

### Options

- -l -- 列出所有的proxy

```bash
$ ./act2ctl idc proxy -l
proxy                           idc     status
http://192.168.1.17:5555        杭州    running
```


## 主机操作

### `idc host [option]`

显示和控制idc和proxy信息

### Options

- -l, --list -- 列出所有的主机信息
- -g, --grab -- 抓取idc下的主机信息到master上
- -c, --idc -- 抓取idc信息

#### 示例

```bash
$ ./act2ctl idc host -l
IdcName	EntityID				HostIP
杭州	0BBFE987-D1A3-4365-89E8-0FAF7F90F5D7	192.168.1.218
杭州	6C12A913-756C-4A6B-B149-35E6351BA939	192.168.1.217
```

```bash
$ ./act2ctl idc host -g -c 杭州
act2ctl grab host success
```

## 设备分组

### `group device sub_command [options] command_name` 


#### 子命令

给设备添加组，删除组或者更新组

##### 添加设备分组

给设备`192.168.1.217,192.168.1.218`添加组的名称为`test`，
```bash
$ ./act2ctl group device add -d '192.168.1.217,192.168.1.218' test
```

#### 列出所有的设备分组

```bash
$ ./act2ctl group device list
```

#### 把设备绑定到组上

```bash
$ ./actctl group device attach
```




## 脚本或命令执行

### `run [ip_list][options]`

通过`act2`服务下发命令或者脚本到远程服务器上，并进行执行, `ip_list`以英文逗号分割的ip地址列表，为远程服务器地址

### Options

- -H hostfile -- 指定要下发执行的远程服务器地址，与ip_list互斥
- -t type -- 下发时指定通道类型，type可以为salt或puppet或ssh，目前支持salt或ssh，默认为salt
- -f scriptfile -- 需要执行的脚本文件
- -a arg -- 脚本或者命令支持的参数，多个参数需要使用单引号或双引号起来，需要注意bash转义
- -c idc -- 指定idc名称，通道类型为ssh时必须指定
- -u username -- 指定执行账户，如果未指定，在系统类型为windows时默认账户Administrator，其他系统默认账户为root
- -p password -- 指定执行账户的密码，通常类型为ssh时必须指定
- -P port -- 指定通常类型为ssh时的连接端口，默认为22
- -o ostype -- 系统类型，支持windows|linux|aix
- -s scriptType -- 脚本类型，支持bash或bat或python或perl或sls，默认为bash。如果提供了系统类型，在Windows下默认为bat，其他默认为bash
- -e encoding -- 指定执行时编码类型，支持gb18030或utf-8，如果未指定，salt版本为2018.3.3时，
    ​    ​    windows系统默认时使用utf-8，在salt版本为2018.3.3之前，windows系统默认使用gb18030，其他系统默认账户为utf-8
- -C command -- 需要执行的命令，与 -f 互斥，最后与脚本执行是一样的
- -T timeout -- 脚本执行的超时时间，默认为300s
- --nocolor -- 关闭默认输出结果的高亮输出
- -ouput format -- 输出格式，支持 default或json或yaml，其中默认为default，支持高亮输出，json或yaml输出无高亮
-v, --verbose 输出调试信息
--async  异步处理，返回作业执行任务id


### 示例

#### 指定IP地址执行命令

##### 使用salt通道

```bash
$ ./act2ctl run '192.168.1.217,192.168.1.218' -C 'df -h'

[1]     192.168.1.217   18:22:21        32.97(s)                [success]
Filesystem               Size  Used Avail Use% Mounted on
/dev/mapper/centos-root   17G  1.3G   16G   8% /
devtmpfs                 908M     0  908M   0% /dev
tmpfs                    920M   12K  920M   1% /dev/shm
tmpfs                    920M  8.5M  911M   1% /run
tmpfs                    920M     0  920M   0% /sys/fs/cgroup
/dev/sda1               1014M  142M  873M  14% /boot
tmpfs                    184M     0  184M   0% /run/user/0
[2]     192.168.1.218   18:22:21        32.97(s)                [success]
Filesystem               Size  Used Avail Use% Mounted on
/dev/mapper/centos-root   17G  1.2G   16G   8% /
devtmpfs                 908M     0  908M   0% /dev
tmpfs                    920M   12K  920M   1% /dev/shm
tmpfs                    920M  8.5M  911M   1% /run
tmpfs                    920M     0  920M   0% /sys/fs/cgroup
/dev/sda1               1014M  142M  873M  14% /boot
tmpfs                    184M     0  184M   0% /run/user/0

```

#### 通过主机文件执行命令

##### 使用salt通道

```bash
$ cat /tmp/hostfile
192.168.1.217
192.168.1.218
$ ./act2ctl run -H /tmp/hostfile -C 'df -h'

[1]     192.168.1.217   18:22:21        32.97(s)                [success]
Filesystem               Size  Used Avail Use% Mounted on
/dev/mapper/centos-root   17G  1.3G   16G   8% /
devtmpfs                 908M     0  908M   0% /dev
tmpfs                    920M   12K  920M   1% /dev/shm
tmpfs                    920M  8.5M  911M   1% /run
tmpfs                    920M     0  920M   0% /sys/fs/cgroup
/dev/sda1               1014M  142M  873M  14% /boot
tmpfs                    184M     0  184M   0% /run/user/0
[2]     192.168.1.218   18:22:21        32.97(s)                [success]
Filesystem               Size  Used Avail Use% Mounted on
/dev/mapper/centos-root   17G  1.2G   16G   8% /
devtmpfs                 908M     0  908M   0% /dev
tmpfs                    920M   12K  920M   1% /dev/shm
tmpfs                    920M  8.5M  911M   1% /run
tmpfs                    920M     0  920M   0% /sys/fs/cgroup
/dev/sda1               1014M  142M  873M  14% /boot
tmpfs                    184M     0  184M   0% /run/user/0

```

#### 指定IP执行脚本


##### 使用salt通道

```bash
$ cat /tm/script.sh
#!/bin/bash
echo "hello act2"
echo "$@"

$ ./act2ctl run '192.168.1.217,192.168.1.218' -f /tmp/script.sh -a "hello world"

[1]     192.168.1.217   18:25:43        35.74(s)                [success]
hello act2
hello world
[2]     192.168.1.218   18:25:43        35.74(s)                [success]
hello act2
hello world

```

#### 给脚本传递参数

##### 使用ssh通道

```bash
$ ./act2ctl run '192.168.1.217,192.168.1.218' -t ssh -f /tmp/script.sh -a "hello world" -o linux -c 杭州
Password:

[1]     192.168.1.218   18:27:41        13.88(s)                [success]
hello act2
hello world

[2]     192.168.1.217   18:27:41        13.88(s)                [success]
hello act2
hello world

```


#### 异步任务

##### 使用ssh通道

```bash
$ ./act2ctl run '192.168.1.217,192.168.1.218' -t ssh -f /tmp/script.sh -a "hello world" -o linux -c 杭州 -async
Password:
9aa1f438-5cd4-cf64-fb49-8745cd545566
```




## 文件下发

### `file [ip_list] src target [options]`

通过`act2`服务下发文件到远程服务器上，`ip_list`以英文逗号分割的ip地址列表，为远程服务器地址

### Options

- -H hostfile -- 指定要下发执行的远程服务器地址，与ip_list互斥
- -t type -- 下发时指定通道类型，type可以为salt或puppet或ssh，目前支持salt或ssh，默认为salt
- -s scriptType -- 脚本类型，支持bash或bat或python或sls，默认为bash。如果提供了系统类型，在Windows下默认为bat，其他默认为bash
- -C command -- 需要执行的命令，与 -f 互斥，最后与脚本执行是一样的
- -a arg -- 脚本或者命令支持的参数，多个参数需要使用单引号或双引号起来，需要注意bash转义
- --nocolor -- 关闭默认输出结果的高亮输出
- -ouput format -- 输出格式，支持 default或json或yaml，其中默认为default，支持高亮输出，json或yaml输出无高亮
- -T timeout -- 脚本执行的超时时间，默认为300s
- -c idc -- 指定idc名称，通道类型为ssh时必须指定
- -o ostype -- 系统类型，支持windows|linux|aix
- -u username -- 指定执行账户，如果未指定，在系统类型为windows时默认账户Administrator，其他系统默认账户为root
- -p password -- 指定执行账户的密码，通常类型为ssh时必须指定
- -P port -- 指定通常类型为ssh时的连接端口，默认为22
- -e encoding -- 指定执行时编码类型，支持gb18030或utf-8，如果未指定，在系统类型为windows时默认为gb18030，其他系统默认账户为utf-8
- -v -- verbose，输出调试信息
--async  异步处理，返回作业执行任务id


选项类型和run基本一致

注意，`options` 信息必须放在尾部


### 示例

```bash
$ ./act2ctl file '192.168.1.217,192.168.1.218' /tmp/script.sh /tmp/script2.sh
[1]     192.168.1.217   18:39:03        36.61(s)                [success]

[2]     192.168.1.218   18:39:03        36.61(s)                [success]

$ ./act2ctl run '192.168.1.217,192.168.1.218' -t ssh -C 'ls -al /tmp/script*'  -o linux  -c 杭州
Password:
[1]     192.168.1.218   18:41:46        16.38(s)                [success]
-rw-r--r-- 1 root root 40 Oct  7 18:39 /tmp/script2.sh

[2]     192.168.1.217   18:41:46        16.38(s)                [success]
-rw-r--r-- 1 root root 40 Oct  7 18:39 /tmp/script2.sh
```

上面的script2.sh文件已经成功下发对应的服务器上
