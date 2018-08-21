package common

import (
	"sync/atomic"
)

var counter uint64

const increment uint64 = 1

func GetCounter() uint64 {
	return atomic.AddUint64(&counter, increment)
}
