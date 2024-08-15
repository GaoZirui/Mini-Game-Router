package sdk

import (
	"sync"
	"time"
	mybalancer "ziruigao/mini-game-router/core/balancer"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/etcd"
	nettoolkit "ziruigao/mini-game-router/core/netToolkit"
	myresolver "ziruigao/mini-game-router/core/resolver"
	"ziruigao/mini-game-router/core/router"
)

var (
	onceForGrpc sync.Once
	once        sync.Once
)

func InitForGrpc(config *config.EtcdConfig, namespace string) {
	onceForGrpc.Do(func() {
		nettoolkit.Init(config, namespace)
		etcd.Init(nettoolkit.GetEtcdClient(), true)
		myresolver.Init()
		mybalancer.Init()
		mybalancer.InitBalancers()
		mybalancer.InitRegistry()
	})
}

func Init(config *config.EtcdConfig, namespace string) {
	once.Do(func() {
		nettoolkit.Init(config, namespace)
		etcd.Init(nettoolkit.GetEtcdClient(), true)
		mybalancer.InitBalancers()
		mybalancer.InitRegistry()
	})
}

func RegisterAndStart(opts *etcd.ServiceRegisterOpts) (*etcd.ServiceRegister, error) {
	return etcd.NewServiceRegister(opts)
}

func Discovery(svrName string) chan struct{} {
	return etcd.WatchPrefix(nettoolkit.GetNamespace(), svrName)
}

func PickEndpoint(metadata *router.Metadata, svrName string) *router.Endpoint {
	for {
		ep := mybalancer.GetBalancer(nettoolkit.GetNamespace() + "/" + svrName + "/").Pick(metadata)
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

func FlushCache(svrName string) {

}

type RetryFunc func() error

func CallWithRetry(fn RetryFunc) {
	for {
		err := fn()
		if err != nil {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}
