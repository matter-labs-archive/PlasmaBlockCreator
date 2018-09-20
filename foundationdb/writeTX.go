package foundationdb

import (
	"bytes"
	"errors"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/matterinc/PlasmaCommons/transaction"
)

type UTXOWriter struct {
	db                 *fdb.Database
	Concurrency        int
	concurrencyChannel chan bool
}

func NewUTXOWriter(db *fdb.Database, concurrency int) *UTXOWriter {
	c := make(chan bool, concurrency)
	reader := &UTXOWriter{db: db, Concurrency: concurrency, concurrencyChannel: c}
	return reader
}

func (r *UTXOWriter) WriteSpending(res *transaction.ParsedTransactionResult, counter uint64) error {
	r.concurrencyChannel <- true
	defer func() { <-r.concurrencyChannel }()
	transactionIndex := CreateTransactionIndex(counter)
	futureSlices := make([]fdb.FutureByteSlice, len(res.UtxoIndexes))
	// _, err := Transact(func(tr fdb.Transaction) (interface{}, error) {
	_, err := r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		// tr.AddWriteConflictKey(fdb.Key(transactionIndex))
		for i, utxoIndex := range res.UtxoIndexes {
			futureSlices[i] = tr.Get(fdb.Key(utxoIndex.Key))
		}
		futureTxRec := tr.Get(fdb.Key(transactionIndex))
		for i, utxoIndex := range res.UtxoIndexes {
			valueRead := futureSlices[i].MustGet()
			if bytes.Compare(valueRead, utxoIndex.Value) != 0 {
				return nil, errors.New("Double spend")
			}
		}
		if len(futureTxRec.MustGet()) != 0 {
			return nil, errors.New("Such transaction already exists")
		}
		for _, utxoIndex := range res.UtxoIndexes {
			tr.Clear(fdb.Key(utxoIndex.Key))
		}
		tr.Set(fdb.Key(transactionIndex), res.SpendingRecord)
		// tr.ByteMax(fdb.Key(transactionIndex), res.SpendingRecord)
		return nil, nil
	})
	if err != nil {
		// log.Println("Did not write")
		return err
	}
	return nil
}
