package sqlfunctions

import (
	sql "database/sql"
	"errors"

	transaction "github.com/bankex/go-plasma/transaction"
)

type TransactionSpendingWriter struct {
	db *sql.DB
}

func NewTransactionSpendingWriter(db *sql.DB) *TransactionSpendingWriter {
	reader := &TransactionSpendingWriter{db: db}
	return reader
}

func (w *TransactionSpendingWriter) WriteSpending(tx *transaction.SignedTransaction, counter int64) (bool, error) {
	from, err := tx.GetFrom()
	if err != nil {
		return false, err
	}
	fromBytes := from[:]
	if tx.UnsignedTransaction.TransactionType[0] == transaction.TransactionTypeSplit {
		stmt, err := w.db.Prepare("CALL spendSingleUtxo(?, ?, ?, ?, ?)") // rawSpendingTX, utxoNum, sender, amountBuffer, counter
		if err != nil {
			return false, err
		}
		defer stmt.Close()
		input := tx.UnsignedTransaction.Inputs[0]
		num := input.GetReferedUTXO().GetString(10)
		value := input.Value[:]
		_, err = stmt.Exec(tx.RawValue, num, fromBytes, value, counter)
		if err != nil {
			return false, err
		}
		return true, nil
	} else if tx.UnsignedTransaction.TransactionType[0] == transaction.TransactionTypeMerge {
		stmt, err := w.db.Prepare("CALL spendTwoUtxo(?, ?, ?, ?, ?, ?, ?)") // rawSpendingTX, utxoNum0, utxoNum1, sender, amountBuffer0, amountBuffer1, counter
		if err != nil {
			return false, err
		}
		defer stmt.Close()
		input0 := tx.UnsignedTransaction.Inputs[0]
		input1 := tx.UnsignedTransaction.Inputs[1]
		num0 := input0.GetReferedUTXO().GetString(10)
		num1 := input1.GetReferedUTXO().GetString(10)
		value0 := input0.Value[:]
		value1 := input1.Value[:]
		_, err = stmt.Exec(tx.RawValue, num0, num1, fromBytes, value0, value1, counter)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, errors.New("Invalid transaction type")
}
