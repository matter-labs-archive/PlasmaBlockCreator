package sqlfunctions

import (
	sql "database/sql"

	transaction "github.com/bankex/go-plasma/transaction"
)

type UTXOreader struct {
	db *sql.DB
}

func NewUTXOReader(db *sql.DB) *UTXOreader {
	reader := &UTXOreader{db: db}
	return reader
}

func (r *UTXOreader) CheckIfUTXOsExist(tx *transaction.SignedTransaction) (bool, error) {
	from, err := tx.GetFrom()
	if err != nil {
		return false, err
	}
	fromBytes := from[:]
	stmt, err := r.db.Prepare("CALL selectIfNotSpent(?, ?, ?)") // UTXOnum, from, value
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	// var counter string
	for _, input := range tx.UnsignedTransaction.Inputs {
		num := input.GetReferedUTXO().GetString(10)
		value := input.Value[:]
		_, err = stmt.Exec(num, fromBytes, value)
		if err != nil {
			return false, err
		}
		// err = r.db.QueryRow("CALL selectIfNotSpent(?, ?, ?)", num, fromBytes, value).Scan(&counter)
		// if err != nil {
		// 	return false, err
		// }
	}

	return true, nil
}
