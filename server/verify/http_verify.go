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

// getZipData 处理gzip压缩
func getZipData(response *http.Response) (body []byte, err error) {
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
	response.Body = ioutil.NopCloser(bytes.NewReader(body))
	return
}

// HTTPStatusCode 通过 HTTP 状态码判断是否请求成功
func HTTPStatusCode(reqForm *model.RequestForm, response *http.Response) (respCode int, isSucceed bool) {
	// defer func() {
	// 	_ = response.Body.Close()
	// }()
	respCode = response.StatusCode
	exptCode, err := strconv.Atoi(reqForm.Verify["statusCode"])
	if err != nil {
		panic("没有期待的statusCode," + err.Error())
	}
	if respCode == exptCode {
		isSucceed = true
	}
	// 开启调试模式
	if reqForm.GetDebug() {
		body, err := getZipData(response)
		fmt.Printf("请求结果 httpCode:%d body:%s err:%v \n", response.StatusCode, string(body), err)
	}
	io.Copy(ioutil.Discard, response.Body)
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
// code 默认将http code作为返回码，http code 为200时 取body中的返回code
// func HTTPJson(request *model.RequestForm, response *http.Response) (code int, isSucceed bool) {
// 	defer func() {
// 		_ = response.Body.Close()
// 	}()
// 	code = response.StatusCode
// 	if code == http.StatusOK {
// 		body, err := getZipData(response)
// 		if err != nil {
// 			code = model.ParseError
// 			fmt.Printf("请求结果 ioutil.ReadAll err:%v", err)
// 		} else {
// 			responseJSON := &ResponseJSON{}
// 			err = json.Unmarshal(body, responseJSON)
// 			if err != nil {
// 				code = model.ParseError
// 				fmt.Printf("请求结果 json.Unmarshal err:%v", err)
// 			} else {
// 				code = responseJSON.Code
// 				// body 中code返回200为返回数据成功
// 				if responseJSON.Code == request.StatusCode {
// 					isSucceed = true
// 				}
// 			}
// 		}
// 		// 开启调试模式
// 		if request.GetDebug() {
// 			fmt.Printf("请求结果 httpCode:%d body:%s err:%v \n", response.StatusCode, string(body), err)
// 		}
// 	}
// 	io.Copy(ioutil.Discard, response.Body)
// 	return
// }

func HTTPJson(reqForm *model.RequestForm, response *http.Response) (respCode int, isSucceed bool) {
	// defer func() {
	// 	_ = response.Body.Close()
	// }()

	// body检查
	body, err := getZipData(response)
	if err != nil {
		respCode = model.ParseError
		fmt.Printf("请求结果 ioutil.ReadAll err:%v", err)
	} else {
		responseJSON := &ResponseJSON{}
		err = json.Unmarshal(body, responseJSON)
		if err != nil {
			respCode = model.ParseError
			fmt.Printf("请求结果 json.Unmarshal err:%v", err)
		} else {
			respCode = responseJSON.Code
			// body 中code返回200为返回数据成功

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
		}
	}
	// 开启调试模式
	if reqForm.GetDebug() {
		fmt.Printf("请求结果 httpCode:%d body:%s err:%v \n", response.StatusCode, string(body), err)
	}

	io.Copy(ioutil.Discard, response.Body)
	return
}
