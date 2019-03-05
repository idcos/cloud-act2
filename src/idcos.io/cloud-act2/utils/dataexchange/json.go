//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package dataexchange

import "encoding/json"

/**
value to json string
*/
func ToJsonString(v interface{}) string {
	logger := getLogger()

	bytes, err := json.Marshal(v)
	if err != nil {
		logger.Error("fail to marshal value")
	}
	return string(bytes)
}

func ToPrettyJsonString(v interface{}) string {
	logger := getLogger()

	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logger.Error("fail to marshal value")
	}
	return string(bytes)
}