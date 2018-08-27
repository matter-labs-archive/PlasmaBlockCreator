package common

import (
	"strconv"

	sha3 "github.com/matterinc/PlasmaCommons/crypto/sha3"
	"github.com/ethereum/go-ethereum/common"
)

func CreatePersonalHash(message []byte) common.Hash {
	personalHashData := []byte{}
	personalHashData = append(personalHashData, []byte("\x19Ethereum Signed Message:\n")...)
	personalHashData = append(personalHashData, []byte(strconv.Itoa(len(message)))...)
	personalHashData = append(personalHashData, message...)
	hash := sha3.Keccak256Hash(personalHashData)
	return hash
}
