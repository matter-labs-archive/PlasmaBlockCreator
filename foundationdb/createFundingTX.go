package foundationdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	commonConst "github.com/matterinc/PlasmaCommons/common"
	"github.com/matterinc/PlasmaCommons/transaction"
	types "github.com/matterinc/PlasmaCommons/types"
	common "github.com/ethereum/go-ethereum/common"
)

type FundingTXcreator struct {
	db         *fdb.Database
	signingKey []byte
}

func NewFundingTXcreator(db *fdb.Database, signingKey []byte) *FundingTXcreator {
	reader := &FundingTXcreator{db: db, signingKey: signingKey}
	return reader
}

func (r *FundingTXcreator) CreateFundingTX(to common.Address,
	value *types.BigInt,
	counter uint64,
	depositIndex *types.BigInt) error {
	if counter < 0 {
		return errors.New("Invalid counter")
	}

	counterBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBuffer, counter)

	depositIndexKey := []byte{}
	depositIndexKey = append(depositIndexKey, commonConst.DepositIndexPrefix...)
	depositIndexBytes, err := depositIndex.GetLeftPaddedBytes(32)
	if err != nil {
		return err
	}
	depositIndexKey = append(depositIndexKey, depositIndexBytes...)

	fundingTX, err := transaction.CreateRawFundingTX(to, value, depositIndex, r.signingKey)
	if err != nil {
		return err
	}
	err = fundingTX.Validate()
	if err != nil {
		return err
	}
	spendingRecord := transaction.NewSpendingRecord(fundingTX, [][transaction.UTXOIndexLength]byte{})

	var b bytes.Buffer
	i := io.Writer(&b)
	err = spendingRecord.EncodeRLP(i)
	if err != nil {
		return err
	}
	spendingRecordRaw := b.Bytes() // SpendingRecord.RLPencode()

	transactionIndex := CreateTransactionIndex(counter)

	_, err = r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		existing, err := tr.Get(fdb.Key(depositIndexKey)).Get() // check for existing deposit
		if err != nil {
			return nil, err
		}
		if len(existing) != 0 {
			tr.Reset()
			return nil, errors.New("Duplicate funding transaction")
		}
		existing, err = tr.Get(fdb.Key(transactionIndex)).Get()
		if err != nil {
			tr.Reset()
			return nil, err
		}
		if len(existing) != 0 {
			tr.Reset()
			return nil, errors.New("Counter is reused")
		}
		tr.Set(fdb.Key(depositIndexKey), counterBuffer)
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
