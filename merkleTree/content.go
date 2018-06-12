package merkletree

import (
	"bytes"

	helpers "github.com/bankex/go-plasma/common"
)

type TransactionContent struct {
	RLPEncodedTransaction []byte
}

func NewTransactionContent(bytes []byte) TransactionContent {
	content := TransactionContent{}
	content.RLPEncodedTransaction = bytes
	return content
}

//CalculateHash hashes the values of a TestContent
func (t TransactionContent) CalculateHash() []byte {
	hash := helpers.CreatePersonalHash(t.RLPEncodedTransaction)
	bytes := hash[:]
	return bytes
}

//Equals tests for equality of two Contents
func (t TransactionContent) Equals(other Content) bool {
	return bytes.Compare(t.RLPEncodedTransaction, other.(TransactionContent).RLPEncodedTransaction) == 0
}
