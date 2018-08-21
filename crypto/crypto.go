package crypto

import (
	sha3 "github.com/shamatar/go-plasma/crypto/sha3"
	"github.com/ethereum/go-ethereum/common"
)

var EmptyAddress = common.Address{}

func PubkeyToAddress(pubBytes []byte) common.Address {
	if len(pubBytes) == 65 {
		return common.BytesToAddress(sha3.Keccak256(pubBytes[1:])[12:])
	}
	if len(pubBytes) == 64 {
		return common.BytesToAddress(sha3.Keccak256(pubBytes[:])[12:])

	}
	return EmptyAddress
}
