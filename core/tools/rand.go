package tools

import "time"

type RNG struct {
	x uint32
	y uint64
}

func (r *RNG) Uint32() uint32 {
	for r.x == 0 {
		r.x = getRandomUint32()
	}

	// See https://en.wikipedia.org/wiki/Xorshift
	x := r.x
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.x = x
	return x
}

func (r *RNG) Uint32n(maxN uint32) uint32 {
	x := r.Uint32()
	// See http://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return uint32((uint64(x) * uint64(maxN)) >> 32)
}

func (r *RNG) Uint64() uint64 {
	for r.y == 0 {
		r.y = getRandomUint64()
	}

	y := r.y
	y ^= y << 13
	y ^= y >> 7
	y ^= y << 5
	r.y = y
	return y
}
func (r *RNG) Uint64n(maxN uint64) uint64 {
	x := r.Uint64()
	return x % (maxN + 1)
}
func getRandomUint32() uint32 {
	x := time.Now().UnixNano()
	return uint32((x >> 32) ^ x)
}
func getRandomUint64() uint64 {
	x := time.Now().UnixNano()
	return uint64(x)
}
