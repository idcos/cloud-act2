# HA方案

### salt-master配置同步、act2-master高可用方案，已测试方案如下：

SALT-MASTER、SALT-API安装
目前测试环境中直接使用yum安装测试，后期可根据版本需求配置yum源安装：
yum -y install salt-master salt-api


### SALT-MASTER 配置同步

salt-master配置同步方案采用unison + inotifywait，同步如下文件、目录：
【文件】/etc/salt/master
【目录】/etc/salt/master.d
【目录】/etc/salt/pki/master
【目录】/srv/salt （默认state系统目录）
【目录】/srv/pillar （默认pillar目录）
inotifywait -mrq --format '%Xe %w%f' -e modify,create,delete,move,attrib /etc/salt/master /etc/salt/master.d /etc/salt/pki/master /srv/salt /srv/pillar
http://gitlab.idcos.com:8082/automation-scripts/saltstack/blob/master/salt-script-tools/salt-master_sync_ha.sh


### SALt-API配置

```
cat > /etc/salt/master.d/salt-api.conf <<EOF
external_auth:
pam:
  saltapi:
    - .*
​
rest_cherrypy:
port: 8090
host: 0.0.0.0
# disable_ssl: True
ssl_crt: /etc/ssl/certs/cert.pem
ssl_key: /etc/ssl/certs/key.pem
EOF

```


SALT-API接口测试

```
[root@localhost ~]# curl -ki https://127.0.0.1:8090/login -H "Accept: application/json" -d username='saltapi' -d password='******' -d eauth='pam'
HTTP/1.1 200 OK
Content-Length: 176
Access-Control-Expose-Headers: GET, POST
Vary: Accept-Encoding
Server: CherryPy/unknown
Allow: GET, HEAD, POST
Access-Control-Allow-Credentials: true
Date: Wed, 05 Sep 2018 05:01:37 GMT
Access-Control-Allow-Origin: *
X-Auth-Token: fefb6c6c1e8098b263367ca39f88780f80c10b7c
Content-Type: application/json
Set-Cookie: session_id=fefb6c6c1e8098b263367ca39f88780f80c10b7c; expires=Wed, 05 Sep 2018 15:01:37 GMT; Path=/
​
{"return": [{"perms": [".*"], "start": 1536123697.972801, "token": "fefb6c6c1e8098b263367ca39f88780f80c10b7c", "expire": 1536166897.972803, "user": "saltapi", "eauth": "pam"}]}[root@localhost ~]#

```

### NGINX负载均衡、健康检查、KEEPALIVE

负载均衡通过配置upstream实现；

健康检查通过下载官方RPM源码包重新打RPM安装包并加入第三方nginx_upstream_check_module模块实现；

```
patch -p1 < ../nginx_upstream_check_module-master/check_1.12.1+.patch
wget http://gitlab.idcos.com:8082/automation-scripts/saltstack/raw/master/salt-script-tools/nginx-1.12.2-1.el7_4_upstream_check.ngx.x86_64.rpm
yum install nginx-1.12.2-1.el7_4_upstream_check.ngx.x86_64.rpm -y
```


构建新的conf文件(`/etc/nginx/conf.d/upstream.conf`)，填入类似下面内容：

```
upstream act2 {
  server 192.168.1.19:6868;
  server 192.168.1.20:6868;
  #check interval=3000 rise=2 fall=5 timeout=1000 type=tcp;
  #ip_hash;
}


server {
        listen 80;

        location / {
                proxy_pass http://act2;
                proxy_set_header Host $host:$server_port;
        }
}
```

上述 `192.168.1.19:6868` 和 `192.168.1.20:6868` 需要修改为实际部署时的地址 

NGINX之前通过keepalive配置提供VIP访问。