// Package verify 校验
package verify

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"goapistress/model"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

// GetBodyData 处理gzip压缩
func GetStatusCodeAndBody(reqForm *model.RequestForm, response *http.Response) (code int, body []byte, err error) {
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		defer func() {
			_ = reader.Close()
		}()
	default:
		reader = response.Body
	}
	body, err = ioutil.ReadAll(reader)
	if err != nil {
		fmt.Printf("请求结果 ioutil.ReadAll err:%v", err)
	}

	// 开启调试模式
	if reqForm.GetDebug() {
		fmt.Printf("请求结果 httpCode:%d body:%s err:%v \n", response.StatusCode, string(body), err)
	}

	response.Body = ioutil.NopCloser(bytes.NewReader(body))
	code = response.StatusCode
	return
}

// HTTPStatusCode 通过 HTTP 状态码判断是否请求成功
func HTTPStatusCode(reqForm *model.RequestForm, respCode int, respBody []byte) (code int, isSucceed bool) {
	exptCode, err := strconv.Atoi(reqForm.Verify["statusCode"])
	if err != nil {
		panic("没有期待的statusCode," + err.Error())
	}
	if respCode == exptCode {
		isSucceed = true
	}
	code = respCode
	return
}

/***************************  返回值为json  ********************************/

// ResponseJSON 返回数据结构体
type ResponseJSON struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// HTTPJson  通过返回的Body 判断
// 返回示例: {"code":200,"msg":"Success","data":{}}
func HTTPJson(reqForm *model.RequestForm, respCode int, respBody []byte) (code int, isSucceed bool) {
	// statuscode无用
	_ = respCode

	// body检查
	responseJSON := &ResponseJSON{}
	err := json.Unmarshal(respBody, responseJSON)
	if err != nil {
		fmt.Printf("请求结果 json.Unmarshal err:%v", err)
	} else {
		jsonCode := responseJSON.Code

		exptResp := ResponseJSON{}
		json.Unmarshal([]byte(reqForm.Verify["json"]), &exptResp)

		isSucceed = true
		if responseJSON.Code != exptResp.Code {
			isSucceed = false
		}
		if (exptResp.Msg != "" && exptResp.Msg != responseJSON.Msg) ||
			exptResp.Message != "" && exptResp.Message != responseJSON.Message {
			isSucceed = false
		}

		code = jsonCode
	}

	return
}
