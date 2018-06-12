package foundationdb

import (
	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	transaction "github.com/bankex/go-plasma/transaction"
	indexes "github.com/bankex/go-plasma/indexes"
)

type UTXOreader struct {
	db *fdb.Database
}

func NewUTXOReader(db *fdb.Database) *UTXOreader {
	reader := &UTXOreader{db: db}
	return reader
}

func (r *UTXOreader) CheckIfUTXOsExist(tx *transaction.SignedTransaction) error {
	if tx.UnsignedTransaction.TransactionType[0] == transaction.TransactionTypeFund {
		return nil
	}
	numInputs := len(tx.UnsignedTransaction.Inputs)
	utxosToCheck := make ([][]byte, numInputs)
	for i := 0; i < numInputs; i++ {
		idx := []byte{utxoIndexPrefix}
		index, err := indexes.CreateCorrespondingUTXOIndexForInput(tx, i)
		if err != nil {
			return err
		}
		idx = append(idx, index[:]...)
		utxosToCheck[i] = idx
	}
	ret, err := r.db.ReadTransact(func(tr fdb.ReadTransact) (bool, error) {
		for _, index := range utxosToCheck {
			counter, err := tr.Get(fdb.Key(index)).Get()
			if err != nil {
				return false, err
			}
			if len(counter) != 1, counter[0] != UTXOisReadyForSpending {
				return false, nil
			}
		}
	})
	if err != nil || !ret {
		return err
	}
	if !ret {
		return errors.New("Could not write a transaction")
	}
	return nil
}
