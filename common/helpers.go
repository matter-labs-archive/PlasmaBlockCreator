package common

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func CreatePersonalHash(message []byte) common.Hash {
	personalHashData := []byte{}
	personalHashData = append(personalHashData, []byte("\x19Ethereum Signed Message:\n")...)
	personalHashData = append(personalHashData, []byte(strconv.Itoa(len(message)))...)
	personalHashData = append(personalHashData, message...)
	hash := crypto.Keccak256Hash(personalHashData)
	return hash
}
