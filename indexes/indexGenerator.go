package indexes

import (
	"encoding/binary"
	"errors"
	"io"

	transaction "github.com/bankex/go-plasma/transaction"
	rlp "github.com/ethereum/go-ethereum/rlp"
)

const (
	UTXOIndexLength = transaction.AddressLength + transaction.BlockNumberLength + transaction.TransactionNumberLength +
		transaction.OutputNumberLength + transaction.ValueLength
)

type SpendingRecord struct {
	SpendingTransaction *transaction.SignedTransaction
	OutputIndexes       [][UTXOIndexLength]byte
}

func NewSpendingRecord(tx *transaction.SignedTransaction, outputIndexes [][UTXOIndexLength]byte) *SpendingRecord {
	record := &SpendingRecord{}
	record.SpendingTransaction = tx
	record.OutputIndexes = outputIndexes
	return record
}

type rlpSpendingRecord struct {
	SpendingTransaction transaction.SignedTransaction
	OutputIndexes       [][]byte
}

func (tx *SpendingRecord) EncodeRLP(w io.Writer) error {
	switch tx.SpendingTransaction.UnsignedTransaction.TransactionType[0] {
	case transaction.TransactionTypeFund:
		break
	case transaction.TransactionTypeSplit:
		break
	case transaction.TransactionTypeMerge:
		break
	default:
		return errors.New("Invalid transaction type")
	}

	rlpSpending := rlpSpendingRecord{}
	rlpSpending.SpendingTransaction = *tx.SpendingTransaction
	lenSpent := len(tx.SpendingTransaction.UnsignedTransaction.Inputs)
	spentUTXOs := make([][]byte, lenSpent)
	for i := 0; i < lenSpent; i++ {
		index, err := CreateCorrespondingUTXOIndexForInput(tx.SpendingTransaction, i)
		if err != nil {
			return err
		}
		spentUTXOs[i] = index[:]
	}
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
	tx.SpendingTransaction = &dec.SpendingTransaction
	return nil
}

func CreateCorrespondingUTXOIndexForInput(tx *transaction.SignedTransaction, inputNumber int) ([UTXOIndexLength]byte, error) {
	if inputNumber > len(tx.UnsignedTransaction.Inputs) {
		return [UTXOIndexLength]byte{}, errors.New("Invalid input number")
	}
	input := tx.UnsignedTransaction.Inputs[inputNumber]
	from, err := tx.GetFrom()
	if err != nil {
		return [UTXOIndexLength]byte{}, err
	}
	index := []byte{}
	index = append(index, from[:]...)
	index = append(index, input.BlockNumber[:]...)
	index = append(index, input.TransactionNumber[:]...)
	index = append(index, input.OutputNumber[:]...)
	index = append(index, input.Value[:]...)
	if len(index) != UTXOIndexLength {
		return [UTXOIndexLength]byte{}, errors.New("Index length mismatch")
	}
	indexCopy := [UTXOIndexLength]byte{}
	copy(indexCopy[:], index)
	return indexCopy, nil
}

func CreateUTXOIndexForOutput(tx *transaction.NumberedTransaction, outputNumber int, blockNumber uint32) ([UTXOIndexLength]byte, error) {
	if outputNumber > len(tx.SignedTransaction.UnsignedTransaction.Outputs) {
		return [UTXOIndexLength]byte{}, errors.New("Invalid output number")
	}
	output := tx.SignedTransaction.UnsignedTransaction.Outputs[outputNumber]
	index := []byte{}

	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)

	transactionNumberBuffer := tx.TransactionNumber

	index = append(index, output.To[:]...)
	index = append(index, blockNumberBuffer...)
	index = append(index, transactionNumberBuffer[:]...)
	index = append(index, output.OutputNumber[:]...)
	index = append(index, output.Value[:]...)
	if len(index) != UTXOIndexLength {
		return [UTXOIndexLength]byte{}, errors.New("Index length mismatch")
	}
	indexCopy := [UTXOIndexLength]byte{}
	copy(indexCopy[:], index)
	return indexCopy, nil
}
