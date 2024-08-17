package etcd

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	mybalancer "ziruigao/mini-game-router/core/balancer"
	"ziruigao/mini-game-router/core/metrics"
	nettoolkit "ziruigao/mini-game-router/core/netToolkit"
	"ziruigao/mini-game-router/core/router"

	"github.com/rs/zerolog/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ServiceDiscovery 服务发现
type ServiceDiscovery struct {
	serverList map[string]*sync.Map
	client     *clientv3.Client
}

type EtcdEvent int

const (
	EtcdEvent_Add EtcdEvent = iota
	EtcdEvent_Update
	EtcdEvent_Delete
	EtcdEvent_UpdateBalancer
)

var serviceDiscovery *ServiceDiscovery

// NewServiceDiscovery  新建发现服务
func Init(client *clientv3.Client) {
	serviceDiscovery = &ServiceDiscovery{
		serverList: map[string]*sync.Map{},
		client:     client,
	}
}

func WatchPrefix(namespace, svrName string) chan EtcdEvent {
	prefix := namespace + "/" + svrName + "/"
	resp, err := serviceDiscovery.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	if _, exists := serviceDiscovery.serverList[prefix]; !exists {
		serviceDiscovery.serverList[prefix] = &sync.Map{}
		mybalancer.SetBalancer(prefix, mybalancer.LoadBalancer(svrName))
	}

	for _, ev := range resp.Kvs {
		SetServiceList(prefix, string(ev.Key), router.ParseEndpoint(string(ev.Value)))
	}
	notify := make(chan EtcdEvent, 10000)
	// 监视前缀，修改变更的server
	go watcher(prefix, notify)
	// 监视均衡负载器变更
	go watchBalancerRule(namespace, svrName, notify)
	return notify
}

// watcher 监听前缀
func watcher(prefix string, notify chan EtcdEvent) {
	rch := serviceDiscovery.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Info().Msg(fmt.Sprintf("watching prefix: %s now...\n", prefix))
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: //修改或者新增
				ep := router.ParseEndpoint(string(ev.Kv.Value))
				_, exists := serviceDiscovery.serverList[prefix]
				SetServiceList(prefix, string(ev.Kv.Key), ep)
				if exists {
					notify <- EtcdEvent_Update
					log.Debug().Msg(fmt.Sprintf("update server to: %v\n", ep.ToString()))
				} else {
					notify <- EtcdEvent_Add
				}
			case mvccpb.DELETE: //删除
				DelServiceList(prefix, string(ev.Kv.Key))
				notify <- EtcdEvent_Delete
			}
		}
	}
}

func watchBalancerRule(namespace, svrName string, notify chan EtcdEvent) {
	prefix := namespace + "/" + svrName + "/"
	rch := serviceDiscovery.client.Watch(context.Background(), "config/"+namespace+"/"+svrName, clientv3.WithPrefix())
	log.Info().Msg(fmt.Sprintf("watching prefix: %s now...\n", "config/"+namespace+"/"+svrName))
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: //修改或者新增
				newBalancer := mybalancer.LoadBalancer(svrName)
				for _, ep := range mybalancer.GetBalancer(prefix).GetAll() {
					newBalancer.Add(ep)
				}
				mybalancer.SetBalancer(prefix, newBalancer)
				log.Info().Msg(fmt.Sprintf("change to balancer type: %v\n", mybalancer.GetBalancer(prefix).Name()))
				notify <- EtcdEvent_UpdateBalancer
			case mvccpb.DELETE: //删除
				log.Fatal().Msg("invalid oprations")
			}
		}
	}
}

// SetServiceList 新增服务地址
func SetServiceList(prefix, key string, ep *router.Endpoint) {
	serviceDiscovery.serverList[prefix].Store(key, ep)
	mybalancer.GetBalancer(prefix).Add(ep)
	log.Debug().Msg(fmt.Sprintf("put key: %v val: %v", key, ep.ToString()))
}

// DelServiceList 删除服务地址
func DelServiceList(prefix, key string) {
	e, _ := serviceDiscovery.serverList[prefix].LoadAndDelete(key)
	ep, _ := e.(*router.Endpoint)
	mybalancer.GetBalancer(prefix).Remove(ep)
	log.Debug().Msg(fmt.Sprintf("del key: %v", key))
}

func GetEndpoints(prefix string) []*router.Endpoint {
	if mybalancer.GetBalancer(prefix) != nil {
		return mybalancer.GetBalancer(prefix).GetAll()
	}
	addrs := make([]*router.Endpoint, 0, 10)
	serviceDiscovery.serverList[prefix].Range(func(k, v interface{}) bool {
		addrs = append(addrs, v.(*router.Endpoint))
		return true
	})
	return addrs
}

func PickEndpoint(svrName string, pickRule PickRule) *router.Endpoint {
	prefix := nettoolkit.GetNamespace() + "/" + svrName + "/"
	if _, exists := serviceDiscovery.serverList[prefix]; !exists {
		serviceDiscovery.serverList[prefix] = &sync.Map{}
		mybalancer.SetBalancer(prefix, mybalancer.LoadBalancer(svrName))
		resp, err := serviceDiscovery.client.Get(context.Background(), prefix, clientv3.WithPrefix())
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		for _, ev := range resp.Kvs {
			SetServiceList(prefix, string(ev.Key), router.ParseEndpoint(string(ev.Value)))
		}
	}

	switch pickRule {
	case Least_Gorutine:
		return FindMinEndpoint(GetEndpoints(prefix), CompareByGorutine)
	case Least_Cpu_Percent:
		return FindMinEndpoint(GetEndpoints(prefix), CompareByCpuPercent)
	case Least_Mem_Percent:
		return FindMinEndpoint(GetEndpoints(prefix), CompareByMemPercent)
	case Least_Disk_Percent:
		return FindMinEndpoint(GetEndpoints(prefix), CompareByDiskPercent)
	case Least_Request_In_Duration:
		return FindMinEndpoint(GetEndpoints(prefix), CompareByRequestInDuration)
	case Random:
		eps := GetEndpoints(prefix)
		if len(eps) == 0 {
			return nil
		}
		return eps[rand.Intn(len(eps))]
	default:
		return nil
	}
}

type PickRule int

const (
	Least_Gorutine PickRule = iota
	Least_Cpu_Percent
	Least_Mem_Percent
	Least_Disk_Percent
	Least_Request_In_Duration
	Random
)

func FindMinEndpoint(eps []*router.Endpoint, comp Comparator) *router.Endpoint {
	if len(eps) == 0 {
		return nil
	}
	min := nettoolkit.GetServerPerformance(eps[0])
	fmt.Printf("port %v, sp: %+v\n", eps[0].Port, *min)
	pos := 0
	for i := 1; i < len(eps); i++ {
		sp := nettoolkit.GetServerPerformance(eps[i])
		fmt.Printf("port %v, sp: %+v\n", eps[i].Port, *sp)
		if comp(sp, min) {
			min = sp
			pos = i
		}
	}
	return eps[pos]
}

type Comparator func(o1, o2 *metrics.ServerPerformance) bool

func CompareByGorutine(o1, o2 *metrics.ServerPerformance) bool {
	return o1.NumGoroutine < o2.NumGoroutine
}

func CompareByCpuPercent(o1, o2 *metrics.ServerPerformance) bool {
	return o1.CpuPercent < o2.CpuPercent
}

func CompareByMemPercent(o1, o2 *metrics.ServerPerformance) bool {
	return o1.MemPercent < o2.MemPercent
}

func CompareByDiskPercent(o1, o2 *metrics.ServerPerformance) bool {
	return o1.DiskPercent < o2.DiskPercent
}

func CompareByRequestInDuration(o1, o2 *metrics.ServerPerformance) bool {
	return o1.RequestInDuration < o2.RequestInDuration
}
