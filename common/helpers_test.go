package common

import (
	"testing"

	common "github.com/ethereum/go-ethereum/common"
)

func TestInputCreation(t *testing.T) {
	data := [10]byte{}
	hash := CreatePersonalHash(data[:])
	if "0xc5d4c2cc9e8a390b084ba30ed2acea32fe46b6e3d86a480369a9f95021344409" != common.ToHex(hash[:]) {
		panic("Hash mismatch")
	}
}
