package httplongclinet

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"goapistress/model"

	"golang.org/x/net/http2"
)

var (
	mutex   sync.RWMutex
	clients = make(map[uint64]*http.Client, 0)
)

// NewClient new
func NewClient(i uint64, reqForm *model.RequestForm) *http.Client {
	client := getClient(i)
	if client != nil {
		return client
	}
	return setClient(i, reqForm)
}

func getClient(i uint64) *http.Client {
	mutex.RLock()
	defer mutex.RUnlock()
	return clients[i]
}

func setClient(i uint64, reqForm *model.RequestForm) *http.Client {
	mutex.Lock()
	defer mutex.Unlock()
	client := createLongHttpClient(reqForm)
	clients[i] = client
	return client
}

// createLangHttpClient 初始化长连接客户端参数
func createLongHttpClient(reqForm *model.RequestForm) *http.Client {
	tr := &http.Transport{}
	if reqForm.HTTP2 {
		// 使用真实证书 验证证书 模拟真实请求
		tr = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        0,                // 最大连接数,默认0无穷大
			MaxIdleConnsPerHost: reqForm.MaxCon,   // 对每个host的最大连接数量(MaxIdleConnsPerHost<=MaxIdleConns)
			IdleConnTimeout:     90 * time.Second, // 多长时间未使用自动关闭连接
			TLSClientConfig: &tls.Config{
				// InsecureSkipVerify: false,
				InsecureSkipVerify: true,
				Certificates:       []tls.Certificate{*reqForm.TLSCertificate},
			},
		}
		_ = http2.ConfigureTransport(tr)
	} else {
		// 跳过证书验证
		tr = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        0,                // 最大连接数,默认0无穷大
			MaxIdleConnsPerHost: reqForm.MaxCon,   // 对每个host的最大连接数量(MaxIdleConnsPerHost<=MaxIdleConns)
			IdleConnTimeout:     90 * time.Second, // 多长时间未使用自动关闭连接
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				Certificates:       []tls.Certificate{*reqForm.TLSCertificate},
			},
		}
	}
	return &http.Client{
		Transport: tr,
	}
}
