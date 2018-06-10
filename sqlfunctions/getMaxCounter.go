package sqlfunctions

import (
	sql "database/sql"
)

type MaxCounterReader struct {
	db *sql.DB
}

func NewMaxCounterReader(db *sql.DB) *MaxCounterReader {
	reader := &MaxCounterReader{db: db}
	return reader
}

func (r *MaxCounterReader) GetMaxCounter() (uint64, error) {
	var counter uint64
	transaction, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	err = transaction.QueryRow("SELECT MAX(UTXOcounter) AS c FROM Utxos").Scan(&counter)
	if err != nil {
		_ = transaction.Rollback()
		return 0, err
	}
	err = transaction.Commit()
	if err != nil {
		_ = transaction.Rollback()
		return 0, err
	}
	return counter, nil
}
