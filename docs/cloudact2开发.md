## cloudact2开发

### 项目位置

http://gitlab.idcos.com:8082/CloudConf/cloud-act2


### 方案

- 使用keepalived作为ha方案（主备切换）
- 使用kingshard作为数据库分库分表处理方案（透明代理）





### 开发

- github.com/voidint/gbb的包信息编译
- restful框架：github.com/go-chi/chi
- Log: cloudboot的log接口和具体实现有beegolog和logrus
- 需要可以接受一个信号，dump出go运行时的堆栈信息到当前目录下（用pid+时间作为文件名）
- proxy和master做到一个程序中，通过cloudact2 proxy|cloudact2 master来运行不同的客户端
- 命令采用cli的库进行控制: (github.com/urfave/cli)
- 使用golint进行代码规范检测（https://github.com/golang/lint）
- golang的版本使用最新的release版本（1.10.3）：https://golang.org/dl/
- 使用godep进行包管理，相关包放入vendor
- http的client，使用beego的http client库: https://github.com/astaxie/beego/blob/08c3ca642eb437f1dd65a617e107bdd9aafb66b0/httplib/README.md
- golang的错误管理，使用: github.com/pkg/errors  (https://godoc.org/github.com/pkg/errors)(https://www.youtube.com/watch?v=lsBF58Q-DnY)
- 需要用changelog.md记录当前版本的变化




openssh可以参考:

- https://github.com/idcos/cloud-cli
- http://gitlab.idcos.com:8082/CloudJ/cli




## ide

### vscode

推荐使用vscode

vscode下`command + shift + p`执行`ext install go` 

### GoLand

https://www.jetbrains.com/go/


### vim
https://github.com/fatih/vim-go



