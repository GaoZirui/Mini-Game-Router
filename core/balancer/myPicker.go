package mybalancer

import (
	"fmt"
	"strings"
	nettoolkit "ziruigao/mini-game-router/core/netToolkit"
	"ziruigao/mini-game-router/core/router"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

const (
	name = "mybalancer"
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
		ep := parseAddress(addr.Address)
		if !firstFlag {
			firstFlag = true
			svrName = ep.Name
			// 直到连接上全部子连接才允许 Build
			if len(info.ReadySCs) != len(GetBalancer(nettoolkit.GetNamespace()+"/"+svrName+"/").GetAll()) {
				return nil
			}
		}
		scs[ep.ToString()] = subConn
	}
	return &myPicker{
		subConns: scs,
		blc:      GetBalancer(nettoolkit.GetNamespace() + "/" + svrName + "/"),
	}
}

type myPicker struct {
	subConns map[string]balancer.SubConn
	blc      MyBalancer
}

func (p *myPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	ep := p.blc.Pick(router.NewMetadata(info.Ctx))
	// 还没刷新，等一会儿
	if ep == nil {
		return balancer.PickResult{}, nil
	}
	log.Debug().Msg(fmt.Sprintf("pick: %v\n", ep.Port))
	return balancer.PickResult{
		SubConn: p.subConns[ep.ToString()],
	}, nil
}

func Init() {
	balancer.Register(base.NewBalancerBuilder(name, &myPickerBuilder{
		blcs: map[string]MyBalancer{},
	}, base.Config{HealthCheck: true}))
}

func parseAddr(s string) (string, string) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		log.Fatal().Msg("invalid addr")
	}
	return parts[0], parts[1]
}

func parseAddress(a resolver.Address) *router.Endpoint {
	ip, port := parseAddr(a.Addr)
	v := a.Attributes.Value(router.Attribute_Key_For_Weight)
	weight, _ := v.(int)
	v = a.Attributes.Value(router.Attribute_Key_For_Wants)
	w, _ := v.(string)
	wants := router.ParseWants(w)
	v = a.Attributes.Value(router.Attribute_Key_For_Wants_Type)
	wantsType, _ := v.(router.WantsType)
	v = a.Attributes.Value(router.Attribute_Key_For_Namespace)
	namespace, _ := v.(string)
	return &router.Endpoint{
		Name:      a.ServerName,
		Namespace: namespace,
		IP:        ip,
		Port:      port,
		Weight:    weight,
		Wants:     wants,
		WantsType: wantsType,
	}
}
