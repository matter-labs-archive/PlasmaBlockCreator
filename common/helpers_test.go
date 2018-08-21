package common

import (
	"fmt"
	"testing"
	"time"

	common "github.com/ethereum/go-ethereum/common"
)

func TestInputCreation(t *testing.T) {
	data := [10]byte{}
	hash := CreatePersonalHash(data[:])
	if "0xc5d4c2cc9e8a390b084ba30ed2acea32fe46b6e3d86a480369a9f95021344409" != common.ToHex(hash[:]) {
		panic("Hash mismatch")
	}
}

func BenchmarkAtomicCounter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		start := time.Now()
		counter = 0
		for j := 0; j < 1000000; j++ {
			_ = GetCounter()
		}
		elapsed := time.Since(start)
		txSpeed := float64(1000000) / elapsed.Seconds()
		fmt.Println("Counter speed = " + fmt.Sprintf("%f", txSpeed))
	}
}
