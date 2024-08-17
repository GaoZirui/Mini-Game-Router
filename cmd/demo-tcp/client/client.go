package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"
	mybalancer "ziruigao/mini-game-router/core/balancer"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/metrics"
	"ziruigao/mini-game-router/core/router"
	"ziruigao/mini-game-router/core/sdk"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	configPath = flag.String("configPath", "clientConfig.yaml", "config file path")
	userNum    = flag.Int64("userNum", 10000, "number of users")
	requestNum = flag.Int64("requestNum", 1000, "request number for each user")
	namespace  = flag.String("namespace", "produce", "namespace for grpc call")
	debug      = flag.Bool("debug", false, "set log level to debug")
	coreNum    = flag.Int("coreNum", 1, "num of cpu to use")
)

const (
	payload = "AAAAAAAAAA" +
		"AAAAAAAAAA" +
		"AAAAAAAAAA" +
		"AAAAAAAAAA" +
		"AAAAAAAAAA" +
		"AAAAAAAAAA" +
		"AAAAAAAAAA" +
		"AAAAAAAAAA" +
		"AAAAAAAAAA" +
		"AAAAAAAAAA"
)

type Client struct {
	conn     net.Conn
	response chan struct{}
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	client := &Client{
		conn:     conn,
		response: make(chan struct{}),
	}
	go client.readResponse()
	return client, nil
}

func (c *Client) readResponse() {
	reader := bufio.NewReader(c.conn)
	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			close(c.response)
			return
		}
		// fmt.Print(message)
		c.response <- struct{}{}
	}
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) SendMessage(message string) error {
	_, err := c.conn.Write([]byte(message + "\n"))
	if err != nil {
		return err
	}
	_ = <-c.response
	return nil
}

func main() {
	f, _ := os.OpenFile("cpu.pprof", os.O_CREATE|os.O_RDWR, 0644)
	defer f.Close()
	flag.Parse()

	runtime.GOMAXPROCS(*coreNum)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	config, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Panic().Msg(err.Error())
	}

	sdk.Init(config.Etcd, *namespace)
	sdk.Discovery("chatsvr")

	clients := map[string]*Client{}

	for _, ep := range sdk.GetAllEndpoints("chatsvr") {
		client, err := NewClient(ep.ToAddr())
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		defer client.Close()
		clients[ep.ToAddr()] = client
	}

	wg := sync.WaitGroup{}

	clientMetrics := metrics.NewClientMetrics()

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	var (
		cnt1 int64 = 0
		cnt2 int64 = 0
		cnt3 int64 = 0
	)

	defer func(start time.Time) {
		elapsed := time.Since(start)
		fmt.Printf("total time: %s\n", elapsed)
		fmt.Printf("cnt1: %v cnt2 %v cnt3 %v\n", cnt1, cnt2, cnt3)

		blc := mybalancer.GetBalancer(*namespace + "/chatsvr/")
		if blc != nil {
			if dynamicBalancer, ok := blc.(*mybalancer.DynamicBalancer); ok {
				dynamicBalancer.Rate()
			}
		}
	}(time.Now())

	for user := 0; user < int(*userNum); user++ {
		userID := user
		// ep := sdk.PickEndpointByServerPerformance("chatsvr", etcd.Random)
		// fmt.Printf("user%v choose: %v\n", userID, ep.Port)
		// sdk.SetEndpoint("chat-user"+strconv.Itoa(userID), ep, 0)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < int(*requestNum); i++ {
				timer := prometheus.NewTimer(clientMetrics.RequestDurations)
				ep := sdk.PickEndpoint(&router.Metadata{
					Metadata: map[string]string{
						"hash-key":     "user" + strconv.Itoa(userID),
						"chat-user-id": "chat-user" + strconv.Itoa(userID),
						"user-id":      "user" + strconv.Itoa(userID),
					},
				}, "chatsvr")
				sdk.CallWithRetry(func() error {
					err := clients[ep.ToAddr()].SendMessage(payload)
					return err
				}, func() {})
				timer.ObserveDuration()
				// if strings.HasSuffix(reply.Message, "10000") {
				// 	atomic.AddInt64(&cnt1, 1)
				// } else if strings.HasSuffix(reply.Message, "12000") {
				// 	atomic.AddInt64(&cnt2, 1)
				// } else {
				// 	atomic.AddInt64(&cnt3, 1)
				// }
				// fmt.Println(reply.Message)
				clientMetrics.TotalResponseNumber.Inc()
			}
		}()
	}
	wg.Wait()
}
