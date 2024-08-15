package mybalancer

import (
	"fmt"
	"strings"
	"sync"
	"ziruigao/mini-game-router/core/cache"
	"ziruigao/mini-game-router/core/config"
	nettoolkit "ziruigao/mini-game-router/core/netToolkit"
	"ziruigao/mini-game-router/core/router"
	"ziruigao/mini-game-router/core/tools"

	"github.com/rs/zerolog/log"
)

type DynamicBalancer struct {
	key           string
	pointerTable  map[string]*router.Endpoint
	mu            sync.RWMutex
	randomPickMap *tools.RandomPickMap
	cache         *cache.LRUCache
	hit           int64
	miss          int64
}

func (r *DynamicBalancer) New() MyBalancer {
	return &DynamicBalancer{}
}

func (r *DynamicBalancer) Init(config *config.BalancerRule) {
	r.key = config.DynamicConfig.Key
	r.pointerTable = map[string]*router.Endpoint{}
	r.mu = sync.RWMutex{}
	r.randomPickMap = tools.NewRandomPickMap()
	if config.DynamicConfig.Cache {
		r.cache = cache.NewLRUCache(config.DynamicConfig.CacheSize)
		go nettoolkit.Subscribe(strings.Split(r.key, "-")[0], r.cache)
	} else {
		r.cache = nil
	}
	r.hit = 0
	r.miss = 0
}

func (r *DynamicBalancer) Pick(metadata *router.Metadata) *router.Endpoint {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := metadata.Get(r.key)

	if r.cache != nil {
		e := r.cache.Get(key)
		if e != nil {
			if r.randomPickMap.Contains(e) {
				log.Debug().Msg("pick from cahche")
				// atomic.AddInt64(&r.hit, 1)
				return e
			} else {
				r.cache.Delete(key)
			}
		}
	}

	ep := nettoolkit.GetEndpoint(key)
	if ep == "" {
		return nil
	}
	e, exists := r.pointerTable[ep]
	if !exists {
		return nil
	}
	if r.cache != nil {
		r.cache.Put(key, e)
	}
	log.Debug().Msg("pick from redis")
	// atomic.AddInt64(&r.miss, 1)
	return e
}

func (r *DynamicBalancer) Add(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.pointerTable[ep.ToString()]; !exists {
		r.pointerTable[ep.ToString()] = ep
		r.randomPickMap.Add(ep)
	}
}

func (r *DynamicBalancer) Remove(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ep = r.pointerTable[ep.ToString()]
	delete(r.pointerTable, ep.ToString())
	r.randomPickMap.Remove(ep)
}

func (r *DynamicBalancer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pointerTable = map[string]*router.Endpoint{}
	r.randomPickMap.Reset()
}

func (r *DynamicBalancer) GetAll() []*router.Endpoint {
	return r.randomPickMap.GetAll()
}

func (r *DynamicBalancer) Rate() {
	fmt.Printf("hit rate: %v miss rate %v", float64(r.hit)/float64(r.hit+r.miss), float64(r.miss)/float64(r.hit+r.miss))
}
