package mybalancer

import (
	"fmt"
	"time"
	nettoolkit "ziruigao/mini-game-router/core/netToolkit"
	"ziruigao/mini-game-router/core/router"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const (
	name = "mybalancer"
)

var (
	currentPicker *myPicker
)

type myPickerBuilder struct {
	blcs map[string]MyBalancer
}

func (r *myPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	scs := map[string]balancer.SubConn{}
	firstFlag := false
	var svrName string
	for subConn, addr := range info.ReadySCs {
		if !firstFlag {
			firstFlag = true
			svrName = addr.Address.ServerName
			// 直到连接上全部子连接才允许 Build
			if len(info.ReadySCs) != len(GetBalancer(nettoolkit.GetNamespace()+"/"+svrName+"/").GetAll()) {
				if currentPicker == nil {
					return nil
				} else {
					return currentPicker
				}
			}
		}
		scs[addr.Address.Addr] = subConn
	}
	currentPicker = &myPicker{
		subConns: scs,
		blc:      GetBalancer(nettoolkit.GetNamespace() + "/" + svrName + "/"),
	}
	return currentPicker
}

type myPicker struct {
	subConns map[string]balancer.SubConn
	blc      MyBalancer
}

func (p *myPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	ep := p.blc.Pick(router.NewMetadata(info.Ctx))
	// 还没刷新，等一会儿
	if ep == nil {
		time.Sleep(time.Second)
		return balancer.PickResult{}, nil
	}
	log.Debug().Msg(fmt.Sprintf("pick: %v\n", ep.Port))
	return balancer.PickResult{
		SubConn: p.subConns[ep.ToAddr()],
	}, nil
}

func Init() {
	currentPicker = nil
	balancer.Register(base.NewBalancerBuilder(name, &myPickerBuilder{
		blcs: map[string]MyBalancer{},
	}, base.Config{HealthCheck: true}))
}
