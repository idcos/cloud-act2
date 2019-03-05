//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package main

import (
	"fmt"
	"os"

	"idcos.io/cloud-act2/cmd/act2ctl/install"
	"idcos.io/cloud-act2/cmd/act2ctl/proxy"

	"github.com/spf13/cobra"
	"idcos.io/cloud-act2/build"
	"idcos.io/cloud-act2/cmd/act2ctl/action/cmd"
	"idcos.io/cloud-act2/cmd/act2ctl/action/file"
	"idcos.io/cloud-act2/cmd/act2ctl/config"
	"idcos.io/cloud-act2/cmd/act2ctl/host"
	"idcos.io/cloud-act2/cmd/act2ctl/idc"
	"idcos.io/cloud-act2/cmd/act2ctl/record"
)

var rootCmd = &cobra.Command{
	Use:   "act2ctl",
	Short: "cloud-act2 client controller tool",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var (
	v bool
)

func Execute() {

	rootCmd.AddCommand(host.GetHostCommand())
	rootCmd.AddCommand(idc.GetIDCCmd())
	rootCmd.AddCommand(cmd.GetRunCmd())
	rootCmd.AddCommand(file.GetFileCmd())
	rootCmd.AddCommand(record.GetRecordCmd())
	rootCmd.AddCommand(config.GetConfigCmd())
	rootCmd.AddCommand(install.GetSaltInstallCmd())
	rootCmd.AddCommand(proxy.GetProxyCmd())
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "act2ctl version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("branch: %s\ncommit: %s\ndate: %s\n", build.GitBranch, build.Commit, build.Date)
		},
	})

	//Global value
	rootCmd.PersistentFlags().BoolVarP(&v, "verbose", "v", false, "act2ctl verbose")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	Execute()
}
