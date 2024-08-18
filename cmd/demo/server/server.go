package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/etcd"
	"ziruigao/mini-game-router/core/metrics"
	"ziruigao/mini-game-router/core/sdk"
	pb "ziruigao/mini-game-router/proto"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var (
	configPath    = flag.String("configPath", "serverConfig.yaml", "config file path")
	svrID         = flag.String("svrID", "svr", "id for server")
	endpointsNum  = flag.Int64("endpointsNum", 1, "num of endpoints for this server")
	showReceive   = flag.Bool("showReceive", false, "whether to show receive")
	conf          *config.ServerConfig
	serverMetrics *metrics.ServerMetrics
)

type HelloServer struct {
	pb.UnimplementedHelloServiceServer
}

func (s *HelloServer) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	timer := prometheus.NewTimer(serverMetrics.RequestDurations)
	defer func() {
		timer.ObserveDuration()
		serverMetrics.TotalRequestNumber.Inc()
		serverMetrics.AddRequestNum()
	}()
	if *showReceive {
		fmt.Println("receive: ", request.Name)
	}
	return &pb.HelloReply{
		Message: "hello: " + request.Name + " this is " + conf.Endpoint.Port,
	}, nil
}

func main() {
	flag.Parse()

	config, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Panic().Msg(err.Error())
	}

	conf = config.Server[*svrID]

	serverMetrics = metrics.NewServerMetrics()
	for i := 0; i < int(*endpointsNum); i++ {
		ep := conf.Endpoint

		port, _ := strconv.Atoi(ep.Port)
		port += i
		ep.Port = strconv.Itoa(port)

		listener, err := net.Listen("tcp", conf.Endpoint.IP+":"+ep.Port)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		fmt.Println("listening at ", listener.Addr())

		go func() {
			s := grpc.NewServer()
			pb.RegisterHelloServiceServer(s, &HelloServer{})
			if err := s.Serve(listener); err != nil {
				log.Fatal().Msg(err.Error())
			}
		}()
		_, err = sdk.RegisterAndStart(&etcd.ServiceRegisterOpts{
			EtcdConfig:     config.Etcd,
			EtcdLease:      conf.Lease,
			Endpoint:       &ep,
			ServiceMetrics: serverMetrics,
		})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	fmt.Println("server quit")
}
