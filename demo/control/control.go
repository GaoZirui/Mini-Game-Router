package main

import (
	"flag"
	"fmt"
	"ziruigao/mini-game-router/core/config"

	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	configPath = flag.String("configPath", "clientConfig.yaml", "config file path")
	op         = flag.String("op", "set", "oprations to do for config in etcd")
	namespace  = flag.String("namespace", "produce", "namespace need to check")
	srvName    = flag.String("svrName", "chatsvr", "svr config need to check")
)

func main() {
	flag.Parse()

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Panic().Msg(err.Error())
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Etcd.Endpoints,
		DialTimeout: conf.Etcd.DialTimeout,
		Username:    conf.Etcd.Username,
		Password:    conf.Etcd.Password,
	})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	switch *op {
	case "set":
		config.SetBalancerRule(conf, client)
		config.SetRedisConfig(conf, client)
	case "clear":
		config.Clear(client)
	case "check-rule":
		fmt.Printf("%+v\n", *config.GetBalancerRule(*namespace, *srvName, client))
	case "check-redis":
		fmt.Printf("%+v\n", *config.GetRedisConfig(*namespace, client))
	}
}
