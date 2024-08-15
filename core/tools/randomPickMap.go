package tools

import (
	"ziruigao/mini-game-router/core/router"
)

type RandomPickMap struct {
	epsMap map[*router.Endpoint]int
	eps    []*router.Endpoint
	rng    *RNG
}

func NewRandomPickMap() *RandomPickMap {
	return &RandomPickMap{
		epsMap: map[*router.Endpoint]int{},
		eps:    []*router.Endpoint{},
		rng:    &RNG{},
	}
}

func (r *RandomPickMap) Add(ep *router.Endpoint) {
	if _, exists := r.epsMap[ep]; exists {
		return
	}
	r.epsMap[ep] = len(r.eps)
	r.eps = append(r.eps, ep)
}

func (r *RandomPickMap) Remove(ep *router.Endpoint) {
	index, exists := r.epsMap[ep]
	if !exists {
		return
	}

	last := len(r.eps) - 1
	r.eps[index] = r.eps[last]
	r.epsMap[r.eps[last]] = index

	r.eps = r.eps[:last]
	delete(r.epsMap, ep)
}

func (r *RandomPickMap) Reset() {
	r.epsMap = map[*router.Endpoint]int{}
	r.eps = []*router.Endpoint{}
}

func (r *RandomPickMap) GetAll() []*router.Endpoint {
	return r.eps
}

func (r *RandomPickMap) GetLast() *router.Endpoint {
	return r.eps[len(r.eps)-1]
}

func (r *RandomPickMap) Len() int {
	return len(r.eps)
}

func (r *RandomPickMap) RandomPick() *router.Endpoint {
	if len(r.eps) == 0 {
		return nil
	}
	return r.eps[r.rng.Uint32n(uint32(len(r.eps)))]
}

func (r *RandomPickMap) Contains(ep *router.Endpoint) bool {
	_, exists := r.epsMap[ep]
	return exists
}
