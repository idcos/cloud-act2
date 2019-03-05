//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Act2Cluster struct {
	Cluster      string `yaml:"cluster"`
	AuthType     string `yaml:"auth"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	WaitInterval int    `yaml:"wait_interval"`
	// 默认值为2018.3.3版本
	SaltVersion string `yaml:"salt_version"`
}

type Act2Ctl struct {
	Act2Cluster `yaml:"act2ctl"`
}

// DbConfig db config
type DbConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`

	Name            string `yaml:"name"` // db name
	Encoding        string `yaml:"encoding,omitempty"`
	Debug           bool   `yaml:"debug,omitempty"`
	PoolSize        int    `yaml:"pool_size,omitempty"`
	PoolIdleSize    int    `yaml:"poll_idle_size"`
	ConnMaxLifeLime int    `yaml:"conn_max_life_time"`
}

// 注册监控
type Heartbeat struct {
	TimeoutInterval  string `yaml:"timeout_interval"`  // 注册检测定时任务表达式
	RegisterInterval string `yaml:"register_interval"` // 注册间隔时间
	ReportInterval   string `yaml:"report_interval"`   // 上报时间间隔
}

// Logger logger结构
type Logger struct {
	Facility      string `yaml:"facility,omitempty"`
	LogProtocol   string `yaml:"rsyslog_protocol,omitempty"`
	LogServer     string `yaml:"rsyslog_server,omitempty"`
	LogLevel      string `yaml:"level,omitempty"`
	LogFile       string `yaml:"file,omitempty"`
	LogDateFormat string `yaml:"datefmt,omitempty"`
	HttpLogFile   string `yaml:"http,omitempty"`
}

type JobExpire struct {
	// 检测过期时间的定时表达式
	TimeoutInterval string `yaml:"timeout_interval"`
	// 过期时间，单位s
	Expire int `yaml:"expire"`
}

// Act2 config information
type Act2 struct {
	ClusterServer     string `yaml:"cluster_server"`
	ProxyServer       string `yaml:"proxy_server,omitempty"`
	FileReversed      bool   `yaml:"file_reversed"`
	TimeoutCorrection int    `yaml:"timeout_correction"`
	DelayReport       int    `yaml:"delay_report,omitempty"`
	ResultCallRetry   uint   `yaml:"result_call_retry,omitempty"`
	ACL               bool   `yaml:"acl,omitempty"`
	ACLAuth           string `yaml:"acl_auth,omitempty"`
}

type SaltMaster struct {
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	SYSPath    string `yaml:"syspath"`
	Server     string `yaml:"server"`
	SchemeURI  string `yaml:"scheme_uri"`
	Python     string `yaml:"python"`
	SaltPath   string `yaml:"salt_path"`   // 默认使用/usr/bin/salt
	SaltConfig string `yaml:"salt_config"` // salt的默认配置路径
}

type SSHConfig struct {
	PassMethod string   `yaml:"pass_method"`
	PassServer string   `yaml:"pass_server"`
	Cache      string   `yaml:"cache"`
	Ciphers    []string `yaml:"ciphers"`
	Lang       []string `yaml:"lang"`
}

type PuppetConfig struct {
	RabbitMQ       string `yaml:"rabbitmq"`
	PuppetDBServer string `yaml:"puppetdb"`
	ReplyQueue     string `yaml:"reply_queue"`
	Ruby           string `yaml:"ruby"`
}

type RedisServer struct {
	Server string `yaml:"server"`
	Port   int    `yaml:"port"`
}

type RedisAddr struct {
	Masters []RedisServer `yaml:"master"`
	Slavers []RedisServer `yaml:"slaver"`
}

type RedisConfig struct {
	Cluster     bool      `yaml:"cluster"`
	Addr        RedisAddr `yaml:"addrs"`
	Password    string    `yaml:"password"`
	MaxIdle     int       `yaml:"max_idle"`
	MaxActive   int       `yaml:"max_active"`
	IdleTimeout int       `yaml:"idle_timeout"`
}

// ACLUser 设定ACL的用户
type ACLUser struct {
	Name     string   `yaml:"name"`
	Password string   `yaml:"password"`
	Role     []string `yaml:"role"`
}

type HTTPSConfig struct {
	Open     bool   `yaml:"open"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

// Config config 数据结构体
type Config struct {
	IsMasterInfo string `yaml:"is_master"`
	Independent  bool   `yaml:"independent"`
	DependRedis  bool   `yaml:"depend_redis"`
	Port         string `yaml:"port"`
	ChannelType  string `yaml:"channel_type"`
	PubSub       string `yaml:"pub_sub"`
	IDC          string `yaml:"idc"`
	ProjectPath  string `yaml:"project_path"`
	CacheType    string `yaml:"cache_type"`
	CryptoType   string `yaml:"crypto_type"`
	CryptoKey    string `yaml:"crypto_key"`

	Db        DbConfig     `yaml:"db,omitempty"`
	Logger    Logger       `yaml:"log,omitempty"`
	Heartbeat Heartbeat    `yaml:"heartbeat,omitempty"`
	JobExpire JobExpire    `yaml:"job_expire"`
	Act2      Act2         `yaml:"act2,omitempty"`
	Salt      SaltMaster   `yaml:"salt,omitempty"`
	SSH       SSHConfig    `yaml:"ssh,omitempty"`
	Puppet    PuppetConfig `yaml:"puppet,omitempty"`

	Redis   RedisConfig `yaml:"redis,omitempty"`
	ACLUser []ACLUser   `yaml:"acl_user,omitempty"`
	HTTPS   HTTPSConfig `yaml:"https"`
}

func (c *Config) IsMaster() bool {
	return c.IsMasterInfo == "true"
}

func (c *Config) GetName() string {
	if c.IsMaster() {
		return "cloud-act2"
	} else {
		return "cloud-act2-proxy"
	}
}

var (
	// Conf global config of act2
	Conf *Config
)

func loadMasterDefaultConfig() Config {
	conf := Config{
		IsMasterInfo: "true",
		// 不独立部署
		Independent: false,
		// 默认部支持redis
		DependRedis: false,
		// master的默认端口为6868
		Port: ":6868",
		// 默认使用salt的通道类型
		ChannelType: "salt",
		// 发布订阅的通道
		PubSub: "redis",
		// 工程部署路径
		ProjectPath: "/usr/yunji",
		// 缓存类型
		CacheType: "redis",

		// 加密类型
		CryptoType: "aead",

		// 数据库中的默认配置
		Db: DbConfig{
			Name:     "cloud-act2",
			Encoding: "utf8mb4",
			PoolSize: 20,
			Port:     "3306",
		},

		// Logger的配置
		Logger: Logger{
			Facility:      "file",
			LogProtocol:   "tcp",
			LogLevel:      "info",
			LogFile:       "/usr/yunji/cloud-act2/logs/cloud-act2.log",
			HttpLogFile:   "/usr/yunji/cloud-act2/logs/cloud-act2-http.log",
			LogDateFormat: `2006-01-02 15:04:05`,
		},

		// 默认心跳的配置
		Heartbeat: Heartbeat{
			TimeoutInterval:  "10m",
			RegisterInterval: "30m",
		},

		// 默认任务超时时间
		JobExpire: JobExpire{
			TimeoutInterval: "30m",
			Expire:          10,
		},

	}
	return conf
}

func loadProxyDefaultConfig() Config {
	conf := Config{
		IsMasterInfo: "false",
		// master的默认端口为6868
		Port: ":5555",
		// 默认使用salt的通道类型
		ChannelType: "salt",
		// 发布订阅的通道
		PubSub: "redis",
		// 工程部署路径
		ProjectPath: "/usr/yunji",
		// 缓存类型
		CacheType: "redis",

		// 加密类型
		CryptoType: "aead",

		// Logger的配置
		Logger: Logger{
			Facility:      "file",
			LogProtocol:   "tcp",
			LogLevel:      "info",
			LogFile:       "/usr/yunji/cloud-act2/logs/cloud-act2-proxy.log",
			HttpLogFile:   "/usr/yunji/cloud-act2/logs/cloud-act2-proxy-http.log",
			LogDateFormat: `2006-01-02 15:04:05`,
		},

		Salt: SaltMaster{
			Username:   "salt-api",
			SYSPath:    "/srv/salt",
			Server:     "http://127.0.0.1:8001",
			Python:     "/usr/bin/python",
			SaltPath:   "/usr/bin/salt",
			SaltConfig: "/ect/salt",
		},

		Puppet: PuppetConfig{
			Ruby: "/opt/puppetlabs/puppet/bin/ruby",
		},

		SSH: SSHConfig{
			PassMethod: "native",
			PassServer: "http://localhost:5000",
			Cache:      "/var/cache/act2",
			Ciphers:    []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-cbc", "3des-cbc"},
		},

		// 默认心跳的配置
		Heartbeat: Heartbeat{
			ReportInterval:   "30s",
			RegisterInterval: "30m",
		},

		// Act2配置
		Act2: Act2{
			TimeoutCorrection: 3,
			DelayReport:       0,
			ResultCallRetry:   3,
		},

	}
	return conf
}

func setStringDefault(src *string, defaultVal string) {
	if *src == "" {
		*src = defaultVal
	}
}

func setIntDefault(src *int, defaultVal int) {
	if *src == 0 {
		*src = defaultVal
	}
}

func setUintDefault(src *uint, defaultVal uint) {
	if *src == 0 {
		*src = defaultVal
	}
}

func mergeMasterConfig(masterConfig *Config) *Config {
	defaultConfig := loadMasterDefaultConfig()

	setStringDefault(&masterConfig.Port, defaultConfig.Port)
	setStringDefault(&masterConfig.ChannelType, defaultConfig.ChannelType)
	setStringDefault(&masterConfig.PubSub, defaultConfig.PubSub)
	setStringDefault(&masterConfig.ProjectPath, defaultConfig.ProjectPath)
	setStringDefault(&masterConfig.CacheType, defaultConfig.CacheType)
	setStringDefault(&masterConfig.CryptoType, defaultConfig.CryptoType)

	setStringDefault(&masterConfig.Logger.Facility, defaultConfig.Logger.Facility)
	setStringDefault(&masterConfig.Logger.LogProtocol, defaultConfig.Logger.LogProtocol)
	setStringDefault(&masterConfig.Logger.LogLevel, defaultConfig.Logger.LogLevel)
	setStringDefault(&masterConfig.Logger.LogFile, defaultConfig.Logger.LogFile)
	setStringDefault(&masterConfig.Logger.HttpLogFile, defaultConfig.Logger.HttpLogFile)
	setStringDefault(&masterConfig.Logger.LogDateFormat, defaultConfig.Logger.LogDateFormat)

	setStringDefault(&masterConfig.Db.Name, defaultConfig.Db.Name)
	setStringDefault(&masterConfig.Db.Port, defaultConfig.Db.Port)
	setStringDefault(&masterConfig.Db.Encoding, defaultConfig.Db.Encoding)
	setIntDefault(&masterConfig.Db.PoolSize, defaultConfig.Db.PoolSize)

	setStringDefault(&masterConfig.Heartbeat.TimeoutInterval, defaultConfig.Heartbeat.TimeoutInterval)
	setStringDefault(&masterConfig.Heartbeat.RegisterInterval, defaultConfig.Heartbeat.RegisterInterval)

	setStringDefault(&masterConfig.JobExpire.TimeoutInterval, defaultConfig.JobExpire.TimeoutInterval)
	setIntDefault(&masterConfig.JobExpire.Expire, defaultConfig.JobExpire.Expire)

	return masterConfig
}

func mergeProxyConfig(proxyConfig *Config) *Config {
	defaultConfig := loadProxyDefaultConfig()

	setStringDefault(&proxyConfig.Port, defaultConfig.Port)
	setStringDefault(&proxyConfig.ChannelType, defaultConfig.ChannelType)
	setStringDefault(&proxyConfig.PubSub, defaultConfig.PubSub)
	setStringDefault(&proxyConfig.ProjectPath, defaultConfig.ProjectPath)
	setStringDefault(&proxyConfig.CacheType, defaultConfig.CacheType)
	setStringDefault(&proxyConfig.CryptoType, defaultConfig.CryptoType)

	setStringDefault(&proxyConfig.Logger.Facility, defaultConfig.Logger.Facility)
	setStringDefault(&proxyConfig.Logger.LogProtocol, defaultConfig.Logger.LogProtocol)
	setStringDefault(&proxyConfig.Logger.LogLevel, defaultConfig.Logger.LogLevel)
	setStringDefault(&proxyConfig.Logger.LogFile, defaultConfig.Logger.LogFile)
	setStringDefault(&proxyConfig.Logger.HttpLogFile, defaultConfig.Logger.HttpLogFile)
	setStringDefault(&proxyConfig.Logger.LogDateFormat, defaultConfig.Logger.LogDateFormat)

	setStringDefault(&proxyConfig.Salt.Username, defaultConfig.Salt.Username)
	setStringDefault(&proxyConfig.Salt.SYSPath, defaultConfig.Salt.SYSPath)
	setStringDefault(&proxyConfig.Salt.Server, defaultConfig.Salt.Server)
	setStringDefault(&proxyConfig.Salt.Python, defaultConfig.Salt.Python)
	setStringDefault(&proxyConfig.Salt.SaltPath, defaultConfig.Salt.SaltPath)
	setStringDefault(&proxyConfig.Salt.SaltConfig, defaultConfig.Salt.SaltConfig)

	setStringDefault(&proxyConfig.Puppet.Ruby, defaultConfig.Puppet.Ruby)

	setStringDefault(&proxyConfig.SSH.PassMethod, defaultConfig.SSH.PassMethod)
	setStringDefault(&proxyConfig.SSH.PassServer, defaultConfig.SSH.PassServer)
	setStringDefault(&proxyConfig.SSH.Cache, defaultConfig.SSH.Cache)
	if len(proxyConfig.SSH.Ciphers) == 0 {
		copy(proxyConfig.SSH.Ciphers, defaultConfig.SSH.Ciphers)
	}

	setStringDefault(&proxyConfig.Heartbeat.ReportInterval, defaultConfig.Heartbeat.ReportInterval)
	setStringDefault(&proxyConfig.Heartbeat.RegisterInterval, defaultConfig.Heartbeat.RegisterInterval)

	setIntDefault(&proxyConfig.Act2.TimeoutCorrection, defaultConfig.Act2.TimeoutCorrection)
	setIntDefault(&proxyConfig.Act2.DelayReport, defaultConfig.Act2.DelayReport)
	setUintDefault(&proxyConfig.Act2.ResultCallRetry, defaultConfig.Act2.ResultCallRetry)

	return proxyConfig
}

func LoadConfig(filename string) error {
	conf := &Config{}
	err := LoadConfigFile(filename, conf)
	if err != nil {
		return err
	}

	if conf.IsMasterInfo == "true" {
		conf = mergeMasterConfig(conf)
	} else {
		conf = mergeProxyConfig(conf)
	}

	Conf = conf
	return nil
}

// LoadConfig load config
func LoadConfigFile(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("read config file error %s \n", err)
		return err
	}
	err = yaml.Unmarshal(data, v)
	if err != nil {
		fmt.Printf("laod yaml config error %s\n", err)
		return err
	}
	return nil
}
