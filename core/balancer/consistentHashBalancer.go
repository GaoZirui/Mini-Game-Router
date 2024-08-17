package mybalancer

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
	"ziruigao/mini-game-router/core/cache"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/router"
)

type hash func(data []byte) uint32

type ConsistentHashBalancer struct {
	hash         hash
	replicas     int
	ring         []int
	nodes        map[int]*router.Endpoint
	mu           sync.RWMutex
	key          string
	pointerTable map[string]*router.Endpoint
}

func (c *ConsistentHashBalancer) Name() string {
	return "consistent-hash"
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
	c.key = conf.Key
	c.pointerTable = map[string]*router.Endpoint{}
}

func (c *ConsistentHashBalancer) Pick(metadata *router.Metadata) *router.Endpoint {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := metadata.Get(c.key)

	hash := int(c.hash([]byte(key)))

	idx := sort.Search(len(c.ring), func(i int) bool { return c.ring[i] >= hash })

	if idx == len(c.ring) {
		idx = 0
	}

	if len(c.nodes) == 0 {
		return nil
	}

	ep := c.nodes[c.ring[idx]]
	return ep
}

func (c *ConsistentHashBalancer) Add(ep *router.Endpoint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.pointerTable[ep.ToAddr()]; !exists {
		for i := 0; i < c.replicas; i++ {
			hash := int(c.hash([]byte(ep.ToAddr() + "@" + strconv.Itoa(i))))
			c.ring = append(c.ring, hash)
			c.nodes[hash] = ep
		}
		sort.Ints(c.ring)
	}

	c.pointerTable[ep.ToAddr()] = ep
}

func (c *ConsistentHashBalancer) Remove(ep *router.Endpoint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ep = c.pointerTable[ep.ToAddr()]
	delete(c.pointerTable, ep.ToAddr())

	for i := 0; i < c.replicas; i++ {
		hash := int(c.hash([]byte(ep.ToAddr() + "@" + strconv.Itoa(i))))
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

func (c *ConsistentHashBalancer) GetAll() []*router.Endpoint {
	keys := make([]*router.Endpoint, 0, len(c.pointerTable))
	for _, ep := range c.pointerTable {
		keys = append(keys, ep)
	}
	return keys
}

func (c *ConsistentHashBalancer) Stop() {

}

func (c *ConsistentHashBalancer) GetCache() cache.Cache {
	return nil
}
