//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package model

import (
	"fmt"
	"time"

	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"idcos.io/cloud-act2/config"
	act2Log "idcos.io/cloud-act2/log"
)

var (
	globalDb *gorm.DB
	logChan  = make(chan struct{})
)

//如果系统爆出：this user requires mysql native password authentication
//需要在my.conf文件中的[mysqld]下添加
//default-authentication-plugin = mysql_native_password
func OpenConn(config *config.Config) error {

	logger := act2Log.L()

	dbConfig := config.Db
	mysqlURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local&timeout=5s&readTimeout=5m&writeTimeout=5m", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name, dbConfig.Encoding)

	db, err := gorm.Open("mysql", mysqlURL)

	if err != nil {
		logger.Debug(fmt.Sprintf("fail to connect db url:%s, error %v ", mysqlURL, err))
		return err
	}

	poolSize := 64
	if config.Db.PoolSize != 0 {
		poolSize = config.Db.PoolSize
	}
	pollIdleSize := 64
	if config.Db.PoolIdleSize != 0 {
		pollIdleSize = config.Db.PoolIdleSize
	}
	connMaxLifeTime := 60
	if config.Db.ConnMaxLifeLime != 0 {
		connMaxLifeTime = config.Db.ConnMaxLifeLime
	}
	db.DB().SetMaxOpenConns(poolSize)
	db.DB().SetMaxIdleConns(pollIdleSize)
	db.DB().SetConnMaxLifetime(time.Duration(connMaxLifeTime) * time.Second)

	resetLogger := func() {
		gormLogger := gorm.Logger{LogWriter: log.New(act2Log.Writer(), "\r\n", 0)}
		db.SetLogger(gormLogger)
	}
	resetLogger()

	act2Log.AddLogNotify(logChan)
	go func() {
		// 在日志重定向后，需要重新设定logger
		for range logChan {
			resetLogger()
		}
	}()

	globalDb = db
	globalDb.LogMode(config.Db.Debug)

	return nil
}

func GetDb() *gorm.DB {
	return globalDb
}

func CloseDb() {
	logger := act2Log.L()

	if globalDb != nil {
		globalDb.Close()
	} else {
		logger.Warn("fail to close db, db connect not init")
	}
}
