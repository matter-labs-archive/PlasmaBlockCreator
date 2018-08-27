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

// func (r *UTXOWriter) WriteSpending(tx *transaction.SignedTransaction, counter uint64) error {
// 	numInputs := len(tx.UnsignedTransaction.Inputs)
// 	utxosToCheck := make([]subspace.Subspace, numInputs)
// 	// outputIndexes := make([][transaction.UTXOIndexLength]byte, numInputs)
// 	for i := 0; i < numInputs; i++ {
// 		// idx := []byte(utxoIndexPrefix)
// 		sub, err := transaction.CreateFdbUTXOIndexForInput(*r.db, tx, i)
// 		if err != nil {
// 			return err
// 		}
// 		utxosToCheck[i] = sub
// 		// idx = append(idx, index[:]...)
// 		// outputIndexes[i] = index
// 		// utxosToCheck[i] = idx
// 	}

// 	// record := transaction.NewSpendingRecord(tx, outputIndexes)
// 	// record := transaction.NewSpendingRecord(tx, [][transaction.UTXOIndexLength]byte{})
// 	// var b bytes.Buffer
// 	// i := io.Writer(&b)
// 	// err := record.EncodeRLP(i)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// spendingRecordRaw := b.Bytes()
// 	// transactionIndex := CreateTransactionIndex(counter)
// 	_, err := r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
// 		for _, index := range utxosToCheck {
// 			existing, err := tr.Get(index).Get()
// 			if err != nil {
// 				return nil, err
// 			}
// 			if len(existing) != 1 {
// 				return nil, errors.New("No such UTXO")
// 			}
// 		}
// 		// existing, err := tr.Get(fdb.Key(transactionIndex)).Get()
// 		// if err != nil {
// 		// 	return nil, err
// 		// }
// 		// if len(existing) != 0 {
// 		// 	return nil, errors.New("Double spend")
// 		// }
// 		for _, index := range utxosToCheck {
// 			tr.Clear(index)
// 		}
// 		// tr.Set(fdb.Key(transactionIndex), spendingRecordRaw)
// 		// existing, err = tr.Get(fdb.Key(transactionIndex)).Get()
// 		// if err != nil {
// 		// 	tr.Reset()
// 		// 	return nil, err
// 		// }
// 		// if len(existing) == 0 || bytes.Compare(existing, spendingRecordRaw) != 0 {
// 		// 	tr.Reset()
// 		// 	return nil, errors.New("Reading mismatch")
// 		// }
// 		return nil, nil
// 	})
// 	if err != nil {
// 		// log.Println("Did not write")
// 		return err
// 	}
// 	// _, err = r.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
// 	// 	for _, index := range utxosToCheck {
// 	// 		existing, err := tr.Get(fdb.Key(index)).Get()
// 	// 		if err != nil {
// 	// 			return nil, err
// 	// 		}
// 	// 		if len(existing) != 0 {
// 	// 			return nil, errors.New("Did not pass reading after writing check")
// 	// 		}
// 	// 	}
// 	// 	existing, err := tr.Get(fdb.Key(transactionIndex)).Get()
// 	// 	if err != nil {
// 	// 		return nil, err
// 	// 	}
// 	// 	if len(existing) == 0 {
// 	// 		return nil, errors.New("Did not pass reading after writing check")
// 	// 	}
// 	// 	return nil, nil
// 	// })
// 	// if err != nil {
// 	// 	// log.Println("Did not pass reading after writing check")
// 	// 	return err
// 	// }

// 	return nil
// }

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
	// _, err = r.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
	// 	for _, index := range utxosToCheck {
	// 		existing, err := tr.Get(fdb.Key(index)).Get()
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		if len(existing) != 0 {
	// 			return nil, errors.New("Did not pass reading after writing check")
	// 		}
	// 	}
	// 	existing, err := tr.Get(fdb.Key(transactionIndex)).Get()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if len(existing) == 0 {
	// 		return nil, errors.New("Did not pass reading after writing check")
	// 	}
	// 	return nil, nil
	// })
	// if err != nil {
	// 	// log.Println("Did not pass reading after writing check")
	// 	return err
	// }

	return nil
}

// func (r *UTXOWriter) WriteSpending(res *transaction.ParsedTransactionResult, counter uint64) error {
// 	r.concurrencyChannel <- true
// 	defer func() { <-r.concurrencyChannel }()
// 	transactionIndex := CreateTransactionIndex(counter)
// 	shardID := res.ShardID
// 	futureSlices := make([]fdb.FutureByteSlice, len(res.UtxoIndexes))
// 	_, err := sharding.Transact(shardID, func(tr fdb.Transaction) (interface{}, error) {
// 		tr.AddWriteConflictKey(fdb.Key(transactionIndex))
// 		for i, utxoIndex := range res.UtxoIndexes {
// 			futureSlices[i] = tr.Get(fdb.Key(utxoIndex.Key))
// 		}
// 		futureTxRec := tr.Get(fdb.Key(transactionIndex))
// 		for i, utxoIndex := range res.UtxoIndexes {
// 			if bytes.Compare(futureSlices[i].MustGet(), utxoIndex.Value) != 0 {
// 				return nil, errors.New("Double spend")
// 			}
// 		}
// 		if len(futureTxRec.MustGet()) != 0 {
// 			return nil, errors.New("Such transaction alreadt exists")
// 		}
// 		for _, utxoIndex := range res.UtxoIndexes {
// 			tr.Clear(fdb.Key(utxoIndex.Key))
// 		}
// 		tr.ByteMax(fdb.Key(transactionIndex), res.SpendingRecord)
// 		return nil, nil
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
