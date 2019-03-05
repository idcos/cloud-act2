//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package test

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"idcos.io/cloud-act2/model"
)

func NewSqlmock() (sqlmock.Sqlmock, error, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err, nil
	}

	close := func() {
		db.Close()
	}

	gormDb, err := gorm.Open("mysql", db)
	if err != nil {
		close()
		return nil, err, nil
	}

	model.SetDb(gormDb)
	return mock, nil, close
}
