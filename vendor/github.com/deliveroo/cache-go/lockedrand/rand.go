package lockedrand

import (
	"math/rand"
	"sync"
)

// New returns a new rand.Rand, which is safe for concurrent use.
func New(src rand.Source) *Rand {
	return &Rand{
		mu:   sync.Mutex{},
		rand: rand.New(src),
	}
}

// Rand wraps rand.Rand with a mutex.
type Rand struct {
	mu   sync.Mutex
	rand *rand.Rand
}

// Float64 calls r.Rand.Float64 and is safe for concurrent use.
func (r *Rand) Float64() float64 {
	r.mu.Lock()
	v := r.rand.Float64()
	r.mu.Unlock()
	return v
}
