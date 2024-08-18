package mybalancer

import (
	"sync"
	"ziruigao/mini-game-router/core/cache"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/router"
	"ziruigao/mini-game-router/core/tools"
)

type RandomBalancer struct {
	randomPickMap *tools.RandomPickMap
	mu            sync.RWMutex
}

func (r *RandomBalancer) Name() string {
	return "random"
}

func (r *RandomBalancer) New() MyBalancer {
	return &RandomBalancer{}
}

func (r *RandomBalancer) Init(config *config.BalancerRule) {
	r.randomPickMap = tools.NewRandomPickMap()
	r.mu = sync.RWMutex{}
}

func (r *RandomBalancer) Pick(metadata *router.Metadata) *router.Endpoint {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.randomPickMap.RandomPick()
}

func (r *RandomBalancer) Add(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.randomPickMap.Add(ep)
}

func (r *RandomBalancer) Remove(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.randomPickMap.Remove(ep)
}

func (r *RandomBalancer) GetAll() []*router.Endpoint {
	return r.randomPickMap.GetAll()
}

func (r *RandomBalancer) Stop() {

}

func (r *RandomBalancer) GetCache() cache.Cache {
	return nil
}
