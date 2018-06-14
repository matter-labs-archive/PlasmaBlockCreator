package foundationdb

import (
	"encoding/binary"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
)

const (
	UTXOexistsButNotSpendable = byte(0x00)
	UTXOisReadyForSpending    = byte(0x01)
)

var (
	depositIndexPrefix     = []byte("deposit")
	utxoIndexPrefix        = []byte("utxo")
	transactionIndexPrefix = []byte("ctr")
	blockNumberKey         = []byte("blockNumber")
)

func CreateTransactionIndex(counter uint64) []byte {
	counterBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBuffer, counter)
	transactionIndex := []byte(transactionIndexPrefix)
	transactionIndex = append(transactionIndex, counterBuffer...)
	return transactionIndex
}

func GetLastWrittenBlock(db *fdb.Database) (uint32, error) {
	ret, err := db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		return tr.Get(fdb.Key(blockNumberKey)).Get()
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
