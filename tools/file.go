package tools

import (
	"errors"
	"io/ioutil"
	"os"
)

func FileRead(path string) (data string, err error) {
	if path == "" {
		err = errors.New("路径不能为空")
		return
	}

	file, err := os.Open(path)
	if err != nil {
		err = errors.New("打开文件失败:" + err.Error())
		return
	}
	defer func() {
		_ = file.Close()
	}()

	dataBytes, err1 := ioutil.ReadAll(file)
	if err1 != nil {
		err = errors.New("读取文件失败:" + err.Error())
		return
	}
	data = string(dataBytes)

	return
}
