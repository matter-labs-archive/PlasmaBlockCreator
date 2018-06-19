package foundationdb

import (
	"encoding/binary"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	commonConst "github.com/bankex/go-plasma/common"
)

func CreateTransactionIndex(counter uint64) []byte {
	counterBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBuffer, counter)
	transactionIndex := []byte(commonConst.TransactionIndexPrefix)
	transactionIndex = append(transactionIndex, counterBuffer...)
	return transactionIndex
}

func GetLastWrittenBlock(db *fdb.Database) (uint32, error) {
	ret, err := db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		return tr.Get(fdb.Key(commonConst.BlockNumberKey)).Get()
	})
	if err != nil {
		return 0, err
	}
	// if ret == nil {
	// 	return 0, errors.New("Could not read list writen block")
	// }
	retBytes := ret.([]byte)
	if len(retBytes) == 0 {
		return 0, nil
	}
	lastBlock := binary.BigEndian.Uint32(retBytes)
	return lastBlock, nil
}
