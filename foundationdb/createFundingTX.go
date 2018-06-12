package foundationdb

import (
	"bytes"
	"encoding/binary"
	"errors"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	transaction "github.com/bankex/go-plasma/transaction"
	types "github.com/bankex/go-plasma/types"
	common "github.com/ethereum/go-ethereum/common"
)

type FundingTXcreator struct {
	db         *fdb.Database
	signingKey []byte
}

func NewFundingTXcreator(db *fdb.Database) *FundingTXcreator {
	reader := &FundingTXcreator{db: db}
	return reader
}

func createRawFundingTX(to common.Address,
	amount *types.BigInt,
	counter int64,
	depositIndex *types.BigInt) (*transaction.SignedTransaction, error) {
	return nil, errors.New("NYI")
}

// ('CALL writeFundingTx(?, ?, ?, ?, ?)', [toAddressBuffer, amountBuffer, counter, rawSpendingTX, depositIndexBN.toString(10)]);
func (r *FundingTXcreator) CreateFundingTX(to common.Address,
	amount *types.BigInt,
	counter uint64,
	depositIndex *types.BigInt) error {
	if counter < 0 {
		return errors.New("Invalid counter")
	}
	index := []byte{}
	index = append(index, depositIndexPrefix...)
	index = append(index, depositIndex.GetBytes()...)

	counterBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBuffer, counter)

	spendingRecordRaw := []byte{} // SpendingRecord.RLPencode()

	transactionIndex := []byte{}
	transactionIndex = append(transactionIndex, transactionIndexPrefix...)
	transactionIndex = append(transactionIndex, counterBuffer...)
	transactionIndex = append(transactionIndex, spendingRecordRaw...)

	_, err := r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		existing, err := tr.Get(fdb.Key(index[:])).Get()
		if err != nil || len(existing) != 0 {
			return nil, err
		}
		existing, err = tr.Get(fdb.Key(transactionIndex)).Get()
		if err != nil || len(existing) != 0 {
			return nil, err
		}
		tr.Set(fdb.Key(transactionIndex), spendingRecordRaw)
		existing, err = tr.Get(fdb.Key(transactionIndex)).Get()
		if err != nil {
			tr.Reset()
			return nil, err
		}
		if len(existing) != len(spendingRecordRaw) || bytes.Compare(existing, spendingRecordRaw) != 0 {
			tr.Reset()
			return nil, errors.New("Reading mismatch")
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}
