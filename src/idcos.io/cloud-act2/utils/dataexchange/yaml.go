//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package dataexchange

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

func LoadYaml(filename string, v interface{}) error {
	byts, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(byts, v)
}

func SaveYaml(filename string, v interface{}) error {
	byts, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, byts, 0640)
}
