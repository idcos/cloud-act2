//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package model

import (
	"fmt"
	"idcos.io/cloud-act2/config"
	"testing"
	"time"

	"github.com/hashicorp/go-uuid"
)

func loadConfig() {
	config.LoadConfig("/usr/yunji/cloud-act2/etc/cloud-act2.yaml")
	OpenConn(config.Conf)
}

func TestAct2JobRecord_Save(t *testing.T) {
	loadConfig()

	//ID, _ := uuid.GenerateUUID()

	record := JobRecord{
		ID:            "41ecc7c4-44f0-4f4e-a291-1a75fd7a4173",
		StartTime:     time.Now(),
		EndTime:       time.Now(),
		ExecuteStatus: "DOING",
		ResultStatus:  "SUCCESS",
		Hosts:         "Hosts",
		Provider:      "salt",
		Script:        "script",
		ScriptType:    "shell",
		Timeout:       100,
		Parameters:    "Parameters",
	}

	err := record.Save()
	if err != nil {
		t.Fatal(err)
	}
}

func TestJobRecord_Update(t *testing.T) {
	loadConfig()

	var jobrecord JobRecord

	jobrecord.GetByID("41ecc7c4-44f0-4f4e-a291-1a75fd7a4172")

	jobrecord.ScriptType = "python"

	jobrecord.Save()

}

func TestJobRecord_Update2(t *testing.T) {
	loadConfig()

	jobrecord := JobRecord{
		ID: "41ecc7c4-44f0-4f4e-a291-1a75fd7a4172",
	}

	jobrecord.Update("SCRIPT_TYPE", "hello")
}

func TestJobRecord_DeleteByID(t *testing.T) {
	loadConfig()

	record := JobRecord{
		ID: "41ecc7c4-44f0-4f4e-a291-1a75fd7a4172",
	}
	record.DeleteByID()
}

func TestFindById(t *testing.T) {
	loadConfig()

	var jobrecord JobRecord

	jobrecord.GetByID("41ecc7c4-44f0-4f4e-a291-1a75fd7a4172")

	fmt.Printf("%v", jobrecord)
}

func TestJobRecord_GetAll(t *testing.T) {
	loadConfig()
	var records []*JobRecord

	globalDb.Find(&records)

	for _, record := range records {
		fmt.Printf("%v\n", record)
	}

}

func TestAct2JobRecord_DeleteByID(t *testing.T) {
	loadConfig()
	ID, _ := uuid.GenerateUUID()

	result := HostResult{
		ID:            ID,
		StartTime:     time.Now(),
		EndTime:       time.Now(),
		ExecuteStatus: "DOING",
		ResultStatus:  "SUCCESS",
		HostIP:        "10.0.0.12",
		Stdout:        "hello world",
		Stderr:        "",
	}

	globalDb.Create(&result)
}
