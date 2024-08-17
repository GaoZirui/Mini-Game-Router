package cache

import (
	"ziruigao/mini-game-router/core/router"

	"github.com/rs/zerolog/log"
)

type Cache interface {
	New(int) Cache
	Get(string) *router.Endpoint
	Put(string, *router.Endpoint)
	Delete(string)
	Reset()
	Name() string
}

var (
	cacheRegistry map[string]Cache
)

func RegisterCache(name string, cache Cache) {
	cacheRegistry[name] = cache
}

func InitRegistry() {
	cacheRegistry = map[string]Cache{}

	RegisterCache("lru", &LRUCache{})
}

func CacheFactory(name string, capacity int) Cache {
	if cache, exists := cacheRegistry[name]; exists {
		c := cache.New(capacity)
		return c
	}
	log.Fatal().Msg("invalid cache type")
	return nil
}
