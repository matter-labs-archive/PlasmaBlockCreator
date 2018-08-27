package transaction

import (
	"bytes"
	"io"

	commonConst "github.com/matterinc/PlasmaCommons/common"
	rlp "github.com/ethereum/go-ethereum/rlp"
)

type ParsedTransactionResult struct {
	From           []byte
	UtxoIndexes    []UTXOindex
	SpendingRecord []byte
	ShardID        []byte
	TX             SignedTransaction
}

type UTXOindex struct {
	Key   []byte
	Value []byte
}

type TransactionParser struct {
	ConcurrencyLimit   int
	concurrencyChannel chan bool
}

func NewTransactionParser(concurrency int) *TransactionParser {
	ch := make(chan bool, concurrency)
	new := &TransactionParser{concurrency, ch}
	return new
}

func (p *TransactionParser) Parse(raw []byte) (*ParsedTransactionResult, error) {
	p.concurrencyChannel <- true
	defer func() { <-p.concurrencyChannel }()
	tx := &(SignedTransaction{})
	err := rlp.DecodeBytes(raw, tx)
	if err != nil {
		return nil, err
	}
	err = tx.Validate()
	if err != nil {
		return nil, err
	}

	from, err := tx.GetFrom()
	if err != nil {
		return nil, err
	}
	if bytes.Compare(from[:], EmptyAddress[:]) == 0 {
		return nil, err
	}
	numInputs := len(tx.UnsignedTransaction.Inputs)
	utxoIndexes := make([]UTXOindex, numInputs)

	outputIndexes := make([][UTXOIndexLength]byte, numInputs) // specific

	expectedValue := []byte{commonConst.UTXOisReadyForSpending}
	for i := 0; i < numInputs; i++ {
		idx := []byte{}
		idx = append(idx, commonConst.UtxoIndexPrefix...)
		index, err := CreateCorrespondingUTXOIndexForInput(tx, i)
		if err != nil {
			return nil, err
		}
		idx = append(idx, index[:]...)
		utxoIndex := UTXOindex{idx, expectedValue}
		utxoIndexes[i] = utxoIndex
		outputIndexes[i] = index
	}

	record := NewSpendingRecord(tx, outputIndexes)
	var b bytes.Buffer
	i := io.Writer(&b)
	err = record.EncodeRLP(i)
	if err != nil {
		return nil, err
	}
	spendingRecordRaw := b.Bytes()
	shardID := []byte{from[:][0]}
	result := &ParsedTransactionResult{from[:], utxoIndexes, spendingRecordRaw, shardID, *tx}
	return result, nil
}
