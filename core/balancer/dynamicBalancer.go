package mybalancer

import (
	"context"
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
	cache         cache.Cache
	hit           int64
	miss          int64
	cancel        context.CancelFunc
}

func (r *DynamicBalancer) Name() string {
	return "dynamic"
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
		r.cache = cache.CacheFactory(config.DynamicConfig.CacheType, config.DynamicConfig.CacheSize)
		var ctx context.Context
		ctx, r.cancel = context.WithCancel(context.Background())
		go nettoolkit.Subscribe(ctx, strings.Split(r.key, "-")[0], r.cache)
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
	if ep == nil {
		return nil
	}
	e, exists := r.pointerTable[ep.ToAddr()]
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

	r.pointerTable[ep.ToAddr()] = ep
	r.randomPickMap.Add(ep)
}

func (r *DynamicBalancer) Remove(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ep = r.pointerTable[ep.ToAddr()]
	delete(r.pointerTable, ep.ToAddr())
	r.randomPickMap.Remove(ep)
}

func (r *DynamicBalancer) GetAll() []*router.Endpoint {
	return r.randomPickMap.GetAll()
}

func (r *DynamicBalancer) Stop() {
	r.cancel()
}

func (r *DynamicBalancer) Rate() {
	fmt.Printf("hit rate: %v miss rate %v", float64(r.hit)/float64(r.hit+r.miss), float64(r.miss)/float64(r.hit+r.miss))
}

func (r *DynamicBalancer) GetCache() cache.Cache {
	return r.cache
}
