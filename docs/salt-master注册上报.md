## salt-master注册上报

### 起源

`act2`在执行的时候，需要知道具体机器所在的`salt-master`，从而准确发送给`salt-master`发送执行指令，设想的方案有两种：

1. 由`salt-master`机器上的脚本推送`master`和`minions`给`act2`服务器，`act2`服务器保存相关的数据
2. 由`act2`服务器上配置`salt-master`有哪些，然后`act2`服务器主动去拉`master`和`minions`的信息

经过讨论，目前选择的方案是方案1： 因为推送的方案，是大部分开源软件的选择方案（叶子语），所以推送是更好的方式。

## 脚本

脚本为`heartbeat.py`: 

```bash
[root@salt-master02 ~]# python scripts/heartbeat.py -h
usage: heartbeat.py [-h] [-d [DEBUG]] [--log-level LOG_LEVEL]
                    [--log-file LOG_FILE] [--salt-server SALT_SERVER]
                    [--salt-user SALT_USER] [--salt-password SALT_PASSWORD]
                    [--act2-server ACT2_SERVER] [--srv-type SERVICE_TYPE]
                    [--idc IDC] [-v]

salt-heartbeat

optional arguments:
  -h, --help            show this help message and exit
  -d [DEBUG], --debug [DEBUG]
                        debug, default: false
  --log-level LOG_LEVEL
                        log level, debug, info, warning, error, default: info
  --log-file LOG_FILE   log file, if not set, using stdout instead
  --salt-server SALT_SERVER
                        salt server address, default: http://localhost:8001
  --salt-user SALT_USER
                        salt user, default: salt-api
  --salt-password SALT_PASSWORD
                        salt password
  --act2-server ACT2_SERVER
                        cloud act2 server address, eg: http://192.168.1.9
  --srv-type SERVICE_TYPE
                        service type, eg salt|puppet|openssh
  --idc IDC             idc information
  -v, --version         show program's version number and exit
```

## 上报接口信息

```yaml

/register:
    post：
        summary: "注册上报"
        description: "注册上报"
        parameters:
            - name: "master"
              description: "master设备"
              required: true
              type: "object"
              $ref: "#/definitions/Master"
            - name: "minions"
              description: "master设备"
              required: true
              type: "array"
              item:
                $ref: "#definitions/Minion"



Master:
    type: "object"
    properties:
        - name: "sn"
          type: "string"
          required: true
        - name: "server"
          type: "string"
          required: true
        - name: "status"
          type: "string"
          description: "running or stopped"
          required: true
        - name: "idc"
          type: "string"
          required: true
        - name: "type"
          type: "string"
          required: true
          description: "salt|puppet|openssh"
        - name: "options"
          $ref: "#/definitions/Options"
          required: false


Minion:
    type: "object"
    properties:
        - name: "sn"
          type: "string"
          required: true
        - name: "status"
          type: "string"
          required: true
        - name: "ips"
          type: "array"
          required: true
          item: "string"
        
Options:
    type: "object"
    properties:
        - name: "username"
          type: "string"
          required: true
        - name: "password"
          type: "string"
          required: true


```




## 检测回调

act2需要检测salt-master的上报情况，从而确定salt-master是否还存活以及master和minion的信息

目前上报的定的策略是：每30分钟上报一次， 上报的时候，nginx会转到其中一台机器上，下一次上报的时候，未必一定会上报到上一台机器上。所以上报的时间信息，需要记录到数据库中，在不同机器之间，才可以共享。

考虑到act2存在多个系统，如果heartbeat.py进程异常或者网络异常，这act2无法获取到对应的结果，从而认定salt是挂掉的，为此act2需要一定的时间间隔去检测，salt-master是否挂掉了。

act2自己定时去检测，那么多个act2同时启动的时候，只能一个act2去定时轮询，而不能多个act2同时去检测是否超时。那么其中跑act2定时任务的挂掉后，需要将定时任务转移到其他机器上继续跑。


实现方案：


1、所有act2的server都跑定时任务
2、检测salt-master有效性时间间隔，需要配置文件里面配置，设计一个默认值，比方说2个小时，需要比心跳上报的时间要长。（这个需要实施控制）
2、有冲突的情况下（一个register进来，另一个判定到master已经timeout），谁先达到，谁先处理，无论哪种情况，都需要写日志。


