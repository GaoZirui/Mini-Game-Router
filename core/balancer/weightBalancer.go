package mybalancer

import (
	"math/rand"
	"sync"
	"ziruigao/mini-game-router/core/cache"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/router"
	"ziruigao/mini-game-router/core/tools"
)

type WeightBalancer struct {
	mu            sync.RWMutex
	randomPickMap *tools.RandomPickMap
	totalWeight   int
}

func (r *WeightBalancer) New() MyBalancer {
	return &WeightBalancer{}
}

func (r *WeightBalancer) Name() string {
	return "weight"
}

func (r *WeightBalancer) Init(config *config.BalancerRule) {
	r.mu = sync.RWMutex{}
	r.randomPickMap = tools.NewRandomPickMap()
}

func (r *WeightBalancer) Pick(metadata *router.Metadata) *router.Endpoint {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.randomPickMap.Len() == 0 {
		return nil
	}

	// 生成一个随机数，范围是 [1, r.totalWeight]
	randomNum := rand.Intn(r.totalWeight) + 1

	currentWeight := 0
	for _, ep := range r.randomPickMap.GetAll() {
		currentWeight += ep.Weight
		if randomNum <= currentWeight {
			return ep
		}
	}

	// 如果因为某种原因没有返回端点（理论上不会发生），返回最后一个端点
	return r.randomPickMap.GetLast()
}

func (r *WeightBalancer) Add(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if exists, e := r.randomPickMap.Contains(ep); exists {
		r.totalWeight -= e.Weight
	}

	r.randomPickMap.Add(ep)

	r.totalWeight += ep.Weight
}

func (r *WeightBalancer) Remove(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.randomPickMap.Remove(ep)
	r.totalWeight -= ep.Weight
}

func (r *WeightBalancer) GetAll() []*router.Endpoint {
	return r.randomPickMap.GetAll()
}

func (r *WeightBalancer) Stop() {

}

func (r *WeightBalancer) GetCache() cache.Cache {
	return nil
}
