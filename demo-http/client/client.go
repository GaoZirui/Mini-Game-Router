package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/etcd"
	"ziruigao/mini-game-router/core/router"
	"ziruigao/mini-game-router/core/sdk"

	"github.com/rs/zerolog/log"
)

// CustomRoundTripper 实现了 http.RoundTripper 接口
type CustomRoundTripper struct {
	Transport http.RoundTripper
}

var (
	configPath = flag.String("configPath", "clientConfig.yaml", "config file path")
	userNum    = flag.Int64("userNum", 10000, "number of users")
	requestNum = flag.Int64("requestNum", 1000, "request number for each user")
	sleepTime  = flag.Int64("sleepTime", 50, "sleep time(ms) between each request")
	namespace  = flag.String("namespace", "produce", "namespace for grpc call")
)

// RoundTrip 实现了 http.RoundTripper 接口的 RoundTrip 方法
func (t *CustomRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	metadata := router.NewMetadata(context.Background())
	metadata.Set("hash-key", "test100")
	metadata.Set("user-id", "user3")
	metadata.Set("chat-user-id", "chat-user1")
	ep := sdk.PickEndpoint(metadata, req.Host)
	req.URL.Host = ep.ToAddr()
	req.Host = ep.ToAddr()

	// 使用原始的 Transport 发送请求
	if t.Transport == nil {
		t.Transport = http.DefaultTransport
	}
	return t.Transport.RoundTrip(req)
}

func main() {
	flag.Parse()
	// 创建自定义的 RoundTripper
	customTransport := &CustomRoundTripper{}

	// 创建一个 http.Client 并设置自定义的 RoundTripper
	client := &http.Client{
		Transport: customTransport,
		Timeout:   5 * time.Second,
	}

	config, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Panic().Msg(err.Error())
	}

	sdk.Init(config.Etcd, *namespace)
	sdk.Discovery("chatsvr")

	wg := sync.WaitGroup{}
	sdk.SetEndpoint("chat-user1", sdk.PickEndpointByServerPerformance("chatsvr", etcd.Random), 0)
	for user := 0; user < int(*userNum); user++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < int(*requestNum); i++ {
				// 发送 GET 请求
				var resp *http.Response
				sdk.CallWithRetry(func() error {
					resp, err = client.Get("http://chatsvr")
					return err
				})
				defer resp.Body.Close()

				// 读取响应内容
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				// 打印响应内容
				fmt.Println("Response:", string(body))
				time.Sleep(time.Millisecond * time.Duration(*sleepTime))
			}
		}()
	}
	wg.Wait()
}
