//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package saltclient

// Config ...
type Config struct {
	Server        string
	Username      string
	Password      string
	Debug         bool
	sslSkipVerify bool
}

//ScriptKwargParam salt kwarg result body
type ScriptKwargParam struct {
	Source          string            `json:"source"`
	Args            string            `json:"args"`
	RunAs           string            `json:"runas"`
	Timeout         int               `json:"timeout"`
	Password        string            `json:"password"`
	Act2ProxyServer string            `json:"act2_proxy_server"`
	UseVT           bool              `json:"use_vt"`
	Env             map[string]string `json:"env"`
}

type FileKwargParam struct {
	Source   string `json:"path"`
	Dest     string `json:"dest"`
	MakeDirs bool   `json:"makedirs"`
}

//StateKwargParam salt state sls
type StateKwargParam struct {
	Mods string `json:"mods"`
}

//MinionsPostBody salt /minions result body
type MinionsPostBody struct {
	Fun     string      `json:"fun"`
	Tgt     string      `json:"tgt"`
	TgtType string      `json:"tgt_type"`
	Kwarg   interface{} `json:"kwarg"`
}

//LoginResp salt /login result body
type LoginResp struct {
	Return []TokenResp `json:"return"`
}

//TokenResp salt token result body
type TokenResp struct {
	Perms  []string `json:"perms"`
	Start  float64  `json:"start"`
	Token  string   `json:"token"`
	Expire float64  `json:"expire"`
	User   string   `json:"user"`
	Eauth  string   `json:"eauth"`
}

//RunResp salt run result body
type RunResp struct {
	Return []JidResp `json:"return"`
}

//JidResp salt /job/${jid} result body
type JidResp struct {
	Jid     string   `json:"jid"`
	Minions []string `json:"minions"`
}

//ScriptResultResp salt execution result body
type ScriptResultResp struct {
	Return []map[string]interface{} `yaml:"return"`
}

//MinionReturnResp 单个对象的执行结果
type MinionReturnResp struct {
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Retcode int    `json:"retcode"`
}

//StateResultResp state execution result body
type StateResultResp struct {
	Info []StateInfoResp `yaml:"info"`
}

type StateInfoResp struct {
	Result  map[string]StateMinionResult `yaml:"Result"`
	Minions []string                     `yaml:"Minions,omitempty"`
}

type StateMinionResult struct {
	Success bool        `yaml:"success"`
	Return  interface{} `yaml:"return"`
}

/**
文件下发返回的数据值如下：

```json
{
  "info": [
    {
      "Arguments": [
        {
          "__kwarg__": true,
          "dest": "/tmp/test05.sh",
          "path": "salt://decb6568-864d-4b83-b7c6-8be5dfa328e8.sh"
        }
      ],
      "Function": "cp.get_file",
      "Minions": [
        "F30823F4-4A35-4CFF-82F9-183E97D73921",
        "F88E8415-D66E-4519-9F30-7A5D71024271"
      ],
      "Result": {
        "F30823F4-4A35-4CFF-82F9-183E97D73921": {
          "return": "/tmp/test05.sh"
        },
        "F88E8415-D66E-4519-9F30-7A5D71024271": {
          "return": "/tmp/test05.sh"
        }
      },
      "StartTime": "2018, Aug 27 10:36:31.200401",
      "Target": "F88E8415-D66E-4519-9F30-7A5D71024271,F30823F4-4A35-4CFF-82F9-183E97D73921",
      "Target-type": "list",
      "User": "salt-api",
      "jid": "20180827103631200401"
    }
  ],
  "return": [
    {
      "F30823F4-4A35-4CFF-82F9-183E97D73921": "/tmp/test05.sh",
      "F88E8415-D66E-4519-9F30-7A5D71024271": "/tmp/test05.sh"
    }
  ]
}

```
*/
type FileResultResp struct {
	Info   []map[string]interface{} `yaml:"info"`
	Return []map[string]string      `yaml:"return"`
}
