package solver

import (
	merror "git.garena.com/shopee/bg-logistics/algo/wms/waveopt/localsearch/pkg/error"
	"math/rand"
	"time"
)

//var GlobalSolverRand *rand.Rand
var GlobalSolverRand *ConcurrentSafeRand

func InitReproducibleSolverRand(seed int64) {
	//GlobalSolverRand = rand.New(rand.NewSource(seed))
	GlobalSolverRand, _ = NewConcurrentSafeRand(100, seed)
}

func InitTrueSolverRand() {
	GlobalSolverRand, _ = NewConcurrentSafeRand(100, time.Now().UnixNano())
}

type ConcurrentSafeRand struct {
	rands chan *rand.Rand
}

func (csr *ConcurrentSafeRand) Intn(n int) int {
	r := <-csr.rands
	defer func() {
		csr.rands <- r
	}()
	return r.Intn(n)
}

func (csr *ConcurrentSafeRand) Float64() float64 {
	r := <-csr.rands
	defer func() {
		csr.rands <- r
	}()

	return r.Float64()
}

func NewConcurrentSafeRand(n, seed int64) (*ConcurrentSafeRand, error) {
	if n <= 1 {
		return nil, merror.New("n must be positive")
	}

	r := &ConcurrentSafeRand{rands: make(chan *rand.Rand, n)}
	for i := 0; i < int(n); i++ {
		newSeed := seed + int64(i)
		r.rands <- rand.New(rand.NewSource(newSeed))
	}

	return r, nil
}
