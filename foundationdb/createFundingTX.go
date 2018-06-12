package foundationdb

import (
	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	transaction "github.com/bankex/go-plasma/transaction"
	indexes "github.com/bankex/go-plasma/indexes"
	types "github.com/bankex/go-plasma/types"
	common "github.com/ethereum/go-ethereum/common"
)

type FundingTXcreator struct {
	db *fdb.Database,
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
func (r *UTXOreader) CreateFundingTX(to common.Address, 
		amount *types.BigInt, 
		counter int64, 
		depositIndex *types.BigInt) error {
	if counter < 0 {
		return errors.New("Invalid counter")
	}
	index := []byte{}
	index = append(index, depositIndexPrefix...)
	index = append(index, depositIndex.GetBytes())
	
	counterBuffer := make([]byte, 8)
	err := binary.BigEndian.PutInt64(counterBuffer, counter)
	if err != nil {
		return err
	}

	spendingRecordRaw := []byte{} // SpendingRecord.RLPencode()

	transactionIndex := []byte{transactionIndexPrefix}
	transactionIndex = append(transactionIndex, counterBuffer)
	transactionIndex = append(transactionIndex, spendingRecordRaw)

	ret, err := r.db.Transact(func(tr fdb.Transaction) (bool, error) {
		existing, err := tr.Get(fdb.Key(index[:])).Get()
		if err != nil || len(existing) != 0 {
			return false, err
		}
		existing, err := tr.Get(fdb.Key(transactionIndex)).Get()
		if err != nil || len(existing) != 0 {
			return false, err
		}
		tr.Set(fdb.Key(transactionIndex), spendingRecordRaw)
		existing, err := tr.Get(fdb.Key(transactionIndex)).Get()
		if err != nil || len(existing) != len(spendingRecordRaw) || existing != spendingRecordRaw {
			tr.Reset()
			return false, err
		}
		return true, nil
	})
	if err != nil{
		return err
	}
	if !ret {
		return errors.New("Could not write a transaction")
	}
	return nil
}
