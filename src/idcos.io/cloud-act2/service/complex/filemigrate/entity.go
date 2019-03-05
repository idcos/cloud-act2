//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package filemigrate

import "idcos.io/cloud-act2/service/complex/common"

type MigrateInfo struct {
	SourceHost     common.ComplexHost `json:"sourceHost"`
	TargetHost     common.ComplexHost `json:"targetHost"`
	SourceFilePath string             `json:"sourceFilePath"`
	TargetFilePath string             `json:"targetFilePath"`
	Timeout        int                `json:"timeout"`
	MasterTransfer bool               `json:"masterTransfer"`
}

type MasterMigrateInfo struct {
	SourceHost     common.ComplexHost `json:"sourceHost"`
	TargetHost     common.ComplexHost `json:"targetHost"`
	SourceFilePath string             `json:"sourceFilePath"`
	TargetFilePath string             `json:"targetFilePath"`
	Timeout        int                `json:"timeout"`
}
