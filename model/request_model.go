// Package model 请求数据模型package model
package model

import (
	"errors"
	"fmt"
	"goapistress/tools"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// 返回 code 码
const (
	// HTTPOk 请求成功
	HTTPOk = 200
	// RequestErr 请求错误
	RequestErr = 509
	// ParseError 解析错误
	ParseError = 510 // 解析错误
)

// 支持协议
const (
	// MPTypeHTTP http 协议
	MPTypeHTTP = "http"
	// MPTypeWebSocket webSocket 协议
	MPTypeWebSocket = "webSocket"
	// MPTypeGRPC grpc 协议
	MPTypeGRPC   = "grpc"
	MPTypeRadius = "radius"
)

// 校验函数
var (
	// verifyMapHTTP http 校验函数
	verifyMapHTTP = make(map[string]VerifyHTTP)
	// verifyMapHTTPMutex http 并发锁
	verifyMapHTTPMutex sync.RWMutex
	// verifyMapWebSocket webSocket 校验函数
	verifyMapWebSocket = make(map[string]VerifyWebSocket)
	// verifyMapWebSocketMutex webSocket 并发锁
	verifyMapWebSocketMutex sync.RWMutex
)

// RegisterVerifyHTTP 注册 http 校验函数
func RegisterVerifyHTTP(verify string, verifyFunc VerifyHTTP) {
	verifyMapHTTPMutex.Lock()
	defer verifyMapHTTPMutex.Unlock()
	key := fmt.Sprintf("%s.%s", MPTypeHTTP, verify)
	verifyMapHTTP[key] = verifyFunc
}

// RegisterVerifyWebSocket 注册 webSocket 校验函数
func RegisterVerifyWebSocket(verify string, verifyFunc VerifyWebSocket) {
	verifyMapWebSocketMutex.Lock()
	defer verifyMapWebSocketMutex.Unlock()
	key := fmt.Sprintf("%s.%s", MPTypeWebSocket, verify)
	verifyMapWebSocket[key] = verifyFunc
}

// Verify 验证器
type Verify interface {
	GetCode() int    // 有一个方法，返回code为200为成功
	GetResult() bool // 返回是否成功
}

// VerifyHTTP http 验证
type VerifyHTTP func(request *RequestForm, response *http.Response) (code int, isSucceed bool)

// VerifyWebSocket webSocket 验证
type VerifyWebSocket func(request *RequestForm, seq string, msg []byte) (code int, isSucceed bool)

// RequestForm 请求数据
type RequestForm struct {
	URL           string            // URL
	MP            string            // http/webSocket/tcp
	Method        string            // 方法 GET/POST/PUT
	Headers       map[string]string // Headers
	Body          string            // body
	Verify        string            // 验证的方法
	ClientTimeout time.Duration     // 请求超时时间
	Debug         bool              // 是否开启Debug模式
	MaxCon        int               // 每个连接的请求数
	HTTP2         bool              // 是否使用http2.0
	Keepalive     bool              // 是否开启长连接
	Code          int               // 验证的状态码
}

// GetBody 获取请求数据
func (r *RequestForm) GetBody() (body io.Reader) {
	return strings.NewReader(r.Body)
}

// getVerifyKey 获取校验 key
func (r *RequestForm) getVerifyKey() (key string) {
	return fmt.Sprintf("%s.%s", r.MP, r.Verify)
}

// GetVerifyHTTP 获取数据校验方法
func (r *RequestForm) GetVerifyHTTP() VerifyHTTP {
	verify, ok := verifyMapHTTP[r.getVerifyKey()]
	if !ok {
		panic("GetVerifyHTTP 验证方法不存在:" + r.Verify)
	}
	return verify
}

// GetVerifyWebSocket 获取数据校验方法
func (r *RequestForm) GetVerifyWebSocket() VerifyWebSocket {
	verify, ok := verifyMapWebSocket[r.getVerifyKey()]
	if !ok {
		panic("GetVerifyWebSocket 验证方法不存在:" + r.Verify)
	}
	return verify
}

// NewReqForm 生成请求结构体
// url 压测的url
// verify 验证方法 在server/verify中 http 支持:statusCode、json webSocket支持:json
// timeout 请求超时时间
// debug 是否开启debug
// path curl文件路径 http接口压测，自定义参数设置
func NewReqForm(requrl, method, verify string, statusCode int, clientTimeout time.Duration, debug bool, curlFilePath string, reqHeaders []string, reqBody string, maxCon int, http2, keepalive bool) (request *RequestForm, err error) {
	var (
		headers = make(map[string]string)
		body    string
	)

	// 读取基本参数，赋值
	if curlFilePath != "" { // curl文件转换
		var curl *CURL
		curl, err = ParseTheFile(curlFilePath)
		if err != nil {
			return nil, err
		}
		if requrl == "" {
			requrl = curl.GetURL()
		}
		method = curl.GetMethod()
		headers = curl.GetHeaders()
		body = curl.GetBody()
	} else { // 直接入参转换
		body = reqBody
		for _, v := range reqHeaders {
			tools.GetHeaderValue(v, headers)
		}
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
	}

	// 主protocol分类，以及根据分类覆盖需要修改的url
	mainProtocol := ""
	switch {
	case strings.HasPrefix(requrl, "http://") || strings.HasPrefix(requrl, "https://"):
		mainProtocol = MPTypeHTTP
	case strings.HasPrefix(requrl, "ws://") || strings.HasPrefix(requrl, "wss://"):
		mainProtocol = MPTypeWebSocket
	case strings.HasPrefix(requrl, "grpc://") || strings.HasPrefix(requrl, "rpc://"):
		mainProtocol = MPTypeGRPC
	case strings.HasPrefix(requrl, "radius://"):
		mainProtocol = MPTypeRadius
		requrl = requrl[9:]
	default:
		mainProtocol = MPTypeHTTP
		requrl = fmt.Sprintf("http://%s", requrl)
		if mainProtocol == "" {
			err = fmt.Errorf("url:%s 不合法,必须是完整http、webSocket连接", requrl)
			return
		}
	}

	// http和websocket默认检查方法赋值
	var ok bool
	switch mainProtocol {
	case MPTypeHTTP:
		// verify
		if verify == "" {
			verify = "statusCode"
		}
		key := fmt.Sprintf("%s.%s", mainProtocol, verify)
		_, ok = verifyMapHTTP[key]
		if !ok {
			err = errors.New("验证器不存在:" + key)
			return
		}
	case MPTypeWebSocket:
		// verify
		if verify == "" {
			verify = "json"
		}
		key := fmt.Sprintf("%s.%s", mainProtocol, verify)
		_, ok = verifyMapWebSocket[key]
		if !ok {
			err = errors.New("验证器不存在:" + key)
			return
		}
	}

	request = &RequestForm{
		URL:           requrl,
		MP:            mainProtocol,
		Method:        strings.ToUpper(method),
		Headers:       headers,
		Body:          body,
		Verify:        verify,
		ClientTimeout: clientTimeout,
		Debug:         debug,
		MaxCon:        maxCon,
		HTTP2:         http2,
		Keepalive:     keepalive,
		Code:          statusCode,
	}
	return
}

// Print 格式化打印
func (r *RequestForm) Print() {
	if r == nil {
		return
	}
	result := fmt.Sprintf("request:\n mainprotocol:%s \n url:%s \n method:%s \n headers:%v \n", r.MP, r.URL, r.Method,
		r.Headers)
	result = fmt.Sprintf("%s data:%v \n", result, r.Body)
	result = fmt.Sprintf("%s verify:%s \n clienttimeout:%s \n debug:%v \n", result, r.Verify, r.ClientTimeout, r.Debug)
	result = fmt.Sprintf("%s http2.0:%v \n keepalive:%v \n maxCon:%v ", result, r.HTTP2, r.Keepalive, r.MaxCon)
	fmt.Println(result)
}

// GetDebug 获取 debug 参数
func (r *RequestForm) GetDebug() bool {
	return r.Debug
}

// IsParameterLegal 参数是否合法
func (r *RequestForm) IsParameterLegal() (err error) {
	r.MP = "http"
	// statusCode json
	r.Verify = "json"
	key := fmt.Sprintf("%s.%s", r.MP, r.Verify)
	_, ok := verifyMapHTTP[key]
	if !ok {
		return errors.New("验证器不存在:" + key)
	}
	return
}

// RequestResults 请求结果
type RequestResults struct {
	ID            string // 消息ID
	ChanID        uint64 // 消息ID
	Time          uint64 // 请求时间 纳秒
	IsSucceed     bool   // 是否请求成功
	ErrCode       int    // 错误码
	ReceivedBytes int64
}

// SetID 设置请求唯一ID
func (r *RequestResults) SetID(chanID uint64, number uint64) {
	id := fmt.Sprintf("%d_%d", chanID, number)
	r.ID = id
	r.ChanID = chanID
}
