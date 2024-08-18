package tools

import (
	"ziruigao/mini-game-router/core/router"
)

type RandomPickMap struct {
	epsMap map[string]int
	eps    []*router.Endpoint
	rng    *RNG
}

func NewRandomPickMap() *RandomPickMap {
	return &RandomPickMap{
		epsMap: map[string]int{},
		eps:    []*router.Endpoint{},
		rng:    &RNG{},
	}
}

func (r *RandomPickMap) Add(ep *router.Endpoint) {
	if e, exists := r.epsMap[ep.ToAddr()]; exists {
		r.eps[e] = ep
		return
	}

	r.epsMap[ep.ToAddr()] = len(r.eps)
	r.eps = append(r.eps, ep)
}

func (r *RandomPickMap) Remove(ep *router.Endpoint) {
	index, exists := r.epsMap[ep.ToAddr()]
	if !exists {
		return
	}

	last := len(r.eps) - 1
	r.eps[index] = r.eps[last]
	r.epsMap[r.eps[last].ToAddr()] = index

	r.eps = r.eps[:last]
	delete(r.epsMap, ep.ToAddr())
}

func (r *RandomPickMap) Reset() {
	r.epsMap = map[string]int{}
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

func (r *RandomPickMap) Contains(ep *router.Endpoint) (bool, *router.Endpoint) {
	index, exists := r.epsMap[ep.ToAddr()]
	if exists {
		return exists, r.eps[index]
	} else {
		return exists, nil
	}
}
