// Package client http 客户端
package client

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"goapistress/model"
	httplongclinet "goapistress/server/client/http_longclient"
	"goapistress/tools"

	"golang.org/x/net/http2"
)

// logErr err
var logErr = log.New(os.Stderr, "", 0)

// HTTPRequest HTTP 请求
// method 方法 GET POST
// url 请求的url
// body 请求的body
// headers 请求头信息
// timeout 请求超时时间
func HTTPRequest(chanID uint64, reqForm *model.RequestForm) (resp *http.Response, requestTime uint64, err error) {
	method := reqForm.Method
	url := reqForm.URL
	body := reqForm.GetBody()
	timeout := reqForm.ClientTimeout
	headers := reqForm.Headers

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}

	// 在req中设置Host，解决在header中设置Host不生效问题
	if _, ok := headers["Host"]; ok {
		req.Host = headers["Host"]
	}
	// 设置默认为utf-8编码
	if _, ok := headers["Content-Type"]; !ok {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	var client *http.Client
	if reqForm.Keepalive { // 长链接 client
		client = httplongclinet.NewClient(chanID, reqForm)
		startTime := time.Now()
		resp, err = client.Do(req)
		requestTime = uint64(tools.DiffNano(startTime))
		if err != nil {
			logErr.Println("请求失败:", err)

			return
		}
		return
	} else { // 短链接 client
		req.Close = true
		tr := &http.Transport{}
		if reqForm.HTTP2 {
			// 使用真实证书 验证证书 没必要
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{
					// InsecureSkipVerify: false,
					InsecureSkipVerify: true,
					Certificates:       []tls.Certificate{*reqForm.TLSCertificate},
				},
			}
			if err = http2.ConfigureTransport(tr); err != nil {
				return
			}
		} else {
			// 跳过证书验证
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					Certificates:       []tls.Certificate{*reqForm.TLSCertificate},
				},
			}
		}

		client = &http.Client{
			Transport: tr,
			Timeout:   timeout,
		}
	}

	startTime := time.Now()
	resp, err = client.Do(req)
	requestTime = uint64(tools.DiffNano(startTime))
	if err != nil {
		logErr.Println("请求失败:", err)

		return
	}
	return
}
