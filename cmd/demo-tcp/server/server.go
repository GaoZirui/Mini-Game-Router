package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/etcd"
	"ziruigao/mini-game-router/core/metrics"
	"ziruigao/mini-game-router/core/sdk"

	"github.com/rs/zerolog/log"
)

var (
	configPath    = flag.String("configPath", "serverConfig.yaml", "config file path")
	svrID         = flag.String("svrID", "svr", "id for server")
	endpointsNum  = flag.Int64("endpointsNum", 1, "num of endpoints for this server")
	conf          *config.ServerConfig
	serverMetrics *metrics.ServerMetrics
)

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
		defer listener.Close()

		_, err = sdk.RegisterAndStart(&etcd.ServiceRegisterOpts{
			EtcdConfig:     config.Etcd,
			EtcdLease:      conf.Lease,
			Endpoint:       &ep,
			ServiceMetrics: serverMetrics,
		})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					fmt.Println("Error accepting on port", port, ":", err.Error())
					continue
				}
				go handleConnection(conn)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	fmt.Println("server quit")
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client closed the connection")
			} else {
				fmt.Println("Error reading:", err.Error())
			}
			break
		}
		message = strings.TrimRight(message, "\n")
		fmt.Printf("Received: ")
		_, err = conn.Write([]byte(fmt.Sprintf("Hello: %s this is %v\n", message, conn.LocalAddr())))
		if err != nil {
			fmt.Println("Error sending response:", err.Error())
			break
		}
	}
}
