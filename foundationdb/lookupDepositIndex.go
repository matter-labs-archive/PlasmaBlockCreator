package foundationdb

import (
	"errors"
	"fmt"

	"github.com/matterinc/PlasmaCommons/transaction"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	commonConst "github.com/matterinc/PlasmaCommons/common"
	types "github.com/matterinc/PlasmaCommons/types"
)

type DepositLookupResult struct {
	BlockNumber       int
	TransactionNumber int
}

func LookupDepositIndex(db *fdb.Database, index *types.BigInt) (*DepositLookupResult, error) {
	depositIndexKey := []byte{}
	depositIndexKey = append(depositIndexKey, commonConst.DepositHistoryPrefix...)
	depositIndexBytes, err := index.GetLeftPaddedBytes(32)
	if err != nil {
		return nil, err
	}
	depositIndexKey = append(depositIndexKey, depositIndexBytes...)

	result, err := db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		existing := tr.Get(fdb.Key(depositIndexKey)).MustGet()
		return existing, nil
	})
	if err != nil {
		fmt.Println("Failed to lookup the deposit index")
		return nil, err
	}
	value := result.([]byte)
	if len(value) == 0 {
		return nil, errors.New("Not yet processed")
	}
	blockNumber, transactionNumber, _, err := transaction.ParseUTXOnumber(value)
	if err != nil {
		return nil, errors.New("Invalid history record")
	}
	res := &DepositLookupResult{int(blockNumber), int(transactionNumber)}
	return res, nil

}
