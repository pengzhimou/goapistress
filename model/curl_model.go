// Package model 数据模型
package model

import (
	"encoding/json"
	"strings"

	"goapistress/tools"
)

// CURL curl参数解析
type CURL struct {
	Data map[string][]string
}

// getDataValue 获取数据
func (c *CURL) getDataValue(keys []string) []string {
	var (
		value = make([]string, 0)
	)
	for _, key := range keys {
		var (
			ok bool
		)
		value, ok = c.Data[key]
		if ok {
			break
		}
	}
	return value

	// 不用下面的方式，1 value只创建一次，2 值一定为slice，外面不用判断
	// for _, key := range keys {
	// 	value, ok := c.Data[key]
	// 	if ok {
	// 		return value
	// 	}
	// }
	// return nil

}

// 不能处理选项 curl --location --request GET 'http://xxxxx'
func parseCURL(data string) map[string][]string {
	var (
		urlMap = make(map[string][]string)
	)

	for len(data) > 0 {
		if strings.HasPrefix(data, "curl") {
			data = data[5:]
		}
		data = strings.TrimSpace(data)
		var (
			key   string
			value string
		)
		index := strings.Index(data, " ")
		if index <= 0 {
			break
		}
		key = strings.TrimSpace(data[:index])
		data = data[index+1:]
		data = strings.TrimSpace(data)
		// url
		if !strings.HasPrefix(key, "-") {
			key = strings.Trim(key, "'")
			urlMap["curl"] = []string{key}
			// 去除首尾空格
			data = strings.TrimFunc(data, func(r rune) bool {
				if r == ' ' || r == '\\' || r == '\n' {
					return true
				}
				return false
			})
			continue
		}
		if strings.HasPrefix(data, "-") {
			continue
		}
		var (
			endSymbol = " "
		)
		if strings.HasPrefix(data, "'") {
			endSymbol = "'"
			data = data[1:]
		}
		index = strings.Index(data, endSymbol)
		if index <= -1 {
			index = len(data)
			// break
		}
		value = data[:index]
		if len(data) >= index+1 {
			data = data[index+1:]
		} else {
			data = ""
		}
		// 去除首尾空格
		data = strings.TrimFunc(data, func(r rune) bool {
			if r == ' ' || r == '\\' || r == '\n' {
				return true
			}
			return false
		})
		if key == "" {
			continue
		}
		urlMap[key] = append(urlMap[key], value)
	}

	return urlMap
}

// ParseTheFile 从文件中解析curl
func ParseTheFile(path string) (curl *CURL, err error) {
	curl = &CURL{
		Data: make(map[string][]string),
	}

	data, err := tools.FileRead(path)
	if err != nil {
		return curl, err
	}

	curl.Data = parseCURL(data)
	return
}

// String string
func (c *CURL) String() (url string) {
	curlByte, _ := json.Marshal(c)
	return string(curlByte)
}

// GetURL 获取url
func (c *CURL) GetURL() (url string) {
	keys := []string{"curl", "--url"}
	value := c.getDataValue(keys)
	if len(value) <= 0 {
		return
	}
	url = value[0]
	return
}

// GetMethod 获取 请求方式
func (c *CURL) GetMethod() (method string) {
	keys := []string{"-X", "--request"}
	value := c.getDataValue(keys)
	if len(value) <= 0 {
		return c.defaultMethod()
	}
	method = strings.ToUpper(value[0])
	if tools.InArrayStr(method, []string{"GET", "POST", "PUT", "PATCH", "DELETE"}) {
		return method
	}
	return c.defaultMethod()
}

// defaultMethod 获取默认方法
func (c *CURL) defaultMethod() (method string) {
	method = "GET"
	body := c.GetBody()
	if len(body) > 0 {
		return "POST"
	}
	return
}

// GetHeaders 获取请求头
func (c *CURL) GetHeaders() (headers map[string]string) {
	headers = make(map[string]string, 0)
	keys := []string{"-H", "--header"}
	value := c.getDataValue(keys)
	for _, v := range value {
		tools.GetHeaderValue(v, headers)
	}
	return
}

// GetHeadersStr 获取请求头string
func (c *CURL) GetHeadersStr() string {
	headers := c.GetHeaders()
	bytes, _ := json.Marshal(&headers)
	return string(bytes)
}

// GetBody 获取body
func (c *CURL) GetBody() (body string) {
	keys := []string{"--data", "-d", "--data-urlencode", "--data-raw", "--data-binary"}
	value := c.getDataValue(keys)
	if len(value) <= 0 {
		body = c.getPostForm()
		return
	}
	body = value[0]
	return
}

// getPostForm get post form
func (c *CURL) getPostForm() (body string) {
	keys := []string{"--form", "-F", "--form-string"}
	value := c.getDataValue(keys)
	if len(value) <= 0 {
		return
	}
	body = strings.Join(value, "&")
	return
}
