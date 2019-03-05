//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package maps

import (
	"errors"
	"fmt"
)

func MapSafeToMap(mapData map[interface{}]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range mapData {
		keyValue := fmt.Sprintf("%v", key)
		if v, ok := value.(map[interface{}]interface{}); ok {
			mapValue := MapSafeToMap(v)
			result[keyValue] = mapValue
		} else {
			strValue := fmt.Sprintf("%v", value)
			result[keyValue] = strValue
		}
	}

	return result
}

func MapIsEmpty(name string, v map[string]interface{}) error {
	if v == nil || len(v) == 0 {
		return errors.New(fmt.Sprintf("the %s value is empty", name))
	}
	return nil
}