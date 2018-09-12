package foundationdb

import (
	"encoding/binary"
	"errors"
	"fmt"

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
	depositIndexKey = append(depositIndexKey, commonConst.DepositIndexPrefix...)
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
	counter := binary.BigEndian.Uint64(value)
	// TODO may be wrap in helpers here
	blockNumber := counter >> 32
	transactionNumber := counter % (uint64(1) << 32)

	res := &DepositLookupResult{int(blockNumber), int(transactionNumber)}
	return res, nil

}
