package config

import (
	"context"
	"os"
	"strconv"
	"time"
	"ziruigao/mini-game-router/core/router"

	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Etcd     *EtcdConfig                        `yaml:"etcd"`
	Balancer map[string]map[string]BalancerRule `yaml:"balancer"`
	Redis    map[string]*RedisConfig            `yaml:"redis"`
	Server   map[string]*ServerConfig           `yaml:"server"`
}

type ServerConfig struct {
	Endpoint router.Endpoint `yaml:"endpoint"`
	Lease    int64           `yaml:"lease"`
}

type EtcdConfig struct {
	Endpoints   []string      `yaml:"endpoints"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Username    string        `yaml:"username"`
	Password    string        `yaml:"password"`
	RecoverTime time.Duration `yaml:"recover_time"`
}

type BalancerRule struct {
	BalancerType         string               `yaml:"balancer_type"`
	ConsistentHashConfig ConsistentHashConfig `yaml:"consistent_hash_config"`
	StaticConfig         StaticConfig         `yaml:"static_config"`
	DynamicConfig        DynamicConfig        `yaml:"dynamic_config"`
	CostumConfig         map[string]string    `yaml:"custom_config"`
}

func (b *BalancerRule) ToString() string {
	data, err := yaml.Marshal(b)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return string(data)
}

func ParseBalancerRule(s string) *BalancerRule {
	var b BalancerRule
	err := yaml.Unmarshal([]byte(s), &b)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return &b
}

type ConsistentHashConfig struct {
	HashFunc string `yaml:"hash_func"`
	Replicas int    `yaml:"replicas"`
	Key      string `yaml:"key"`
}

type StaticConfig struct {
	Key string `yaml:"key"`
}

type DynamicConfig struct {
	Key       string `yaml:"key"`
	Cache     bool   `yaml:"cache"`
	CacheType string `yaml:"cache_type"`
	CacheSize int    `yaml:"cache_size"`
	AutoFlush bool   `yaml:"auto_flush"`
}

type RedisConfig struct {
	Addrs        []string      `yaml:"addrs"`
	Username     string        `yaml:"username"`
	Password     string        `yaml:"password"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

func (r *RedisConfig) ToString() string {
	data, err := yaml.Marshal(r)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return string(data)
}

func ParseRedisConfig(s string) *RedisConfig {
	var r RedisConfig
	err := yaml.Unmarshal([]byte(s), &r)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return &r
}

func LoadConfig(configPath string) (*Config, error) {
	var config Config
	var yamlBytes []byte

	if b, err := os.ReadFile(configPath); err != nil {
		return nil, err
	} else {
		// 扩充环境变量
		yamlBytes = []byte(os.ExpandEnv(string(b)))
	}

	if err := yaml.Unmarshal(yamlBytes, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func SetBalancerRule(config *Config, client *clientv3.Client) {
	for namespace, balancerRules := range config.Balancer {
		for svrName, balancerRule := range balancerRules {
			_, err := client.Put(context.Background(), "config/"+namespace+"/"+svrName, balancerRule.ToString())
			if err != nil {
				log.Fatal().Msg(err.Error())
			}
		}
	}
}

func SetServerConfig(config *Config, svrID string, client *clientv3.Client, endpointsNum int) {
	serverConfig := config.Server[svrID]

	for i := 0; i < endpointsNum; i++ {
		ep := serverConfig.Endpoint

		port, _ := strconv.Atoi(ep.Port)
		port += i
		ep.Port = strconv.Itoa(port)

		key := ep.Namespace + "/" + ep.Name + "/" + ep.ToAddr()

		resp, err := client.Get(context.Background(), key)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		if len(resp.Kvs) == 0 {
			log.Fatal().Msg("key not found")
		}
		leaseID := resp.Kvs[0].Lease
		_, err = client.Put(context.Background(), key, ep.ToString(), clientv3.WithLease(clientv3.LeaseID(leaseID)))
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}
}

func SetRedisConfig(config *Config, client *clientv3.Client) {
	for namespace, redisConfig := range config.Redis {
		_, err := client.Put(context.Background(), "config/"+namespace+"/redis", redisConfig.ToString())
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}
}

func GetBalancerRule(namespace, svrName string, client *clientv3.Client) *BalancerRule {
	resp, err := client.Get(context.Background(), "config/"+namespace+"/"+svrName)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return ParseBalancerRule(string(resp.Kvs[0].Value))
}

func GetRedisConfig(namespace string, client *clientv3.Client) *RedisConfig {
	resp, err := client.Get(context.Background(), "config/"+namespace+"/redis")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return ParseRedisConfig(string(resp.Kvs[0].Value))
}

func Clear(client *clientv3.Client) {
	_, err := client.Delete(context.Background(), "config", clientv3.WithPrefix())
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	_, err = client.Delete(context.Background(), "performance", clientv3.WithPrefix())
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
}
