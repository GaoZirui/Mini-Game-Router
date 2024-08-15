package mybalancer

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
	"ziruigao/mini-game-router/core/cache"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/router"

	"github.com/rs/zerolog/log"
)

type hash func(data []byte) uint32

type ConsistentHashBalancer struct {
	hash         hash
	replicas     int
	ring         []int
	nodes        map[int]*router.Endpoint
	mu           sync.RWMutex
	trueNodes    map[*router.Endpoint]struct{}
	key          string
	cache        *cache.LRUCache
	pointerTable map[string]*router.Endpoint
	hit          int64
	miss         int64
}

func (c *ConsistentHashBalancer) New() MyBalancer {
	return &ConsistentHashBalancer{}
}

func (c *ConsistentHashBalancer) Init(config *config.BalancerRule) {
	conf := config.ConsistentHashConfig
	switch conf.HashFunc {
	case "crc32":
		c.hash = crc32.ChecksumIEEE
	default:
		c.hash = crc32.ChecksumIEEE
	}
	c.replicas = conf.Replicas
	c.ring = []int{}
	c.nodes = map[int]*router.Endpoint{}
	c.mu = sync.RWMutex{}
	c.trueNodes = map[*router.Endpoint]struct{}{}
	c.key = conf.Key
	c.pointerTable = map[string]*router.Endpoint{}
	if conf.Cache {
		c.cache = cache.NewLRUCache(conf.CacheSize)
	} else {
		c.cache = nil
	}
}

func (c *ConsistentHashBalancer) Pick(metadata *router.Metadata) *router.Endpoint {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := metadata.Get(c.key)

	if c.cache != nil {
		ep := c.cache.Get(key)
		if ep != nil {
			if _, exists := c.trueNodes[ep]; exists {
				log.Debug().Msg("pick from cache")
				// atomic.AddInt64(&c.hit, 1)
				return ep
			} else {
				c.cache.Delete(key)
			}
		}
	}

	hash := int(c.hash([]byte(key)))

	idx := sort.Search(len(c.ring), func(i int) bool { return c.ring[i] >= hash })

	if idx == len(c.ring) {
		idx = 0
	}

	if len(c.nodes) == 0 {
		return nil
	}

	ep := c.nodes[c.ring[idx]]
	if c.cache != nil {
		c.cache.Put(key, ep)
	}
	log.Debug().Msg(fmt.Sprintf("pick from consistent-hash-calculate: %v %v", key, ep.ToAddr()))
	// atomic.AddInt64(&c.miss, 1)
	return ep
}

func (c *ConsistentHashBalancer) Add(ep *router.Endpoint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.trueNodes[ep] = struct{}{}
	for i := 0; i < c.replicas; i++ {
		hash := int(c.hash([]byte(ep.ToString() + "@" + strconv.Itoa(i))))
		c.ring = append(c.ring, hash)
		c.nodes[hash] = ep
	}
	sort.Ints(c.ring)

	c.pointerTable[ep.ToString()] = ep
}

func (c *ConsistentHashBalancer) Remove(ep *router.Endpoint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ep = c.pointerTable[ep.ToString()]
	delete(c.pointerTable, ep.ToString())

	delete(c.trueNodes, ep)
	for i := 0; i < c.replicas; i++ {
		hash := int(c.hash([]byte(ep.ToString() + "@" + strconv.Itoa(i))))
		delete(c.nodes, hash)
	}

	newRing := []int{}
	for _, h := range c.ring {
		if _, ok := c.nodes[h]; !ok {
			continue
		}
		newRing = append(newRing, h)
	}
	c.ring = newRing
}

func (c *ConsistentHashBalancer) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ring = []int{}
	c.nodes = map[int]*router.Endpoint{}
	c.trueNodes = map[*router.Endpoint]struct{}{}
	c.pointerTable = map[string]*router.Endpoint{}
	if c.cache != nil {
		c.cache.Reset()
	}
}

func (c *ConsistentHashBalancer) GetAll() []*router.Endpoint {
	keys := make([]*router.Endpoint, 0, len(c.trueNodes))
	for r := range c.trueNodes {
		keys = append(keys, r)
	}
	return keys
}

func (c *ConsistentHashBalancer) Rate() {
	fmt.Printf("hit rate: %v miss rate %v", float64(c.hit)/float64(c.hit+c.miss), float64(c.miss)/float64(c.hit+c.miss))
}
