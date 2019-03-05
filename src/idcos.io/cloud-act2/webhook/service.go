//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package webhook

import (
	"encoding/json"
	"fmt"
	"idcos.io/cloud-act2/utils/generator"
	"idcos.io/cloud-act2/utils/stringutil"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"idcos.io/cloud-act2/define"
	"idcos.io/cloud-act2/model"
	"idcos.io/cloud-act2/service/common"
	"idcos.io/cloud-act2/redis"
)

//AddWebHook 添加web钩子
func AddWebHook(form common.WebHookAddForm) (err error) {
	logger := getLogger()

	if !stringutil.StringScliceContains(define.WebHookEventList, form.Event) {
		return errors.New("illegal event")
	}

	headersJSON, dataJSON, err := getCustomJSON(form.CustomHeaders, form.CustomData)
	if err != nil {
		return err
	}

	webHook := model.Act2WebHook{
		ID:            generator.GenUUID(),
		Event:         form.Event,
		Uri:           form.Uri,
		CustomHeaders: headersJSON,
		CustomData:    dataJSON,
		Token:         form.Token,
		AddTime:       time.Now(),
		UpdateTime:    time.Now(),
	}

	tx := model.GetDb().Begin()
	err = tx.Save(&webHook).Error
	if err != nil {
		logger.Error("save web hook fail", "error", err)
		tx.Rollback()
		return err
	}

	//删除redis中该event的数据，让web hook通知时从库中重新读取
	err = deleteRedis(form.Event)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		logger.Error("db commit fail", "error", err)
		return err
	}

	return nil
}

//UpdateWebHook 修改web钩子
func UpdateWebHook(form common.WebHookUpdateForm) (err error) {
	logger := getLogger()

	oldWebHook, err := model.FindWebHookByID(form.ID)
	if gorm.IsRecordNotFoundError(err) {
		return fmt.Errorf("%s not found", form.ID)
	}
	if err != nil {
		logger.Error("find web hook fail", "id", form.ID, "error", err)
		return err
	}

	if !stringutil.StringScliceContains(define.WebHookEventList, form.Event) {
		return errors.New("illegal event")
	}

	headersJSON, dataJSON, err := getCustomJSON(form.CustomHeaders, form.CustomData)
	if err != nil {
		return err
	}

	oldEvent := oldWebHook.Event
	oldWebHook.Event = form.Event
	oldWebHook.CustomData = dataJSON
	oldWebHook.CustomHeaders = headersJSON
	oldWebHook.Token = form.Token
	oldWebHook.Uri = form.Uri
	oldWebHook.UpdateTime = time.Now()

	tx := model.GetDb().Begin()
	err = tx.Save(&oldWebHook).Error
	if err != nil {
		logger.Error("save web hook fail", "error", err)
		tx.Rollback()
		return err
	}

	if oldEvent == form.Event {
		err = deleteRedis(form.Event)
	} else {
		err = deleteRedis(form.Event)
		err = deleteRedis(oldEvent)
	}
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		logger.Error("db commit fail", "error", err)
		return err
	}

	return nil
}

//DeleteByID
func DeleteByID(id string) (err error) {
	logger := getLogger()

	webHook, err := model.FindWebHookByID(id)
	if gorm.IsRecordNotFoundError(err) {
		return nil
	}
	if err != nil {
		logger.Error("find web hook fail", "id", id, "error", err)
		return err
	}

	tx := model.GetDb().Begin()
	err = tx.Where("id = ?", id).Delete(&model.Act2WebHook{}).Error
	if err != nil {
		logger.Error("save web hook fail", "error", err)
		tx.Rollback()
		return err
	}

	err = deleteRedis(webHook.Event)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		logger.Error("db commit fail", "error", err)
		return err
	}

	return nil
}

func getCustomJSON(customHeaders map[string]string, customData interface{}) (headersJSON, dataJSON string, err error) {
	logger := getLogger()

	headersJSON = ""
	if len(customHeaders) > 0 {
		bytes, err := json.Marshal(customHeaders)
		if err != nil {
			logger.Error("custom headers to json fail", "headers", customHeaders, "error", err)
			return "", "", err
		}

		headersJSON = string(bytes)
	}

	dataJSON = ""
	if customData != nil {
		bytes, err := json.Marshal(customData)
		if err != nil {
			logger.Error("custom data to json fail", "data", customData, "error", err)
			return "", "", err
		}

		dataJSON = string(bytes)
	}

	return headersJSON, dataJSON, nil
}

func deleteRedis(event string) (err error) {
	logger := getLogger()

	err = redis.GetRedisClient().Del(define.RedisWebHookEventPre + event).Err()
	if err != nil {
		logger.Error("delete redis key fail", "event", event)
		return err
	}

	return nil
}
