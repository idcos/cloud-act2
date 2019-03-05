//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package arrays

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func ArrayIsEmpty(name string, v []interface{}) error {
	if v == nil || len(v) == 0 {
		return errors.New(fmt.Sprintf("the %s value is empty", name))
	}
	return nil
}

func ArrayTrimSpace(values []string) []string {
	var r []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		r = append(r, value)
	}
	return r
}

func SkipOnlySpaceString(values []string) []string {
	var r []string

	for _, value := range values {
		temp := strings.TrimSpace(value)
		if len(temp) != 0 {
			r = append(r, value)
		}
	}
	return r
}

// RemoveDuplicateKey remove duplicate key
func RemoveDuplicateKey(values []string) []string {
	v := map[string]bool{}
	for _, value := range values {
		if _, ok := v[value]; !ok {
			v[value] = true
		}
	}

	var r []string
	for key := range v {
		r = append(r, key)
	}
	return r
}

/**
  var a = map[int]string{1: "1", 2: "2"}
  fmt.Println(Contain(3, a))
*/
func Contain(obj interface{}, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}
	return false
}

func JoinBytes(bytes []byte, sep string) string {
	strs := make([]string, len(bytes))
	for i, byt := range bytes {
		strs[i] = fmt.Sprintf("%d", byt)
	}
	return strings.Join(strs, sep)
}

func SpliteBytesStr(bytesText string, sep string) ([]byte, error) {
	if len(bytesText) == 0 {
		return []byte{}, nil
	}

	byteStrList := strings.Split(bytesText, sep)
	bytes := make([]byte, len(byteStrList))
	for i, str := range byteStrList {
		number, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		bytes[i] = byte(number)
	}

	return bytes, nil
}
