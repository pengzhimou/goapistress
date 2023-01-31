// Package server 压测启动
package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"goapistress/model"
	"goapistress/server/client"
	"goapistress/server/golink"
	"goapistress/server/statistics"
	"goapistress/server/verify"
)

const (
	connectionMode = 1 // 1:顺序建立长链接 2:并发建立长链接
)

// init 注册验证器
func init() {

	// http
	model.RegisterVerifyHTTP("statusCode", verify.HTTPStatusCode)
	model.RegisterVerifyHTTP("json", verify.HTTPJson)

	// webSocket
	model.RegisterVerifyWebSocket("json", verify.WebSocketJSON)
}

// Dispose 处理函数
func Dispose(ctx context.Context, concurrency, reqNumbersPerProd uint64, reqForm *model.RequestForm) {
	// 设置接收数据缓存
	ch := make(chan *model.RequestResults, 1000)
	var (
		wg          sync.WaitGroup // 发送数据完成
		wgReceiving sync.WaitGroup // 数据处理完成
	)
	wgReceiving.Add(1)
	go statistics.ReceivingResults(concurrency, ch, &wgReceiving)

	for chanID := uint64(0); chanID < concurrency; chanID++ {
		wg.Add(1)
		switch reqForm.MP {
		case model.MPTypeHTTP:
			go golink.HTTP(ctx, chanID, ch, reqNumbersPerProd, &wg, reqForm)
		case model.MPTypeWebSocket:
			switch connectionMode {
			case 1:
				// 连接以后再启动协程
				ws := client.NewWebSocket(reqForm.URL)
				err := ws.GetConn()
				if err != nil {
					fmt.Println("连接失败:", chanID, err)
					continue
				}
				go golink.WebSocket(ctx, chanID, ch, reqNumbersPerProd, &wg, reqForm, ws)
			case 2:
				// 并发建立长链接
				go func(i uint64) {
					// 连接以后再启动协程
					ws := client.NewWebSocket(reqForm.URL)
					err := ws.GetConn()
					if err != nil {
						fmt.Println("连接失败:", i, err)
						return
					}
					golink.WebSocket(ctx, i, ch, reqNumbersPerProd, &wg, reqForm, ws)
				}(chanID)
				// 注意:时间间隔太短会出现连接失败的报错 默认连接时长:20毫秒(公网连接)
				time.Sleep(5 * time.Millisecond)
			default:
				data := fmt.Sprintf("不支持的类型:%d", connectionMode)
				panic(data)
			}
		case model.MPTypeGRPC:
			// 连接以后再启动协程
			ws := client.NewGrpcSocket(reqForm.URL)
			err := ws.Link()
			if err != nil {
				fmt.Println("连接失败:", chanID, err)
				continue
			}
			go golink.Grpc(ctx, chanID, ch, reqNumbersPerProd, &wg, reqForm, ws)
		case model.MPTypeRadius:
			// Radius use udp, does not a connection
			go golink.Radius(ctx, chanID, ch, reqNumbersPerProd, &wg, reqForm)

		default:
			// 类型不支持
			wg.Done()
		}
	}
	// 等待所有的数据都发送完成
	wg.Wait()
	// 延时1毫秒 确保数据都处理完成了
	time.Sleep(1 * time.Millisecond)
	close(ch)
	// 数据全部处理完成了
	wgReceiving.Wait()
	return
}
