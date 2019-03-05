# cloud-act2


## 依赖

- MySQL（5.6+）
- Go1.11及以上版本


## 安装
### 拉取源代码

```bash
git clone https://github.com/idcos/cloud-act2 cloud-act2
```


### *nix安装编译环境

1. 登录golang官网或者golang中国官方镜像下载最新的稳定版本的go安装包并安装。
```bash
$ wget https://dl.google.com/go/go1.12.linux-amd64.tar.gz
# 解压缩后go被安装在/usr/local/go
$ sudo tar -xzv -f ./go1.12.linux-amd64.tar.gz -C /usr/local/
```

2. 配置go环境变量

```bash
echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOROOT/bin' >> ~/.bashrc
source ~/.bashrc
```


### 编译

进入源代码根目录:

```bash
cd cloud-act2
export GOPATH=`pwd`
make
```

编译完毕后，项目根目录下多了`cmd`目录，其中`cmd/bin`目录下包含了多个可执行文件。

```bash
$ tree cmd
cmd
├── bin
│   ├── act2ctl
│   ├── cloud-act2-server
│   └── salt-event
└── etc
    ├── acl
    │   ├── model.conf
    │   └── policy.csv
    ├── cloud-act2-proxy.yaml
    └── cloud-act2.yaml

3 directories, 7 files
```

### 初始化数据

1. 导入SQL文件初始化数据库 将`changelog/init-opensource.sql`导入MySQL。
2. act2-master的配置文件`/usr/yunji/cloud-act2/cloud-act2.yaml`
3. act2-proxy的配置文件`/usr/yunji/cloud-act2/cloud-act2-proxy.yaml`


### 运行

master启动

``` bash
$ /usr/yunji/cloud-act2/bin/cloud-act2-server web start
```

proxy启动

``` bash
$ /usr/yunji/cloud-act2/bin/cloud-act2-server web start -c /usr/yunji/cloud-act2/etc/cloud-act2-proxy.yaml 
```


## 第三方


### go相关工具

- dep
- revive
- swagger


```bash

go get -u github.com/golang/dep/cmd/dep
go get -u github.com/mgechev/revive
go get -u github.com/go-swagger/go-swagger/cmd/swagger

```



### 其他工具

- [cloc](http://cloc.sourceforge.net/)
- [sonar](https://docs.sonarqube.org/latest/)



### 版权

Copyright 2019 Cloud J Tech, Inc and other contributors
Licensed under the GPLv3
