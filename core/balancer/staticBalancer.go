package mybalancer

import (
	"sync"
	"ziruigao/mini-game-router/core/cache"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/router"
	"ziruigao/mini-game-router/core/tools"
)

type StaticBalancer struct {
	randomPickMap *tools.RandomPickMap
	mu            sync.RWMutex
	key           string
}

func (r *StaticBalancer) Name() string {
	return "static"
}

func (r *StaticBalancer) New() MyBalancer {
	return &StaticBalancer{}
}

func (r *StaticBalancer) Init(config *config.BalancerRule) {
	r.randomPickMap = tools.NewRandomPickMap()
	r.mu = sync.RWMutex{}
	r.key = config.StaticConfig.Key
}

func (r *StaticBalancer) Pick(metadata *router.Metadata) *router.Endpoint {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := metadata.Get(r.key)

	var targetEp *router.Endpoint = nil
	for _, ep := range r.randomPickMap.GetAll() {
		if ep.IsWants(key) {
			targetEp = ep
			break
		}
	}
	return targetEp
}

func (r *StaticBalancer) Add(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.randomPickMap.Add(ep)
}

func (r *StaticBalancer) Remove(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.randomPickMap.Remove(ep)
}

func (r *StaticBalancer) GetAll() []*router.Endpoint {
	return r.randomPickMap.GetAll()
}

func (r *StaticBalancer) Stop() {

}

func (r *StaticBalancer) GetCache() cache.Cache {
	return nil
}
