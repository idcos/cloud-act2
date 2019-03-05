//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package ints

import (
	"fmt"
	"strconv"
)

//GetInt from interface
func GetInt(i interface{}) (int, error) {
	var v int
	switch i.(type) {
	case float64:
		v = int(i.(float64))
	case int64:
		v = int(i.(int64))
	case int:
		v = i.(int)
	default:
		s := fmt.Sprintf("%v", i)
		d, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		} else {
			v = d
		}
	}
	return v, nil
}
