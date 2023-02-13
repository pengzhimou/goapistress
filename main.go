// Package main go 实现的压测工具
package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"

	"goapistress/model"
	"goapistress/server"
	"goapistress/tools"
)

var (
	concurrency       uint64          = 1                                        // 并发数
	reqNumbersPerProd uint64          = 1                                        // 请求数(单个并发/协程)
	debugStr                          = "false"                                  // 是否是debug
	curlFilePath                      = ""                                       // curl文件路径 http接口压测，自定义参数设置
	requestURL                        = ""                                       // 压测的url 目前支持，http/https ws/wss
	method                            = "GET"                                    // http 方法
	headers           tools.FlagSlice = make(tools.FlagSlice, 0)                 // 自定义头信息传递给服务器
	body                              = ""                                       // HTTP POST方式传送数据
	verify            tools.FlagMap   = make(tools.FlagMap)                      // verify 验证方法 -v statusCode=200 -v json="{\"code\":0}"
	statusCode                        = 200                                      // 成功状态码
	maxCon                            = 1                                        // 单个连接最大请求数
	http2                             = false                                    // 是否开http2.0
	keepalive                         = false                                    // 是否开启长连接
	cpuNumber                         = runtime.GOMAXPROCS(runtime.NumCPU() - 1) // CPU 核数，一般场景下单核已经够用了
	clientTimeout     int             = 30                                       // http client超时时间，默认不设置
	taskTimeout       int             = 3600                                     // task timeout context 控制
	certfile                          = ""                                       // https cert
	keyfile                           = ""                                       // https key
	authtype                          = ""                                       // auth type, basic tkextoken tkxhttps
	authdata          tools.FlagMap   = make(tools.FlagMap)                      // auth data, -
)

func init() {
	flag.Uint64Var(&concurrency, "c", concurrency, "并发数")
	flag.Uint64Var(&reqNumbersPerProd, "n", reqNumbersPerProd, "请求数(单个并发/协程)")
	flag.StringVar(&debugStr, "d", debugStr, "调试模式")
	flag.StringVar(&curlFilePath, "p", curlFilePath, "curl文件路径")
	flag.StringVar(&requestURL, "u", requestURL, "压测地址")
	flag.StringVar(&method, "x", method, "http请求方法")
	flag.Var(&verify, "v", "验证方法 -v statusCode=200 -v json=\"{\"code\":0}\"")
	flag.Var(&headers, "H", "自定义头信息传递给服务器 示例:-H 'Content-Type: application/json'")
	flag.StringVar(&body, "data", body, "HTTP POST方式传送数据")
	flag.IntVar(&maxCon, "m", maxCon, "单个host最大连接数")
	flag.IntVar(&statusCode, "statusCode", statusCode, "请求成功的状态码")
	flag.BoolVar(&http2, "http2", http2, "是否开http2.0")
	flag.BoolVar(&keepalive, "k", keepalive, "是否开启长连接")
	flag.IntVar(&cpuNumber, "cpuNumber", cpuNumber, "CPU 核数，默认为一核")
	flag.IntVar(&clientTimeout, "clientTimeout", clientTimeout, "超时时间 单位 秒,默认30")
	flag.IntVar(&taskTimeout, "taskTimeout", taskTimeout, "超时时间 单位 秒,默认3600")
	flag.StringVar(&certfile, "cert", certfile, "https cert")
	flag.StringVar(&keyfile, "key", keyfile, "https key")
	flag.StringVar(&authtype, "at", authtype, "auth type, -at basic/tkextoken/forcehttps")
	flag.Var(&authdata, "ad", "auth data, -ad username=xxxxx -ad password=xxxx/-ad tkextoken=xxxx/-ad cert=./xxxxxx -ad key=./xxxxx/-ad token=xxxxxxx")

	// 解析参数
	flag.Parse()
}

// handle args
func argsCheck() bool {
	if concurrency == 0 || reqNumbersPerProd == 0 || (requestURL == "" && curlFilePath == "") {
		fmt.Printf("示例: go run main.go -c 1 -n 1 -u https://www.baidu.com/ \n")
		fmt.Printf("压测地址或curl路径必填 \n")
		fmt.Printf("当前请求参数: -c %d -n %d -d %v -u %s \n", concurrency, reqNumbersPerProd, debugStr, requestURL)
		flag.Usage()
		return false
	}
	return true
}

func genRequestForm() *model.RequestForm {
	debug := strings.ToLower(debugStr) == "true"
	reqform, err := model.NewReqForm(requestURL, method, verify, statusCode, time.Duration(clientTimeout)*time.Second, debug, curlFilePath, headers, body, maxCon, http2, keepalive, certfile, keyfile, authtype, authdata)
	if err != nil {
		fmt.Printf("参数不合法 %v \n", err)
		return nil
	}
	fmt.Printf("\n 开始启动  并发数:%d 请求数:%d 请求参数: \n", concurrency, reqNumbersPerProd)
	reqform.Print()
	return reqform
}

func runStress(reqform *model.RequestForm) {
	ctx := context.Background()
	if taskTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(taskTimeout)*time.Second)
		defer cancel()
		deadline, ok := ctx.Deadline()
		if ok {
			fmt.Printf(" deadline %s", deadline)
		}
	}
	server.Dispose(ctx, concurrency, reqNumbersPerProd, reqform)
}

// main go 实现的压测工具
// 编译可执行文件
//
//go:generate go build main.go
func main() {
	runtime.GOMAXPROCS(cpuNumber)

	// args check
	if !argsCheck() {
		return
	}

	// gen requester
	reqForm := genRequestForm()
	if reqForm == nil {
		return
	}

	// 开始处理
	runStress(reqForm)
}
