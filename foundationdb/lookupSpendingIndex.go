package foundationdb

import (
	"encoding/binary"
	"fmt"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	commonConst "github.com/matterinc/PlasmaCommons/common"
	"github.com/matterinc/PlasmaCommons/transaction"
	types "github.com/matterinc/PlasmaCommons/types"
)

type SpendingLookupResult struct {
	BlockNumber       int
	TransactionNumber int
	InputNumber       int
}

func LookupSpendingIndex(db *fdb.Database, index *types.BigInt) (*SpendingLookupResult, error) {
	details, err := transaction.ParseUTXOindexNumberIntoDetails(index)
	if err != nil {
		return nil, err
	}

	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, details.BlockNumber)
	transactionNumberBuffer := make([]byte, transaction.TransactionNumberLength)
	binary.BigEndian.PutUint32(transactionNumberBuffer, details.TransactionNumber)
	outputNumberBuffer := make([]byte, transaction.OutputNumberLength)
	outputNumberBuffer[0] = details.OutputNumber

	lookupIndex := []byte{}
	lookupIndex = append(lookupIndex, commonConst.SpendingIndexKey...)
	lookupIndex = append(lookupIndex, blockNumberBuffer[:]...)
	lookupIndex = append(lookupIndex, transactionNumberBuffer[:]...)
	lookupIndex = append(lookupIndex, outputNumberBuffer[:]...)
	result, err := db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		existing := tr.Get(fdb.Key(lookupIndex)).MustGet()
		return existing, nil
	})
	if err != nil {
		fmt.Println("Failed to lookup the spending index")
		return nil, err
	}
	value := result.([]byte)
	blockNumber, transactionNumber, inputNumber, err := transaction.ParseUTXOnumber(value)
	if err != nil {
		fmt.Println("Failed to lookup the spending index")
		return nil, err
	}
	res := &SpendingLookupResult{int(blockNumber), int(transactionNumber), int(inputNumber)}
	return res, nil

}
