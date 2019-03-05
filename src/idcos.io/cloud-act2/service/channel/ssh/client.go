//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package ssh

import (
	"idcos.io/cloud-act2/crypto"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/service/channel/common"
	serviceCommon "idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/encoding"
	"idcos.io/cloud-act2/utils/fileutil"
	"idcos.io/cloud-act2/utils/generator"
	"idcos.io/cloud-act2/utils/httputil"
	"idcos.io/cloud-act2/utils/promise"

	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/astaxie/beego/httplib"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/service/channel/ssh/auth"
	"idcos.io/cloud-act2/service/channel/ssh/system"
	"idcos.io/cloud-act2/utils"

	switchssh "github.com/shenbowei/switch-ssh-go"
)

var (
	ErrAuthentication = errors.New("username error or password error")
)

type SSHClient struct {
}

func NewSSHClient() *SSHClient {
	return &SSHClient{}
}

func (sshClient *SSHClient) callbackSystemResult(host serviceCommon.ExecHost, osType string) error {
	logger := getLogger()

	url := strings.TrimRight(config.Conf.Act2.ClusterServer, "/") + define.HostEntityUri
	body := map[string]string{
		"entityId": host.EntityID,
		"hostId":   host.HostID,
		"osType":   osType,
		"proxyId":  host.ProxyID,
	}

	resp, err := httputil.HttpPost(url, body)
	if err != nil {
		logger.Error("host update system id", "error", err)
		return err
	}

	logger.Info(fmt.Sprintf("host %s update", host.HostID), "result", resp)
	return nil
}

func (sshClient *SSHClient) newSystemSession(session *ssh.Session, osType string, hostID string) system.Systemer {
	var systemer system.Systemer
	switch osType {
	case define.Win:
		systemer = system.NewWindowSession(session)
	case define.Aix:
		systemer = system.NewAixSession(session)
	case define.Linux:
		systemer = system.NewLinuxSession(session)
	case define.Cisco:
		fallthrough
	case define.Junior:
		fallthrough
	case define.Huawei:
		fallthrough
	case define.H3c:
		fallthrough
	case define.Network:
		systemer = system.NewNetworkSession(hostID)
	default:
		// 默认为linux系统
		systemer = system.NewLinuxSession(session)
	}
	return systemer
}

func (sshClient *SSHClient) fetchEntityIDToMaster(execHost serviceCommon.ExecHost, user string, password string) {
	logger := getLogger()

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: func(string, net.Addr, ssh.PublicKey) error {
			return nil
		},
		// ssh连接的超时时间，默认为2s
		Timeout: time.Duration(2) * time.Second,
	}

	host := execHost.HostIP
	var port string
	if execHost.HostPort == 0 {
		port = "22"
	} else {
		port = fmt.Sprintf("%d", execHost.HostPort)
	}

	logger.Debug(fmt.Sprintf("start to dial client, host:%s, port: %s, network: tcp, user: %s, auth:%#v", host, port, sshConfig.User, sshConfig.Auth))

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), sshConfig)
	if err != nil {
		logger.Error("connect to remote", "server", host, "port", port, "error", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		logger.Error("failed to create session", "error", err)
		return
	}
	defer session.Close()

	systemer := sshClient.newSystemSession(session, execHost.OsType, execHost.HostID)
	systemID, err := systemer.SystemID()
	if err != nil {
		logger.Error("failed to get system id", "error", err)
		return
	}

	err = sshClient.callbackSystemResult(execHost, systemID)
	if err != nil {
		logger.Error("failed to backup system id", "error", err)
		return
	}
}

func getSSHAuthMethod(host, port, runAs, password string) (ssh.AuthMethod, error) {
	var sshAuth auth.SSHAuth
	switch config.Conf.SSH.PassMethod {
	case "native":
		sshAuth = auth.NewNativeAuth(password)
	case "uri":
		sshAuth = auth.NewHttpAuth(config.Conf.SSH.PassServer)
	case "keyUri":
		sshAuth = auth.NewHttpKeyAuth(config.Conf.SSH.PassServer)
	}

	return sshAuth.Auth(host, port, runAs)
}

func (sshClient *SSHClient) NetworkExecute(execHost serviceCommon.ExecHost, params common.ExecScriptParam) (*bytes.Buffer, *bytes.Buffer, error) {
	logger := getLogger()

	user := params.RunAs
	password := params.Password
	ipPort := fmt.Sprintf("%s:%v", execHost.HostIP, execHost.HostPort)

	logger.Debug("network information", "user", user, "password", password, "ip port", ipPort)

	client := crypto.GetClient()
	password, err := client.Decode(password)
	if err != nil {
		logger.Error("ssh client", "error", err)
		return nil, nil, err
	}

	var cmds []string
	cmds = append(cmds, params.Script)
	result, err := switchssh.RunCommands(user, password, ipPort, cmds...)
	if err != nil {
		logger.Error("get ssh brand", "error", err)
		return nil, nil, err
	}

	if strings.HasPrefix(result, params.Script) {
		result = result[len(params.Script):]
	}

	return bytes.NewBufferString(result), nil, nil
}

//GetClient 获取client
func (sshClient *SSHClient) GetClient(execHost serviceCommon.ExecHost, runAs, password string) (*ssh.Client, error) {
	logger := getLogger()

	sshAuthMethod, err := getSSHAuthMethod(execHost.HostIP, fmt.Sprintf("%v", execHost.HostPort), runAs, password)
	if err != nil {
		return nil, err
	}

	// 需要独立，在windows下共用同一个client连接ssh-server(https://github.com/PowerShell/Win32-OpenSSH)，会有一些问题
	if execHost.HostID == "" && config.Conf.Act2.ClusterServer != "" {
		promise.NewGoPromise(func(chan struct{}) {
			promise.NewGoPromise(func(chan struct{}) {
				sshClient.fetchEntityIDToMaster(execHost, runAs, password)
			}, nil)
		}, nil)
	}

	sshConfig := &ssh.ClientConfig{
		User: runAs,
		Auth: []ssh.AuthMethod{
			sshAuthMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// ssh连接的超时时间，默认为5s，考虑private认证会需要较长的时间
		Timeout: time.Duration(5) * time.Second,
	}

	sshConfig.Ciphers = append(sshConfig.Ciphers, config.Conf.SSH.Ciphers...)

	host := execHost.HostIP
	var port string
	if execHost.HostPort == 0 {
		port = "22"
	} else {
		port = fmt.Sprintf("%d", execHost.HostPort)
	}

	logger.Debug(fmt.Sprintf("start to dial client, host:%s, port: %s, network: tcp, user: %s, password %s, auth:%#v", host, port, sshConfig.User, password, sshConfig.Auth))

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), sshConfig)
	if err != nil {
		logger.Error("connect to remote", "server", host, "port", port, "error", err)

		if _, ok := err.(*net.OpError); ok {
			return nil, err
		} else {
			return nil, ErrAuthentication
		}
	}
	return client, nil
}

func (sshClient *SSHClient) GetSession(execHost serviceCommon.ExecHost, runAs, password string) (session *ssh.Session, err error) {
	client, err := sshClient.GetClient(execHost, runAs, password)
	if err != nil {
		return nil, err
	}

	session, err = client.NewSession()
	return
}

func (sshClient *SSHClient) GetSSHExeuctorClient(execHost serviceCommon.ExecHost, runAs, password, pattern string) (c SshSessionExecutor, err error) {
	logger := getLogger()

	client, err := sshClient.GetClient(execHost, runAs, password)
	if err != nil {
		return nil, err
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		logger.Error("failed to create session", "error", err)
		return nil, err
	}

	var executor SshSessionExecutor
	if pattern == "script" {
		executor = &sshScriptSession{
			client:   client,
			session:  session,
			execHost: execHost,
		}
	} else if pattern == "file" {
		executor = &sshFileSession{
			client:   client,
			session:  session,
			execHost: execHost,
		}
	} else {
		err = errors.New("pattern not in (script,file)")
		logger.Error(err.Error())
		return nil, err
	}

	return executor, nil
}

func (sshClient *SSHClient) HostExecute(execHost serviceCommon.ExecHost, params common.ExecScriptParam) (*bytes.Buffer, *bytes.Buffer, error) {
	logger := getLogger()

	client := crypto.GetClient()
	password, err := client.Decode(params.Password)
	if err != nil {
		logger.Error("ssh client host execute", "error", err)
		return nil, nil, err
	}

	executor, err := sshClient.GetSSHExeuctorClient(execHost, params.RunAs, password, params.Pattern)
	if err != nil {
		return nil, nil, err
	}
	defer executor.Close()

	stdout, _, err := executor.Execute(params)
	if err != nil {
		logger.Error("failed to execute", "error", err)
		return nil, nil, err
	}

	logger.Debug("ssh host exec success", "stdout", string(stdout))
	return bytes.NewBuffer(stdout), nil, nil
}

func (sshClient *SSHClient) Execute(execHost serviceCommon.ExecHost, params common.ExecScriptParam) (*bytes.Buffer, *bytes.Buffer, error) {
	// 主机和网络，采用不同的执行方式
	if execHost.OsType == define.Cisco || execHost.OsType == define.Junior || execHost.OsType == define.Huawei || execHost.OsType == define.H3c {
		return sshClient.NetworkExecute(execHost, params)
	} else {
		return sshClient.HostExecute(execHost, params)
	}
}

type SshSessionExecutor interface {
	Execute(common.ExecScriptParam) ([]byte, []byte, error)
	Close()
}

type sshFileSession struct {
	client   *ssh.Client
	session  *ssh.Session
	execHost serviceCommon.ExecHost
}

func (c *sshFileSession) Close() {
	c.session.Close()
	c.client.Close()
}

func (s *sshFileSession) configFile(fileContent []byte, target string) ([]byte, []byte, error) {
	scp := NewScp(s.session)

	fileContent, err := fileutil.GetDataURI(fileContent)
	if err != nil {
		return nil, nil, err
	}

	err = scp.PushFile(fileContent, target)
	return nil, nil, err
}

func (s *sshFileSession) httpSingleFile(url string, fileName string, target string, timeout int) error {
	httpRequest := httplib.Get(url).SetTimeout(3*time.Second, time.Duration(timeout)*time.Second)
	resp, err := httpRequest.DoRequest()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scp := NewScp(s.session)

	target = filepath.Join(target, fileName)

	return scp.PushFileReader(resp.Body, target)
}

func (s *sshFileSession) httpFile(fileContent string, fileName string, target string, timeout int) ([]byte, []byte, error) {
	var urls []string
	err := json.Unmarshal([]byte(fileContent), &urls)
	if err != nil {
		return nil, nil, err
	}

	// 如果传递多个文件，则target必须是目标机器的目录，否则前面文件会被后面的target给覆盖
	// 依据url的在json数组中的顺序依次处理
	for _, url := range urls {
		err := s.httpSingleFile(url, fileName, target, timeout)
		if err != nil {
			return nil, nil, err
		}
	}

	// 复制远程的服务器上的文件，然后scp到远程服务器上
	return nil, nil, nil
}

func getEncodingData(content string, encodingType string) ([]byte, error) {
	logger := getLogger()

	encodingType = strings.TrimSpace(strings.ToLower(encodingType))

	var result []byte
	var err error
	if encodingType == "utf-8" || encodingType == "utf8" || encodingType == "" {
		result = []byte(content)
	} else {
		result, err = encoding.EncodingTo([]byte(content), encodingType)
		if err != nil {
			logger.Error("encoding", "error", err, "target_encoding", encodingType)
			return nil, err
		}
	}
	return result, nil
}

func (s *sshFileSession) Execute(param common.ExecScriptParam) ([]byte, []byte, error) {
	var target string
	var fileName string
	if t, ok := param.Params["target"]; !ok {
		return nil, nil, errors.New("file not have target")
	} else {
		target = t.(string)
	}
	if f, ok := param.Params["fileName"]; !ok {
		return nil, nil, errors.New("file not have fileName")
	} else {
		fileName = f.(string)
	}

	if param.ScriptType == "conf" {
		script, err := getEncodingData(param.Script, param.Encoding)
		if err != nil {
			return nil, nil, err
		}
		return s.configFile(script, target)
	} else if param.ScriptType == "url" {
		return s.httpFile(param.Script, fileName, target, param.Timeout)
	} else {
		// TODO: 不存在的
	}

	return nil, nil, nil
}

type sshScriptSession struct {
	client   *ssh.Client
	session  *ssh.Session
	execHost serviceCommon.ExecHost
}

func (c sshScriptSession) Close() {
	c.session.Close()
	c.client.Close()
}

func (s *sshScriptSession) createScpAndExecName(execHost serviceCommon.ExecHost, scriptType string) (scpName string, execName string) {
	if execHost.IsWindows() {
		var ext string
		if scriptType == "bat" || len(scriptType) == 0 {
			ext = ".bat"
		} else if scriptType == "python" {
			ext = ".py"
		}
		filename := generator.GenUUID() + ext
		scpName = filepath.Join("/tmp/", filename)
		execName = "C:\\tmp\\" + filename
	} else {
		scpName = "/tmp/" + generator.GenUUID()
		execName = scpName
	}

	return
}

func (s *sshScriptSession) buildCommand(param common.ExecScriptParam, execName string, scriptArgs string) (string, error) {
	scriptLangs := getScriptLangs()

	// TODO: 参数数据的编码，参数数据的转义，参数数据中的空格处理
	var command string
	if s.execHost.IsWindows() {
		if param.ScriptType == define.BatType || len(param.ScriptType) == 0 {
			command = execName + " " + scriptArgs
		} else if lang, ok := scriptLangs[param.ScriptType]; ok {
			command = lang + "" + execName + " " + scriptArgs
		} else {
			return "", fmt.Errorf("unsupported command type %s", param.ScriptType)
		}

	} else {
		// python, bash必须在path中
		if param.ScriptType == define.BashType || param.ScriptType == define.ShellType || len(param.ScriptType) == 0 {
			command = "bash " + execName + " " + scriptArgs
		} else if lang, ok := scriptLangs[param.ScriptType]; ok {
			command = lang + " " + execName + " " + scriptArgs
		} else {
			// 其他语言类型，修改为具体bash -c方式执行
			return "", fmt.Errorf("unsupported script type %s", param.ScriptType)
		}
	}
	return command, nil
}

func getScriptLangs() map[string]string {
	scriptLangs := map[string]string{
		define.PerlType:   "perl",
		define.PythonType: "python",
		define.RubyType:   "ruby",
	}

	// 扩展支持的语言
	if len(config.Conf.SSH.Lang) > 0 {
		for _, lang := range config.Conf.SSH.Lang {
			langType := filepath.Base(lang)
			scriptLangs[langType] = lang
		}
	}

	return scriptLangs
}

// 在ssh协议中，openssh扩展了ssh协议，见下面链接的 `2.1. connection: Channel write close extension "eow@openssh.com"``
// https://cvsweb.openbsd.org/src/usr.bin/ssh/PROTOCOL?annotate=HEAD
// 该协议的扩展，仅可以在openssh的client端上使用，第三方不推荐使用,golang的ssh协议是没有
// 这方面支持的，所以为了能够扩展ssh的带参数的执行，需要将文件复制到远程服务器上，然后
// 调用远程服务器上的命令来执行，获取执行后的结果。
func (s *sshScriptSession) Execute(param common.ExecScriptParam) ([]byte, []byte, error) {
	logger := getLogger()

	script, err := getEncodingData(param.Script, param.Encoding)
	if err != nil {
		logger.Error("script encoding", "error", err)
		return nil, nil, err
	}

	// TODO: 此处会打印password信息，存在安全隐患
	logger.Debug(fmt.Sprintf("start to exec script, param: %s", dataexchange.ToJsonString(param)))

	newSession, err := s.client.NewSession()
	if err != nil {
		logger.Error("create new session", "error", err)
		return nil, nil, err
	}
	fileSession := sshFileSession{session: newSession}

	scpname, execname := s.createScpAndExecName(s.execHost, param.ScriptType)

	// 临时文件放置位置
	_, _, err = fileSession.configFile(script, scpname)
	if err != nil {
		newSession.Close()
		logger.Error("run script upload file", "error", err)
		return nil, nil, err
	}
	newSession.Close()

	// 此处不支持中文参数编码
	var scriptArgs string
	if args, ok := param.Params["args"]; ok {
		scriptArgs = args.(string)
	}

	// 保留，目前不需要做转义处理
	scriptArgs = utils.EscapeArgs(scriptArgs, s.execHost.IsWindows())

	command, err := s.buildCommand(param, execname, scriptArgs)
	if err != nil {
		logger.Error("script command", "error", err)
		return nil, nil, err
	}
	logger.Debug("will run ", "script", string(script), "command", command)

	out, err := s.session.CombinedOutput(command)
	if err != nil {
		logger.Error("failed to run", "error", err)
		return nil, nil, err
	}

	out, err = encoding.DecodingTo(out, param.Encoding)
	if err != nil {
		logger.Error("failed to encoding to utf-8", "error", err)
		return nil, nil, err
	}

	logger.Info("script run result", "out", string(out))

	return out, nil, nil
}
