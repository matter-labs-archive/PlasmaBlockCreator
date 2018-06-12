package foundationdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	indexes "github.com/bankex/go-plasma/indexes"
	transaction "github.com/bankex/go-plasma/transaction"
)

type UTXOwriter struct {
	db *fdb.Database
}

func NewUTXOwriter(db *fdb.Database) *UTXOwriter {
	reader := &UTXOwriter{db: db}
	return reader
}

func (r *UTXOwriter) WriteTX(tx *transaction.SignedTransaction, int64 counter) error {
	if counter < 0 {
		return errors.New("Invalid counter")
	}
	numInputs := len(tx.UnsignedTransaction.Inputs)
	utxosToCheck := make([][UTXOIndexLength]byte, numInputs)
	outputIndexes := make([][indexes.UTXOIndexLength]byte, numOutputs)
	for i := 0; i < numInputs; i++ {
		idx := []byte{utxoIndexPrefix}
		index, err := indexes.CreateCorrespondingUTXOIndexForInput(tx, i)
		if err != nil {
			return err
		}
		idx = append(idx, index[:]...)
		outputIndexes[i] = index
		utxosToCheck[i] = idx
	}

	record := indexes.NewSpendingRecord(tx, outputIndexes)
	var b bytes.Buffer
	i := io.Writer(&b)
	err := record.EncodeRLP(i)
	if err != nil {
		return err
	}
	spendingRecordRaw := b.Bytes()

	counterBuffer := make([]byte, 8)
	err := binary.BigEndian.PutInt64(counterBuffer, counter)
	if err != nil {
		return err
	}

	transactionIndex := []byte{transactionIndexPrefix}
	transactionIndex = append(transactionIndex, counterBuffer)

	ret, err := r.db.Transact(func(tr fdb.Transaction) (bool, error) {
		for _, index := range utxoIndexes {
			existing, err := tr.Get(fdb.Key(index)).Get()
			if err != nil || len(existing) != 0 {
				return false, err
			}
		}
		for _, index := range utxoIndexes {
			tr.Clear(fdb.Key(index))
		}
		tr.Set(fdb.Key(transactionIndex), spendingRecordRaw)
		existing, err := tr.Get(fdb.Key(transactionIndex)).Get()
		if err != nil || len(existing) != 0 || existing != spendingRecordRaw {
			tr.Reset()
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	if !ret {
		return errors.New("Could not write a transaction")
	}
	return nil
}
