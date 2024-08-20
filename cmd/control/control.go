package main

import (
	"flag"
	"fmt"
	"strings"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/router"

	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	clientConfigPath = flag.String("clientConfigPath", "../../config/clientConfig.yaml", "config file path for client")
	serverConfigPath = flag.String("serverConfigPath", "../../config/serverConfig.yaml", "config file path for server")
	svrID            = flag.String("svrID", "server-2", "svrID(s) for server(s) need to update, split by ','")
	endpointsNum     = flag.Int("endpointsNum", 5, "number of endpoints need to change")
	op               = flag.String("op", "init", "oprations to do for config in etcd")
	namespace        = flag.String("namespace", "produce", "namespace need to check")
	srvName          = flag.String("svrName", "chatsvr", "svr config need to check")
)

func main() {
	flag.Parse()

	clientConf, err := config.LoadConfig(*clientConfigPath)
	if err != nil {
		log.Panic().Msg(err.Error())
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   clientConf.Etcd.Endpoints,
		DialTimeout: clientConf.Etcd.DialTimeout,
		Username:    clientConf.Etcd.Username,
		Password:    clientConf.Etcd.Password,
	})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	switch *op {
	case "init":
		config.SetBalancerRule(clientConf, client)
		config.SetRedisConfig(clientConf, client)
	case "set-client":
		config.SetBalancerRule(clientConf, client)
	case "set-server":
		serverConf, err := config.LoadConfig(*serverConfigPath)
		if err != nil {
			log.Panic().Msg(err.Error())
		}

		for _, svr := range strings.Split(*svrID, ",") {
			config.SetServerConfig(serverConf, svr, client, *endpointsNum)
		}
	case "close-server":
		serverConf, err := config.LoadConfig(*serverConfigPath)
		if err != nil {
			log.Panic().Msg(err.Error())
		}

		for _, svr := range strings.Split(*svrID, ",") {
			serverConf.Server[svr].Endpoint.State = router.State_Closing
			config.SetServerConfig(serverConf, svr, client, *endpointsNum)
		}
	case "up-server":
		serverConf, err := config.LoadConfig(*serverConfigPath)
		if err != nil {
			log.Panic().Msg(err.Error())
		}

		for _, svr := range strings.Split(*svrID, ",") {
			serverConf.Server[svr].Endpoint.State = router.State_Alive
			config.SetServerConfig(serverConf, svr, client, *endpointsNum)
		}
	case "clear":
		config.Clear(client)
	case "check-rule":
		fmt.Printf("%+v\n", *config.GetBalancerRule(*namespace, *srvName, client))
	case "check-redis":
		fmt.Printf("%+v\n", *config.GetRedisConfig(*namespace, client))
	}
}
