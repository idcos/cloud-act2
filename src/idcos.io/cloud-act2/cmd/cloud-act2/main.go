//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package main

import (
	"fmt"
	"idcos.io/cloud-act2/utils/fileutil"
	"idcos.io/cloud-act2/utils/httputil"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"idcos.io/cloud-act2/build"

	"github.com/spf13/cobra"
	"idcos.io/cloud-act2/config"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/log"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/server/master"
	"idcos.io/cloud-act2/server/middleware"
	"idcos.io/cloud-act2/server/proxy"
	channelCommon "idcos.io/cloud-act2/service/channel/common"
	"idcos.io/cloud-act2/service/heartbeat"
	"idcos.io/cloud-act2/service/job"
	"idcos.io/cloud-act2/redis"
	"idcos.io/cloud-act2/service/schedule"
	"idcos.io/cloud-act2/timingwheel"
	"idcos.io/cloud-act2/utils"
	"idcos.io/cloud-act2/utils/oom"
)

var (
	port          string
	configPath    string
	logLevel      string
	skipOomAdjust bool
	pidFile       string
)

func commonInit() error {
	// init validate
	utils.InitValidate()

	fmt.Printf("config path '%s'\n", configPath)
	err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("load config error %s", err)
		return err
	}
	log.InitLogger("master")

	// 初始化timingwheel
	timingwheel.Load()

	err = config.LoadCommonData()
	if err != nil {
		return err
	}

	if config.Conf.CacheType == define.Redis {
		err = redis.InitRedisClient(config.Conf.Redis)
		if err != nil {
			return err
		}
	}
	return nil
}

func masterInit() error {
	// load database connection
	err := model.OpenConn(config.Conf)
	if err != nil {
		return err
	}

	//timeout Recovery
	err = job.RestartRecovery()
	if err != nil {
		return err
	}

	//webhook.Load()
	return nil
}

func proxyInit() {
	log.InitLogger("proxy")
	if !config.Conf.Independent {
		go func() {
			err := heartbeat.RegisterSaltInfo(true)
			if err != nil {
				fmt.Printf("register salt info %s", err)
			}
		}()
	}
}

func writePidFile() {
	if len(pidFile) == 0 {
		return
	}

	pidPath := filepath.Base(pidFile)
	if ok, _ := fileutil.FileExists(pidPath); !ok {
		err := os.Mkdir(pidPath, 0644)
		if err != nil {
			fmt.Printf("pid directory make error %s\n", err)
			return
		}
	}

	pid := fmt.Sprintf("%d", os.Getpid())
	ioutil.WriteFile(pidFile, []byte(pid), 0644)
}

func startServer(appName string, localPort string, r http.Handler) error {
	logger := log.L()
	showDebug := config.Conf.Logger.LogLevel == "debug"
	httputil.InitHttpLibSetting("cloud-act2-"+appName, showDebug)

	writePidFile()

	if !strings.Contains(localPort, ":") {
		localPort = ":" + localPort
	}
	fmt.Printf("will listen port %s\n", localPort)
	//logger.Info("start", "app", appName, "version", ctx.App.Version)
	logger.Info("will listen", "port", localPort)

	https := config.Conf.HTTPS
	if https.Open {
		certFile := https.CertFile
		keyFile := https.KeyFile
		if len(certFile) == 0 {
			certFile = "/usr/yunji/cloud-act2/etc/cert/server.crt"
		}
		if len(keyFile) == 0 {
			keyFile = "/usr/yunji/cloud-act2/etc/cert/server.key"
		}
		return http.ListenAndServeTLS(localPort, certFile, keyFile, r)
	} else {
		return http.ListenAndServe(localPort, r)
	}

}

func masterStart() error {
	r := master.NewRouter()
	if r == nil {
		return fmt.Errorf("could not make router")
	}
	appName := "master"

	localPort := config.Conf.Port
	if localPort == "" {
		localPort = port
	}
	// proxy的默认端口，5555
	if localPort == "" {
		localPort = define.MasterDefaultPort
	}

	schedule.MasterSchedule()
	return startServer(appName, localPort, r)
}

func proxyStart() error {
	// 创建默认的salt的syspath路径
	if err := fileutil.CreateDirIfNotExist(config.Conf.Salt.SYSPath); err != nil {
		fmt.Printf("create sys path %s error %s, please created by manual\n,", config.Conf.Salt.SYSPath, err)
	}

	r := proxy.NewRouter()
	appName := "proxy"

	localPort := config.Conf.Port
	if localPort == "" {
		localPort = port
	}

	// proxy的默认端口，5555
	if localPort == "" {
		localPort = define.ProxyDefaultPort
	}

	if !config.Conf.Independent {
		fmt.Println("cloud-act2 proxy schedule start")
		// 启动web的时候，会按照时间间隔上报，每30s上报一次
		schedule.ProxySchedule()
	}

	return startServer(appName, localPort, r)
}

func webStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "start",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			commonInit()

			if config.Conf.IsMaster() {
				if err := masterInit(); err != nil {
					return err
				}
			} else {
				proxyInit()
			}

			// 初始化resultpoll
			channelCommon.Load()

			if !skipOomAdjust {
				fmt.Println("apply oom adjust")
				oom.ApplyOOMScoreAdj(0, define.OomScoreAdj)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Conf.IsMaster() {
				return masterStart()
			} else {
				return proxyStart()
			}
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			if config.Conf.IsMaster() {
				// close db
				model.CloseDb()
			}

			middleware.CloseHttpLogFile()

			timingwheel.Unload()

			return nil
		},
	}
}

func webCommands() *cobra.Command {
	webCommand := cobra.Command{
		Use:   "web",
		Short: "web",
	}
	webCommand.AddCommand(webStartCommand())
	return &webCommand
}

func version() *cobra.Command {
	cmd := cobra.Command{
		Use:   "version",
		Short: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("branch: %s\ncommit: %s\ndate: %s\n", build.GitBranch, build.Commit, build.Date)
		},
	}

	return &cmd

}

func main() {

	var rootCmd = &cobra.Command{
		Use:   "cloud-act-server",
		Short: "cloud act2 server",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
	rootCmd.AddCommand(version())

	rootCmd.PersistentFlags().StringVarP(&configPath, "conf", "c", "/usr/yunji/cloud-act2/etc/cloud-act2.yaml", "config file path")
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", ":6868", "port")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "level", "l", "debug", "log level, optional values: debug|info|warn|error")
	rootCmd.PersistentFlags().BoolVarP(&skipOomAdjust, "skipOomAdjust", "", false, "skip oom adjust")
	rootCmd.PersistentFlags().StringVarP(&pidFile, "pid", "", "", "pid file")

	rootCmd.AddCommand(webCommands())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
