package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/etcd"
	"ziruigao/mini-game-router/core/metrics"
	"ziruigao/mini-game-router/core/sdk"
	pb "ziruigao/mini-game-router/proto"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

	sdk.InitForGrpc(config.Etcd, *namespace)

	clientMetrics := metrics.NewClientMetrics()

	conn, err := grpc.NewClient("grpclb:///chatsvr", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"mybalancer"}`))
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	defer conn.Close()

	wg := sync.WaitGroup{}

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

		// blc := mybalancer.GetBalancer(*namespace + "/chatsvr/")
		// if blc != nil {
		// 	if dynamicBalancer, ok := blc.(*mybalancer.DynamicBalancer); ok {
		// 		dynamicBalancer.Rate()
		// 	}
		// }

		// blc := mybalancer.GetBalancer(*namespace + "/chatsvr/")
		// if blc != nil {
		// 	if balancer, ok := blc.(*mybalancer.ConsistentHashBalancer); ok {
		// 		balancer.Rate()
		// 	}
		// }
	}(time.Now())

	for user := 0; user < int(*userNum); user++ {
		userID := user
		ep := sdk.PickEndpointByServerPerformance("chatsvr", etcd.Random)
		// fmt.Printf("user%v choose: %v\n", userID, ep.Port)
		sdk.SetEndpoint("chat-user"+strconv.Itoa(userID), ep, 0)
		wg.Add(1)
		client := pb.NewHelloServiceClient(conn)
		go func() {
			defer wg.Done()
			for i := 0; i < int(*requestNum); i++ {
				ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("hash-key", "user"+strconv.Itoa(userID), "chat-user-id", "chat-user"+strconv.Itoa(userID), "user-id", "user"+strconv.Itoa(userID)))
				reply := &pb.HelloReply{}
				sdk.CallWithRetry(func() error {
					timer := prometheus.NewTimer(clientMetrics.RequestDurations)
					reply, err = client.SayHello(ctx, &pb.HelloRequest{
						Name: payload,
					})
					timer.ObserveDuration()
					return err
				}, func() {})
				if strings.HasSuffix(reply.Message, "10000") {
					atomic.AddInt64(&cnt1, 1)
				} else if strings.HasSuffix(reply.Message, "12000") {
					atomic.AddInt64(&cnt2, 1)
				} else {
					atomic.AddInt64(&cnt3, 1)
				}
				// fmt.Println(reply.Message)
				clientMetrics.TotalResponseNumber.Inc()
				time.Sleep(time.Millisecond * 500)
			}
		}()
	}
	wg.Wait()
}
