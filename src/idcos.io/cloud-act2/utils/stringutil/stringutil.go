//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package stringutil

import (
	"errors"
	"fmt"
)

func StringIsEmpty(name string, v string) error {
	if v == "" || len(v) == 0 {
		return errors.New(fmt.Sprintf("the %s value is blank", name))
	}
	return nil
}

func StringScliceContains(slices []string, str string) bool {
	isContains := false
	for _, s := range slices {
		if s == str {
			isContains = true
		}
	}
	return isContains
}

func StringScliceContainsIndex(slices []string, str string) int {
	index := -1
	for i, s := range slices {
		if s == str {
			index = i
			break
		}
	}
	return index
}

func DifferenceStringSlice(source []string, target []string) (resultHosts []string) {
	resultHosts = make([]string, 0, 100)
	for _, sourceItem := range source {
		if !StringScliceContains(target, sourceItem) {
			resultHosts = append(resultHosts, sourceItem)
		}
	}
	return
}