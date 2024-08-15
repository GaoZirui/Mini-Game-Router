package nettoolkit

import (
	"context"
	"fmt"
	"strings"
	"time"
	"ziruigao/mini-game-router/core/cache"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/metrics"
	"ziruigao/mini-game-router/core/router"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type NetToolkit struct {
	EtcdClient  *clientv3.Client
	RedisClient *redis.ClusterClient
	Namespace   string
}

var netToolkit *NetToolkit

func GetNetToolkit() *NetToolkit {
	return netToolkit
}

func GetEtcdClient() *clientv3.Client {
	return netToolkit.EtcdClient
}

func GetRedisClient() *redis.ClusterClient {
	return netToolkit.RedisClient
}

func GetNamespace() string {
	return netToolkit.Namespace
}

func Init(conf *config.EtcdConfig, namespace string) {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Endpoints,
		DialTimeout: conf.DialTimeout,
		Username:    conf.Username,
		Password:    conf.Password,
	})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	redisConfig := config.GetRedisConfig(namespace, etcdCli)

	redisCli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        redisConfig.Addrs,
		Username:     redisConfig.Username,
		Password:     redisConfig.Password,
		DialTimeout:  redisConfig.DialTimeout,
		ReadTimeout:  redisConfig.ReadTimeout,
		WriteTimeout: redisConfig.WriteTimeout,
	})

	netToolkit = &NetToolkit{
		EtcdClient:  etcdCli,
		RedisClient: redisCli,
		Namespace:   namespace,
	}
}

func SetEndpoint(key string, endpoint *router.Endpoint, timeout time.Duration) {
	luaScript := `
		local oldValue = redis.call('GET', KEYS[1])
		redis.call('SET', KEYS[1], ARGV[1])
		local expiration = tonumber(ARGV[2])
		if expiration ~= 0 then
			redis.call('EXPIRE', KEYS[1], expiration)
		end
		if (oldValue ~= nil) and (oldValue ~= ARGV[1]) then
			redis.call('PUBLISH', ARGV[3], KEYS[1])
		end
	`
	err := netToolkit.RedisClient.Eval(context.Background(), luaScript, []string{netToolkit.Namespace + "/" + key}, endpoint.ToString(), timeout.Seconds(), strings.Split(key, "-")[0]).Err()
	if err != nil && err != redis.Nil {
		log.Fatal().Msg(err.Error())
	}
	log.Debug().Msg(fmt.Sprintf("[redis] | put key: %v val: %v\n", netToolkit.Namespace+"/"+key, endpoint.ToString()))
}

func GetEndpoint(key string) string {
	ep, err := netToolkit.RedisClient.Get(context.Background(), netToolkit.Namespace+"/"+key).Result()
	if err == redis.Nil {
		return ""
	}
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return ep
}

func ClearEndpoint(key string) {
	luaScript := `
		redis.call('DEL', KEYS[1])
		redis.call('PUBLISH', 'ARGV[1]', KEYS[1])
	`
	err := netToolkit.RedisClient.Eval(context.Background(), luaScript, []string{netToolkit.Namespace + "/" + key}, strings.Split(key, "-")[0]).Err()
	if err != nil && err != redis.Nil {
		log.Fatal().Msg(err.Error())
	}
	log.Debug().Msg(fmt.Sprintf("[redis] | del key: %v\n", netToolkit.Namespace+"/"+key))
}

func RenewalEndpoint(key string, expiration time.Duration) {
	err := netToolkit.RedisClient.Expire(context.Background(), netToolkit.Namespace+"/"+key, expiration).Err()
	if err != nil && err != redis.Nil {
		log.Fatal().Msg(err.Error())
	}
}

func Close() {
	netToolkit.EtcdClient.Close()
	netToolkit.RedisClient.Close()
}

func GetServerPerformance(ep *router.Endpoint) *metrics.ServerPerformance {
	key := "performance/" + ep.Namespace + "/" + ep.Name + "/" + ep.ToAddr()

	resp, err := netToolkit.EtcdClient.Get(context.Background(), key)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return metrics.ParseServerPerformance(string(resp.Kvs[0].Value))
}

func Subscribe(channel string, cache *cache.LRUCache) {
	pubsub := netToolkit.RedisClient.Subscribe(context.Background(), channel)
	defer pubsub.Close()
	ch := pubsub.Channel()
	for msg := range ch {
		cache.Delete(msg.Payload)
	}
}
