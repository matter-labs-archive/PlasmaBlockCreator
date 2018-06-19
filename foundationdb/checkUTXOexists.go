package foundationdb

import (
	"errors"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	transaction "github.com/bankex/go-plasma/transaction"
)

type UTXOReader struct {
	db *fdb.Database
}

func NewUTXOReader(db *fdb.Database) *UTXOReader {
	reader := &UTXOReader{db: db}
	return reader
}

func (r *UTXOReader) CheckIfUTXOsExist(tx *transaction.SignedTransaction) error {
	if tx.UnsignedTransaction.TransactionType[0] == transaction.TransactionTypeFund {
		return nil
	}
	numInputs := len(tx.UnsignedTransaction.Inputs)
	utxosToCheck := make([][]byte, numInputs)
	for i := 0; i < numInputs; i++ {
		idx := []byte{}
		idx = append(idx, utxoIndexPrefix...)
		index, err := transaction.CreateCorrespondingUTXOIndexForInput(tx, i)
		if err != nil {
			return err
		}
		idx = append(idx, index[:]...)
		utxosToCheck[i] = idx
	}
	_, err := r.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		for _, index := range utxosToCheck {
			status, err := tr.Snapshot().Get(fdb.Key(index)).Get()
			if err != nil {
				return nil, err
			}
			if len(status) != 1 || status[0] != UTXOisReadyForSpending {
				return nil, errors.New("UTXO doesn't exist or invalid")
			}
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}
