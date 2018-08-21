package foundationdb

import (
	"encoding/binary"
	"errors"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	commonConst "github.com/shamatar/go-plasma/common"
)

func GetMaxTransactionCounter(db *fdb.Database) (uint64, error) {
	prefix := commonConst.TransactionIndexPrefix

	maxNumberBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(maxNumberBuffer, ^uint64(0))

	fullPrefix := []byte{}
	fullPrefix = append(fullPrefix, prefix...)
	fullPrefix = append(fullPrefix, maxNumberBuffer...)

	keySelector := fdb.LastLessOrEqual(fdb.Key(fullPrefix))

	// options := fdb.RangeOptions{}
	// options.Mode = fdb.StreamingMode(fdb.StreamingModeExact)

	ret, err := db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		values, err := tr.GetKey(keySelector).Get()
		if err != nil {
			return nil, err
		}
		return values, nil
	})
	if err != nil {
		return uint64(0), err
	}
	if ret == nil {
		return uint64(0), nil
	}
	key := ret.(fdb.Key)
	if len(key) == 0 {
		return uint64(0), nil
	}
	slice := key[len(prefix):]
	if len(slice) != 8 {
		return uint64(0), errors.New("Key length is invalid")
	}
	maxCounter := binary.BigEndian.Uint64(slice)
	return maxCounter, nil
}
