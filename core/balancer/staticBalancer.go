package mybalancer

import (
	"sync"
	"ziruigao/mini-game-router/core/cache"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/router"
	"ziruigao/mini-game-router/core/tools"

	"github.com/rs/zerolog/log"
)

type StaticBalancer struct {
	randomPickMap *tools.RandomPickMap
	pointerTable  map[string]*router.Endpoint
	mu            sync.RWMutex
	key           string
	cache         *cache.LRUCache
}

func (r *StaticBalancer) New() MyBalancer {
	return &StaticBalancer{}
}

func (r *StaticBalancer) Init(config *config.BalancerRule) {
	r.randomPickMap = tools.NewRandomPickMap()
	r.mu = sync.RWMutex{}
	r.key = config.StaticConfig.Key
	r.pointerTable = map[string]*router.Endpoint{}
	if config.StaticConfig.Cache {
		r.cache = cache.NewLRUCache(config.StaticConfig.CacheSize)
	} else {
		r.cache = nil
	}
}

func (r *StaticBalancer) Pick(metadata *router.Metadata) *router.Endpoint {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := metadata.Get(r.key)

	if r.cache != nil {
		ep := r.cache.Get(key)
		if ep != nil {
			if r.randomPickMap.Contains(ep) {
				log.Debug().Msg("pick from cache")
				return ep
			} else {
				r.cache.Delete(key)
			}
		}
	}

	var targetEp *router.Endpoint = nil
	for _, ep := range r.randomPickMap.GetAll() {
		if ep.IsWants(key) {
			targetEp = ep
			break
		}
	}
	if r.cache != nil {
		r.cache.Put(key, targetEp)
	}
	log.Debug().Msg("pick from static rule matching")
	return targetEp
}

func (r *StaticBalancer) Add(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pointerTable[ep.ToString()] = ep
	r.randomPickMap.Add(ep)
}

func (r *StaticBalancer) Remove(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ep = r.pointerTable[ep.ToString()]
	delete(r.pointerTable, ep.ToString())

	r.randomPickMap.Remove(ep)
}

func (r *StaticBalancer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.randomPickMap.Reset()
	r.pointerTable = map[string]*router.Endpoint{}
	if r.cache != nil {
		r.cache.Reset()
	}
}

func (r *StaticBalancer) GetAll() []*router.Endpoint {
	return r.randomPickMap.GetAll()
}
