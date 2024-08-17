package mybalancer

import (
	"ziruigao/mini-game-router/core/cache"
	"ziruigao/mini-game-router/core/config"
	nettoolkit "ziruigao/mini-game-router/core/netToolkit"
	"ziruigao/mini-game-router/core/router"

	"github.com/rs/zerolog/log"
)

type MyBalancer interface {
	Init(*config.BalancerRule)
	Pick(*router.Metadata) *router.Endpoint
	Add(*router.Endpoint)
	Remove(*router.Endpoint)
	GetAll() []*router.Endpoint
	GetCache() cache.Cache
	Name() string
	New() MyBalancer
	Stop()
}

var (
	balancerRegistry map[string]MyBalancer
	balancers        map[string]MyBalancer
)

func RegisterBalancer(name string, balancer MyBalancer) {
	balancerRegistry[name] = balancer
}

func InitBalancers() {
	balancers = map[string]MyBalancer{}
}

func GetBalancer(name string) MyBalancer {
	return balancers[name]
}

func SetBalancer(name string, balancer MyBalancer) {
	balancers[name] = balancer
}

func InitRegistry() {
	balancerRegistry = map[string]MyBalancer{}

	RegisterBalancer("random", &RandomBalancer{})
	RegisterBalancer("weight", &WeightBalancer{})
	RegisterBalancer("consistent-hash", &ConsistentHashBalancer{})
	RegisterBalancer("static", &StaticBalancer{})
	RegisterBalancer("dynamic", &DynamicBalancer{})

}

func MyBalancerFactory(config *config.BalancerRule) MyBalancer {
	if balancer, exists := balancerRegistry[config.BalancerType]; exists {
		blc := balancer.New()
		blc.Init(config)
		return blc
	}
	log.Fatal().Msg("invalid balancer type")
	return nil
}

func LoadBalancer(svrName string) MyBalancer {
	balancerRule := config.GetBalancerRule(nettoolkit.GetNamespace(), svrName, nettoolkit.GetEtcdClient())
	return MyBalancerFactory(balancerRule)
}
