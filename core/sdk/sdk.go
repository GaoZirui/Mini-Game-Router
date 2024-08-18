package sdk

import (
	"sync"
	"time"
	mybalancer "ziruigao/mini-game-router/core/balancer"
	"ziruigao/mini-game-router/core/cache"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/etcd"
	nettoolkit "ziruigao/mini-game-router/core/netToolkit"
	myresolver "ziruigao/mini-game-router/core/resolver"
	"ziruigao/mini-game-router/core/router"

	"github.com/rs/zerolog/log"
)

var (
	onceForGrpc sync.Once
	once        sync.Once
)

func InitForGrpc(config *config.EtcdConfig, namespace string) {
	onceForGrpc.Do(func() {
		nettoolkit.Init(config, namespace)
		etcd.Init(nettoolkit.GetEtcdClient(), config.RecoverTime)
		myresolver.Init()
		mybalancer.Init()
		mybalancer.InitBalancers()
		mybalancer.InitRegistry()
		cache.InitRegistry()
	})
}

func Init(config *config.EtcdConfig, namespace string) {
	once.Do(func() {
		nettoolkit.Init(config, namespace)
		etcd.Init(nettoolkit.GetEtcdClient(), config.RecoverTime)
		mybalancer.InitBalancers()
		mybalancer.InitRegistry()
		cache.InitRegistry()
	})
}

func RegisterAndStart(opts *etcd.ServiceRegisterOpts) (*etcd.ServiceRegister, error) {
	return etcd.NewServiceRegister(opts)
}

func CloseServer(s *etcd.ServiceRegister) {
	s.Close()
}

func Discovery(svrName string) chan etcd.EtcdEvent {
	return etcd.WatchPrefix(nettoolkit.GetNamespace(), svrName)
}

func RegisterBalancer(name string, balancer mybalancer.MyBalancer) {
	mybalancer.RegisterBalancer(name, balancer)
}

func RegisterCache(name string, c cache.Cache) {
	cache.RegisterCache(name, c)
}

func GetAllEndpoints(svrName string) []*router.Endpoint {
	blc := mybalancer.GetBalancer(nettoolkit.GetNamespace() + "/" + svrName + "/")
	if blc == nil {
		log.Fatal().Msg("balancer not exists")
		return nil
	}
	return blc.GetAll()
}

func GetBalancer(svrName string) mybalancer.MyBalancer {
	return mybalancer.GetBalancer(nettoolkit.GetNamespace() + "/" + svrName + "/")
}

func PickEndpoint(metadata *router.Metadata, svrName string) *router.Endpoint {
	for {
		ep := GetBalancer(svrName).Pick(metadata)
		if ep != nil {
			return ep
		}
	}
}

func PickEndpointByServerPerformance(svrName string, rule etcd.PickRule) *router.Endpoint {
	return etcd.PickEndpoint(svrName, rule)
}

func SetEndpoint(key string, endpoint *router.Endpoint, timeout time.Duration) {
	nettoolkit.SetEndpoint(key, endpoint, timeout)
}

type RetryFunc func() error
type DealFunc func()

func CallWithRetry(retry RetryFunc, deal DealFunc) {
	for {
		err := retry()
		if err != nil {
			log.Debug().Msg(err.Error())
			time.Sleep(time.Second)
			deal()
		} else {
			break
		}
	}
}

func FlushCache(key string, svrName string) {
	GetBalancer(svrName).GetCache().Delete(key)
}
