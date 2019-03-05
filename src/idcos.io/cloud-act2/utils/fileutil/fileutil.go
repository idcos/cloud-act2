//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package fileutil

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"idcos.io/cloud-act2/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/satori/go.uuid"
)

const (
	octStreamHead = "data:application/octet-stream;base64,"
)

//SaveScriptToFileWithSaltSys 保存脚本到文件
func SaveScriptToFileWithSaltSys(fileFolder string, script []byte, scriptType string) (fileName string, err error) {
	suffix := GetFileSuffix(scriptType)
	fileName = uuid.NewV4().String() + suffix
	filePath := fileFolder + "/" + fileName

	err = ioutil.WriteFile(filePath, script, 0644)
	return
}

//RemoveFile 根据路径和文件名删除文件
func RemoveFile(fileFolder, fileName string) (err error) {
	filePath := fileFolder + "/" + fileName

	err = os.Remove(filePath)
	return
}

//GetFileSuffix 根据脚本类型生成文件后缀
func GetFileSuffix(scriptType string) (suffix string) {
	scriptType = strings.TrimSpace(strings.ToLower(scriptType))
	switch scriptType {
	case "python":
		suffix = ".py"
	case "bat":
		suffix = ".bat"
	case "sls":
		suffix = ".sls"
	default:
		suffix = ".sh"
	}
	return
}

//RemoveFileSuffix 删除文件名的后缀
func RemoveFileSuffix(fileName string) (preFileName string) {
	index := strings.LastIndex(fileName, ".")
	preFileName = fileName[:index]
	return
}

/**
获取文件内容
*/
func ReadFile(path string) (string, error) {
	logger := getLogger()

	bytes, err := ReadFileToBytes(path)

	if err != nil {
		logger.Error("read file error", "error", err)
		return "", err
	}
	return string(bytes), err
}

/**

 */
func ReadFileToBytes(path string) ([]byte, error) {
	logger := getLogger()

	path, err := ExpandUser(path)
	if err != nil {
		logger.Error("could not expand user path", "error", err)
		return nil, err
	}

	return ioutil.ReadFile(path)

}

//WriteFile 保存文件
func WriteFile(path string, data []byte) (err error) {
	logger := getLogger()

	logger.Info("write file", "path", path, "length", len(data))
	err = ioutil.WriteFile(path, data, 0777)
	if err != nil {
		logger.Error("write file error", "error", err)
	}

	return
}

func FileExists(filepath string) (bool, error) {
	filepath, err := ExpandUser(filepath)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, err
	}
}

// GetDataURI 获取数据
func GetDataURI(data []byte) ([]byte, error) {
	if bytes.HasPrefix(data, []byte(octStreamHead)) {
		data = data[len(octStreamHead):]
		dest := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
		n, err := base64.StdEncoding.Decode(dest, data)
		if err != nil {
			return nil, err
		}
		return dest[:n], nil
	}
	return data, nil
}

//OpenAndCreate 打开文件，若不存在则新建
func OpenAndCreate(filePath string) (file *os.File, err error) {
	file, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0766)
	return
}

//OpenAndCreateOrCover 打开文件，若不存在则新建，若存在则覆盖
func OpenAndCreateOrCover(filePath string) (file *os.File, err error) {
	file, err = os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0766)
	return
}

//GetFilesByPath 获取文件夹下所有文件的内容
func GetFilesByPath(filePath string) (data [][]byte, err error) {
	rd, err := ioutil.ReadDir(filePath)
	if err != nil {
		return
	}

	data = make([][]byte, 0, 10)
	for _, fileInfo := range rd {
		bys, err := ReadFileToBytes(fmt.Sprintf("%s/%s", filePath, fileInfo.Name()))
		if err != nil {
			break
		}
		data = append(data, bys)
	}

	return
}

//CreateDirIfNotExist 创建文件夹如果它不存在
func CreateDirIfNotExist(dir string) (err error) {
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
	}
	return
}

//RemoveFilesByDir 删除文件夹下所有文件
func RemoveFilesByDir(dir string) (err error) {
	if strings.Index(dir, config.Conf.ProjectPath) != 0 {
		return
	}

	rd, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fileInfo := range rd {
		RemoveFile(dir, fileInfo.Name())
	}
	return
}

func ExpandUser(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := Home()
	if err != nil {
		return "", err
	}

	return home + path[1:], nil
}

func Home() (string, error) {
	home := ""

	switch runtime.GOOS {
	case "windows":
		home = filepath.Join(os.Getenv("HomeDrive"), os.Getenv("HomePath"))
		if home == "" {
			home = os.Getenv("UserProfile")
		}

	default:
		home = os.Getenv("HOME")
	}

	if home == "" {
		return "", errors.New("no home found")
	}
	return home, nil
}
