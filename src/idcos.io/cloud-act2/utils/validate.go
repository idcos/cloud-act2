//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package utils

import (
	"gopkg.in/go-playground/validator.v9"
	"sync"
)

var once sync.Once
var Validate *validator.Validate

func InitValidate() {
	once.Do(func() {
		Validate = validator.New()
	})
}
