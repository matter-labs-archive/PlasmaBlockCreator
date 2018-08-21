package merkletree

import (
	"bytes"

	helpers "github.com/shamatar/go-plasma/common"
	"github.com/shamatar/go-plasma/crypto/sha3"
	"github.com/ethereum/go-ethereum/common"
)

var emptyTransactionBytes = common.FromHex("0xf847c300c0c000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000")

type TransactionContent struct {
	RLPEncodedTransaction []byte
}

func NewTransactionContent(bytes []byte) TransactionContent {
	content := TransactionContent{}
	content.RLPEncodedTransaction = bytes
	return content
}

func (t TransactionContent) CalculateHash() []byte {
	hash := helpers.CreatePersonalHash(t.RLPEncodedTransaction)
	bytes := hash[:]
	return bytes
}

func (t TransactionContent) Equals(other Content) bool {
	return bytes.Compare(t.RLPEncodedTransaction, other.(TransactionContent).RLPEncodedTransaction) == 0
}

type PaddingContent struct {
	PaddingElement []byte
}

//CalculateHash hashes the values of a TestContent
func (t PaddingContent) CalculateHash() []byte {
	hash := sha3.Keccak256(t.PaddingElement)
	bytes := hash[:]
	return bytes
}

//Equals tests for equality of two Contents
func (t PaddingContent) Equals(other Content) bool {
	return bytes.Compare(t.PaddingElement, other.(PaddingContent).PaddingElement) == 0
}

func NewPaddingNode() PaddingContent {
	paddingC := PaddingContent{emptyTransactionBytes}
	return paddingC
}
