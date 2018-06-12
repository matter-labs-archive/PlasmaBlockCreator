package foundationdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	transaction "github.com/bankex/go-plasma/transaction"
)

type UTXOWriter struct {
	db *fdb.Database
}

func NewUTXOWriter(db *fdb.Database) *UTXOWriter {
	reader := &UTXOWriter{db: db}
	return reader
}

func (r *UTXOWriter) WriteSpending(tx *transaction.SignedTransaction, counter uint64) error {
	if counter < 0 {
		return errors.New("Invalid counter")
	}
	numInputs := len(tx.UnsignedTransaction.Inputs)
	utxosToCheck := make([][]byte, numInputs)
	outputIndexes := make([][transaction.UTXOIndexLength]byte, numInputs)
	for i := 0; i < numInputs; i++ {
		idx := []byte(utxoIndexPrefix)
		index, err := transaction.CreateCorrespondingUTXOIndexForInput(tx, i)
		if err != nil {
			return err
		}
		idx = append(idx, index[:]...)
		outputIndexes[i] = index
		utxosToCheck[i] = idx
	}

	record := transaction.NewSpendingRecord(tx, outputIndexes)
	var b bytes.Buffer
	i := io.Writer(&b)
	err := record.EncodeRLP(i)
	if err != nil {
		return err
	}
	spendingRecordRaw := b.Bytes()

	counterBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBuffer, counter)

	transactionIndex := []byte(transactionIndexPrefix)
	transactionIndex = append(transactionIndex, counterBuffer...)
	// runs on isolated snapshot without explicit lock
	_, err = r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		for _, index := range utxosToCheck {
			existing, err := tr.Get(fdb.Key(index)).Get()
			if err != nil {
				return nil, err
			}
			if len(existing) != 1 {
				return nil, errors.New("No such UTXO")
			}
		}
		for _, index := range utxosToCheck {
			tr.Clear(fdb.Key(index))
		}
		existing, err := tr.Get(fdb.Key(transactionIndex)).Get()
		if err != nil {
			tr.Reset()
			return nil, err
		}
		if len(existing) != 0 {
			tr.Reset()
			return nil, errors.New("Double spend")
		}
		tr.Set(fdb.Key(transactionIndex), spendingRecordRaw)
		existing, err = tr.Get(fdb.Key(transactionIndex)).Get()
		if err != nil {
			tr.Reset()
			return nil, err
		}
		if len(existing) == 0 || bytes.Compare(existing, spendingRecordRaw) != 0 {
			tr.Reset()
			return nil, errors.New("Reading mismatch")
		}
		return nil, nil
	})
	if err != nil {
		// log.Println("Did not write")
		return err
	}

	_, err = r.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		for _, index := range utxosToCheck {
			existing, err := tr.Get(fdb.Key(index)).Get()
			if err != nil {
				return nil, err
			}
			if len(existing) != 0 {
				return nil, errors.New("Did not pass reading after writing check")
			}
		}
		existing, err := tr.Get(fdb.Key(transactionIndex)).Get()
		if err != nil {
			return nil, err
		}
		if len(existing) == 0 {
			return nil, errors.New("Did not pass reading after writing check")
		}
		return nil, nil
	})
	if err != nil {
		// log.Println("Did not pass reading after writing check")
		return err
	}

	return nil
}
