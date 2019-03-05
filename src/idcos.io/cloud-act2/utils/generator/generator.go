//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package generator

import "github.com/hashicorp/go-uuid"

func GenUUID() string {
	logger := getLogger()

	id, err := uuid.GenerateUUID()
	if err != nil {
		logger.Info("gen id error" + err.Error())
	}
	return id
}
