// Package golink 连接
package golink

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"goapistress/model"
	"goapistress/server/client"
	"goapistress/server/verify"
)

// HTTP 请求
func HTTP(ctx context.Context, chanID uint64, chanResults chan<- *model.RequestResults, totalNumber uint64, wg *sync.WaitGroup, reqForm *model.RequestForm) {
	defer func() {
		wg.Done()
	}()
	// fmt.Printf("启动协程 编号:%05d \n", chanID)
	for i := uint64(0); i < totalNumber; i++ {
		if ctx.Err() != nil {
			fmt.Printf("ctx.Err err: %v \n", ctx.Err())
			break
		}

		listRF := getRequestList(reqForm)
		isSucceed, errCode, requestTime, contentLength := sendList(chanID, listRF)
		requestResults := &model.RequestResults{
			Time:          requestTime,
			IsSucceed:     isSucceed,
			RtnCode:       errCode,
			ReceivedBytes: contentLength,
		}
		requestResults.SetID(chanID, i)
		chanResults <- requestResults
	}
}

// sendList 多个接口分步压测
func sendList(chanID uint64, listRF []*model.RequestForm) (isSucceed bool, errCode map[string]int, requestTime uint64, contentLength int64) {
	errCode = map[string]int{"statusCode": model.HTTPOk}
	for _, rF := range listRF {
		succeed, code, u, length := send(chanID, rF)
		isSucceed = succeed
		errCode = code
		requestTime = requestTime + u
		contentLength = contentLength + length
		if !succeed {
			break
		}
	}
	return
}

// send 发送一次请求
func send(chanID uint64, reqForm *model.RequestForm) (bool, map[string]int, uint64, int64) {
	var (
		isSucceed     = false
		rtnCode       = make(map[string]int)
		contentLength = int64(0)
		err           error
		resp          *http.Response
		requestTime   uint64
	)

	newReqForm := getRequest(reqForm)

	resp, requestTime, err = client.HTTPRequest(chanID, newReqForm)

	if err != nil {
		rtnCode = map[string]int{"reqerr": model.RequestErr} // 请求错误
	} else {
		defer func() {
			_ = resp.Body.Close()
		}()
		// 此处原方式获取的数据长度可能是 -1，换成如下方式获取可获取到正确的长度
		contentLength, err = getBodyLength(resp)
		if err != nil {
			contentLength = resp.ContentLength
		}

		// 验证请求结果
		respCode, respBodyGetBodyData, err := verify.GetStatusCodeAndBody(reqForm, resp)
		if err != nil {
			rtnCode, isSucceed = map[string]int{"reqerr": model.RequestErr}, false
			fmt.Println("从GetStatusCodeAndBody取得statuscode或body有错误")
			return isSucceed, rtnCode, requestTime, contentLength
		}

		tempmap := map[string]int{}
		tempisSucceed := true
		for _, f := range newReqForm.GetVerifyHTTP() {
			rtnCode, isSucceed = f(newReqForm, respCode, respBodyGetBodyData)
			for k, v := range rtnCode {
				tempmap[k] = v
			}
			if !isSucceed {
				tempisSucceed = false
				// return isSucceed, rtnCode, requestTime, contentLength
			}
		}
		isSucceed = tempisSucceed
		rtnCode = tempmap
	}
	return isSucceed, rtnCode, requestTime, contentLength
}

// getBodyLength 获取响应数据长度
func getBodyLength(response *http.Response) (length int64, err error) {
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
	body, err := ioutil.ReadAll(reader)
	response.Body = ioutil.NopCloser(bytes.NewReader(body))
	return int64(len(body)), err
}
