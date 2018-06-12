package foundationdb

import (
	"errors"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	transaction "github.com/bankex/go-plasma/transaction"
)

type UTXOinserter struct {
	db *fdb.Database
}

func NewUTXOinserter(db *fdb.Database) *UTXOinserter {
	reader := &UTXOinserter{db: db}
	return reader
}

func (r *UTXOinserter) InsertUTXO(tx *transaction.NumberedTransaction, blockNumber uint32) error {
	numOutputs := len(tx.SignedTransaction.UnsignedTransaction.Outputs)
	utxoIndexes := make([][]byte, numOutputs)
	for i := 0; i < numOutputs; i++ {
		utxoIndex, err := index.CreateUTXOIndexForOutput(tx, i, blockNumber)
		if err != nil {
			return err
		}
		fullIndex := []byte{utxoIndexPrefix}
		fullIndex = append(fullIndex, utxoIndexes[:])
		utxoIndexes[i] = fullIndex
	}

	ret, err := r.db.Transact(func(tr fdb.Transaction) (bool, error) {
		for _, index := range utxoIndexes {
			existing, err := tr.Get(fdb.Key(index)).Get()
			if err != nil || len(existing) != 0 {
				return false, err
			}
		}
		for _, index := range utxoIndexes {
			tr.Set(fdb.Key(index), UTXOisReadyForSpending)
		}
		for _, index := range utxoIndexes {
			existing, err := tr.Get(fdb.Key(index)).Get()
			if err != nil || len(existing) != 1 || existing != UTXOisReadyForSpending {
				tr.Reset()
				return false, err
			}
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
