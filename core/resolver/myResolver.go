package myresolver

import (
	"strings"
	"ziruigao/mini-game-router/core/etcd"
	nettoolkit "ziruigao/mini-game-router/core/netToolkit"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/resolver"
)

const (
	scheme = "grpclb"
)

// MyResolver 服务发现
type MyResolver struct {
	cc resolver.ClientConn
}

// NewMyResolver  新建发现服务
func NewMyResolver() *MyResolver {
	return &MyResolver{}
}

// Build 为给定目标创建一个新的`resolver`，当调用`grpc.Dial()`时执行
func (s *MyResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	s.cc = cc
	prefix := nettoolkit.GetNamespace() + "/" + target.Endpoint() + "/"
	notify := etcd.WatchPrefix(nettoolkit.GetNamespace(), target.Endpoint())
	s.update(prefix)
	go func() {
		for range notify {
			s.update(prefix)
		}
	}()
	return s, nil
}

func (s *MyResolver) update(prefix string) {
	addrs := make([]resolver.Address, 0, 10)
	for _, ep := range etcd.GetEndpoints(prefix) {
		addr := resolver.Address{
			Addr:       ep.ToAddr(),
			ServerName: getServerName(prefix),
		}
		addrs = append(addrs, addr)
	}
	s.cc.UpdateState(resolver.State{
		Addresses: addrs,
	})
}

// ResolveNow 监视目标更新
func (s *MyResolver) ResolveNow(rn resolver.ResolveNowOptions) {
	log.Debug().Msg("grpc resolve now")
}

// Scheme return schema
func (s *MyResolver) Scheme() string {
	return scheme
}

// Close 关闭
func (s *MyResolver) Close() {
	log.Debug().Msg("resolver close")
}

func Init() {
	resolver.Register(NewMyResolver())
}

func getServerName(prefix string) string {
	parts := strings.Split(prefix, "/")
	return parts[1]
}
