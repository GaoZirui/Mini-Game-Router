package main

import (
	"flag"
	"fmt"
	"net/http"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/etcd"
	"ziruigao/mini-game-router/core/metrics"
	"ziruigao/mini-game-router/core/sdk"

	"github.com/rs/zerolog/log"
)

var (
	configPath    = flag.String("configPath", "serverConfig.yaml", "config file path")
	svrID         = flag.String("svrID", "svr", "id for server")
	conf          config.ServerConfig
	serverMetrics *metrics.ServerMetrics
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, this is %s!", conf.Endpoint.ToAddr())
}

func main() {
	flag.Parse()

	config, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Panic().Msg(err.Error())
	}

	conf = config.Server[*svrID]

	serverMetrics = metrics.NewServerMetrics()

	http.HandleFunc("/", helloHandler)

	_, err = sdk.RegisterAndStart(&etcd.ServiceRegisterOpts{
		EtcdConfig:     config.Etcd,
		EtcdLease:      conf.Lease,
		Endpoint:       &conf.Endpoint,
		ServiceMetrics: serverMetrics,
	})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	fmt.Printf("Start listening at %v\n", conf.Endpoint.ToAddr())
	if err := http.ListenAndServe(conf.Endpoint.ToAddr(), nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
