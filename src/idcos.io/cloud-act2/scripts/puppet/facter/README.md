## README

- 把文件 `oslang.rb` 复制到 `puppetserver` 下的 `/etc/puppetlabs/code/modules/health/lib/facter`，仍旧命名为 `oslang.rb`

```bash
[root@puppet-master-124 facter]# pwd
/etc/puppetlabs/code/modules/health/lib/facter
[root@puppet-master-124 facter]# ls
oslang.rb
```

- 手工执行下发

```bash
puppet agent -t
```

此时系统会下发到远程的 `agent` 端，在 `linux` 上，地址为：`/opt/puppetlabs/puppet/cache/lib/facter/oslang.rb`

- 验证

```bash
curl -X GET http://10.0.20.224:8080/pdb/query/v4/nodes/pocdemo1/facts
```
输出结果中含有：

```
{
      "certname": "pocdemo1",
      "name": "oslang",
      "value": ""en_US.UTF-8"",
      "environment": "production"
}
```

或者登录到 `agent` 端，执行下面代码

```bash
[root@puppet-master-124 facter]# export FACTERLIB="/opt/puppetlabs/server/data/puppetserver/lib/facter"
[root@puppet-master-124 facter]# facter oslang
"en_us.utf-8"
```
