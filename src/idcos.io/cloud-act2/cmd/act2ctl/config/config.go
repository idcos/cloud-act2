//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"idcos.io/cloud-act2/utils/dataexchange"
	"idcos.io/cloud-act2/utils/fileutil"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"idcos.io/cloud-act2/config"
)

const (
	configClusterUsage      = "act2ctl config -c http://nginx.conf"
	configAuthUsage         = "act2ctl config -a idcos"
	configUsernameUsage     = "act2ctl config -u root"
	configPasswordUsage     = "act2ctl config -p 1234"
	configSaltVersionUsage  = "act2ctl config -s 2018.3.3"
	configWaitIntervalUsage = "act2ctl config -i 1"

	act2CtlYamlPath = "~/.act2.yaml"
)

var (
	act2Ctl config.Act2Ctl
	once    sync.Once
)

var (
	waitInterval int
	cluster      string
	auth         string
	username     string
	password     string
	saltVersion  string
	verbose      bool
)

func showConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "show",
		Short: "show act2ctl config",
		Run:   showConfigAction,
	}
	return configCmd
}

func addConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "add",
		Short: "add act2ctl config",
		Run:   configAction,
	}

	configCmd.Flags().StringVarP(&cluster, "cluster", "c", "", configClusterUsage)
	configCmd.Flags().StringVarP(&auth, "auth", "a", "", configAuthUsage)
	configCmd.Flags().StringVarP(&username, "username", "u", "", configUsernameUsage)
	configCmd.Flags().StringVarP(&password, "password", "p", "", configPasswordUsage)
	configCmd.Flags().StringVarP(&saltVersion, "saltVersion", "s", "", configSaltVersionUsage)
	configCmd.Flags().IntVarP(&waitInterval, "waitInterval", "i", 1, configWaitIntervalUsage)
	configCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "")

	return configCmd
}

//GetConfigCmd config命令
func GetConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "config for act2ctl",
	}

	configCmd.AddCommand(showConfigCmd())
	configCmd.AddCommand(addConfigCmd())

	return configCmd
}

func showConfigAction(cmd *cobra.Command, args []string) {

	path, err := fileutil.ExpandUser(act2CtlYamlPath)
	if err != nil {
		fmt.Println("could not expand user")
		return
	}

	var act2Ctl config.Act2Ctl
	err = dataexchange.LoadYaml(path, &act2Ctl)
	if err != nil {
		fmt.Printf("load yaml config error %s", err)
		return
	}

	data, err := yaml.Marshal(act2Ctl)
	if err != nil {
		fmt.Printf("config not valid %s\n", err)
	} else {
		fmt.Printf("%s\n", string(data))
	}
}

func loadAct2CtlConfig() {
	path, _ := fileutil.ExpandUser(act2CtlYamlPath)
	if exists, _ := fileutil.FileExists(path); exists {
		dataexchange.LoadYaml(path, &act2Ctl)
	}
}

func GetAct2ctl() *config.Act2Ctl {
	once.Do(func() {
		err := LoadAct2Ctl(&act2Ctl)
		if err != nil {
			fmt.Println("load act2ctl error")
			panic(err)
		}
	})

	return &act2Ctl
}

//configAction 取消任务Action
func configAction(cmd *cobra.Command, args []string) {
	loadAct2CtlConfig()

	if waitInterval > 0 {
		act2Ctl.Act2Cluster.WaitInterval = waitInterval
	} else {
		act2Ctl.Act2Cluster.WaitInterval = 1
	}

	if cluster != "" {
		act2Ctl.Act2Cluster.Cluster = cluster
	}

	if username != "" {
		act2Ctl.Act2Cluster.Username = username
	}

	if password != "" {
		act2Ctl.Act2Cluster.Password = password
	}

	if auth != "" {
		act2Ctl.Act2Cluster.AuthType = auth
	}

	if saltVersion != "" {
		act2Ctl.Act2Cluster.SaltVersion = saltVersion
	}

	path, err := fileutil.ExpandUser(act2CtlYamlPath)
	if err != nil {
		fmt.Println(err)
	}

	// 保存act2Ctl之前，先将密码修改为base64

	password := base64.StdEncoding.EncodeToString([]byte(act2Ctl.Password))
	act2Ctl.Password = password

	err = dataexchange.SaveYaml(path, &act2Ctl)
	if err != nil {
		fmt.Println(err)
	}
}

func LoadAct2Ctl(act2Ctl *config.Act2Ctl) error {
	// load act2
	path, err := fileutil.ExpandUser(act2CtlYamlPath)
	if err != nil {
		return err
	}

	err = dataexchange.LoadYaml(path, act2Ctl)
	if err != nil {
		return err
	}

	password, err := base64.StdEncoding.DecodeString(act2Ctl.Password)
	if err != nil {
		return err
	}

	act2Ctl.Password = string(password)

	// 如果没有设置，则默认为1
	act2Ctl.Act2Cluster.WaitInterval = 1

	act2Ctl.Act2Cluster.Cluster = strings.TrimSpace(act2Ctl.Act2Cluster.Cluster)

	if act2Ctl.Act2Cluster.Cluster == "" {
		return errors.New("cluster must config first")
	}
	return nil
}
