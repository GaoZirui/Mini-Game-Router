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
	sleepTime  = flag.Int("sleepTime", 500, "sleep time")
	usePayload = flag.Bool("usePayload", false, "whether to use 100B payload")
	showReply  = flag.Bool("showReply", false, "whether to show reply")
	countReply = flag.Bool("countReply", false, "whether to count reply")
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

	defer func(start time.Time) {
		elapsed := time.Since(start)
		fmt.Printf("total time: %s\n", elapsed)
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
					if *usePayload {
						reply, err = client.SayHello(ctx, &pb.HelloRequest{
							Name: payload,
						})
					} else {
						reply, err = client.SayHello(ctx, &pb.HelloRequest{
							Name: "user" + strconv.Itoa(userID) + " test" + strconv.Itoa(i),
						})
					}
					timer.ObserveDuration()
					return err
				}, func() {})
				if *countReply {
					parts := strings.Split(reply.Message, " ")
					clientMetrics.RequestDestination.WithLabelValues(parts[len(parts)-1]).Inc()
				}
				if *showReply {
					fmt.Println(reply.Message)
				}
				clientMetrics.TotalResponseNumber.Inc()
				time.Sleep(time.Millisecond * (time.Duration(*sleepTime)))
			}
		}()
	}
	wg.Wait()
}
