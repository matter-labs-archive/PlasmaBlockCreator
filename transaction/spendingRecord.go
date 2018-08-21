package transaction

import (
	"errors"
	"io"

	"github.com/ethereum/go-ethereum/rlp"
)

type SpendingRecord struct {
	SpendingTransaction *SignedTransaction
	OutputIndexes       [][UTXOIndexLength]byte
}

type rlpSpendingRecord struct {
	SpendingTransaction *SignedTransaction
	OutputIndexes       [][]byte
}

func NewSpendingRecord(tx *SignedTransaction, outputIndexes [][UTXOIndexLength]byte) *SpendingRecord {
	record := &SpendingRecord{}
	record.SpendingTransaction = tx
	record.OutputIndexes = outputIndexes
	return record
}

func (tx *SpendingRecord) EncodeRLP(w io.Writer) error {
	switch tx.SpendingTransaction.UnsignedTransaction.TransactionType[0] {
	case TransactionTypeFund:
		break
	case TransactionTypeSplit:
		break
	case TransactionTypeMerge:
		break
	default:
		return errors.New("Invalid transaction type")
	}

	rlpSpending := rlpSpendingRecord{}
	rlpSpending.SpendingTransaction = tx.SpendingTransaction
	lenSpent := len(tx.SpendingTransaction.UnsignedTransaction.Inputs)
	spentUTXOs := make([][]byte, lenSpent)
	for i := 0; i < lenSpent; i++ {
		index, err := CreateCorrespondingUTXOIndexForInput(tx.SpendingTransaction, i)
		if err != nil {
			return err
		}
		spentUTXOs[i] = index[:]
	}
	rlpSpending.OutputIndexes = spentUTXOs
	return rlp.Encode(w, rlpSpending)
}

func (tx *SpendingRecord) DecodeRLP(s *rlp.Stream) error {
	var dec rlpSpendingRecord
	if err := s.Decode(&dec); err != nil {
		return err
	}
	tx.OutputIndexes = make([][UTXOIndexLength]byte, len(dec.OutputIndexes))
	for i := 0; i < len(dec.OutputIndexes); i++ {
		copy(tx.OutputIndexes[i][:], dec.OutputIndexes[i])
	}
	tx.SpendingTransaction = dec.SpendingTransaction
	return nil
}
