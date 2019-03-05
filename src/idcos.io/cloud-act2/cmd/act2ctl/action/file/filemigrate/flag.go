//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package filemigrate

import "github.com/spf13/cobra"

type FileMigrateFlag struct {
	SourceIDCName  string
	SourceOSType   string
	SourceEncoding string
	SourcePort     int
	TargetIDCName  string
	TargetOSType   string
	TargetEncoding string
	TargetPort     int
	Timeout        int
}

func AddFileMigrateFlags(cmd *cobra.Command, flag *FileMigrateFlag) {
	cmd.PersistentFlags().StringVarP(&flag.SourceIDCName, "sc", "", "", "source host idc name")
	cmd.PersistentFlags().StringVarP(&flag.SourceOSType, "so", "", "linux", "source host os type")
	cmd.PersistentFlags().StringVarP(&flag.SourceEncoding, "se", "", "utf-8", "source host encoding")
	cmd.PersistentFlags().IntVarP(&flag.SourcePort, "sp", "", 22, "source host port")
	cmd.PersistentFlags().StringVarP(&flag.TargetIDCName, "tc", "", "", "target host idc name")
	cmd.PersistentFlags().StringVarP(&flag.TargetOSType, "to", "", "linux", "target host os type")
	cmd.PersistentFlags().StringVarP(&flag.TargetEncoding, "te", "", "utf-8", "target host encoding")
	cmd.PersistentFlags().IntVarP(&flag.TargetPort, "tp", "", 22, "target host port")
}
