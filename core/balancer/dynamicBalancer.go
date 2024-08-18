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
	mu            sync.RWMutex
	randomPickMap *tools.RandomPickMap
	cache         cache.Cache
	hit           int64
	miss          int64
	cancel        context.CancelFunc
	autoFlush     bool
}

func (r *DynamicBalancer) Name() string {
	return "dynamic"
}

func (r *DynamicBalancer) New() MyBalancer {
	return &DynamicBalancer{}
}

func (r *DynamicBalancer) Init(config *config.BalancerRule) {
	r.key = config.DynamicConfig.Key
	r.mu = sync.RWMutex{}
	r.randomPickMap = tools.NewRandomPickMap()
	r.autoFlush = config.DynamicConfig.AutoFlush
	if config.DynamicConfig.Cache {
		r.cache = cache.CacheFactory(config.DynamicConfig.CacheType, config.DynamicConfig.CacheSize)
		if r.autoFlush {
			var ctx context.Context
			ctx, r.cancel = context.WithCancel(context.Background())
			go nettoolkit.Subscribe(ctx, strings.Split(r.key, "-")[0], r.cache)
		} else {
			r.cancel = nil
		}
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
			if exists, _ := r.randomPickMap.Contains(e); exists {
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
	exists, e := r.randomPickMap.Contains(ep)
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

	r.randomPickMap.Add(ep)
}

func (r *DynamicBalancer) Remove(ep *router.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.randomPickMap.Remove(ep)
}

func (r *DynamicBalancer) GetAll() []*router.Endpoint {
	return r.randomPickMap.GetAll()
}

func (r *DynamicBalancer) Stop() {
	if r.autoFlush {
		r.cancel()
	}
}

func (r *DynamicBalancer) Rate() {
	fmt.Printf("hit rate: %v miss rate %v", float64(r.hit)/float64(r.hit+r.miss), float64(r.miss)/float64(r.hit+r.miss))
}

func (r *DynamicBalancer) GetCache() cache.Cache {
	return r.cache
}
