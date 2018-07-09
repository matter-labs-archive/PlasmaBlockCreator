package foundationdb

import (
	"encoding/binary"
	"errors"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	commonConst "github.com/bankex/go-plasma/common"
	transaction "github.com/bankex/go-plasma/transaction"
	"github.com/bankex/go-plasma/types"
	common "github.com/ethereum/go-ethereum/common"
)

type UTXOlister struct {
	db *fdb.Database
}

func NewUTXOlister(db *fdb.Database) *UTXOlister {
	reader := &UTXOlister{db: db}
	return reader
}

type utxoRange struct {
	beginingKey []byte
	endingKey   []byte
}

// func (r utxoRange) FDBRangeKeySelectors() (beginig, end fdb.Selectable) {
// 	rng := fdb.SelectorRange
// 	return fdb.Selectable(r.beginingKey), fdb.Selectable(r.endingKey)
// }

func newUtxoRange(address common.Address, afterBlock uint32, afterTransaction uint32, afterOutput uint8) (*utxoRange, error) {
	r := &utxoRange{}
	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, afterBlock)
	transactionNumberBuffer := make([]byte, transaction.TransactionNumberLength)
	binary.BigEndian.PutUint32(transactionNumberBuffer, afterTransaction)
	outputNumberBuffer := make([]byte, transaction.OutputNumberLength)
	outputNumberBuffer[0] = afterOutput
	valueBuffer := make([]byte, transaction.ValueLength)
	key := []byte{}
	key = append(key, address[:]...)
	key = append(key, blockNumberBuffer...)
	key = append(key, transactionNumberBuffer...)
	key = append(key, outputNumberBuffer...)
	key = append(key, valueBuffer...)

	addressBN := types.NewBigInt(0)
	addressBN.SetBytes(address[:])
	addressBNNext := types.NewBigInt(0)
	addressBNNext.Bigint.Add(addressBN.Bigint, types.NewBigInt(1).Bigint)
	addressNextBytes := addressBNNext.GetBytes()
	endingKey := []byte{}
	// endingKey = append(endingKey, address[:]...)
	endingKey = append(endingKey, addressNextBytes...)
	endingBlockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	endingTransactionNumberBuffer := make([]byte, transaction.TransactionNumberLength)
	endingOutputNumberBuffer := make([]byte, transaction.OutputNumberLength)
	endingValueBuffer := make([]byte, transaction.ValueLength)
	// for i := range endingBlockNumberBuffer {
	// 	endingBlockNumberBuffer[i] = byte(0xff)
	// }
	// for i := range endingTransactionNumberBuffer {
	// 	endingTransactionNumberBuffer[i] = byte(0xff)
	// }
	// for i := range endingOutputNumberBuffer {
	// 	endingOutputNumberBuffer[i] = byte(0xff)
	// }
	// for i := range endingValueBuffer {
	// 	endingValueBuffer[i] = byte(0xff)
	// }
	endingKey = append(endingKey, endingBlockNumberBuffer...)
	endingKey = append(endingKey, endingTransactionNumberBuffer...)
	endingKey = append(endingKey, endingOutputNumberBuffer...)
	endingKey = append(endingKey, endingValueBuffer...)
	r.beginingKey = key
	r.endingKey = endingKey
	return r, nil
}

func (r *UTXOlister) GetUTXOsForAddress(address common.Address, afterBlock uint32, afterTransaction uint32, afterOutput uint8, limit int) ([][transaction.UTXOIndexLength]byte, error) {
	options := fdb.RangeOptions{}
	options.Limit = limit
	readingRange, err := newUtxoRange(address, afterBlock, afterTransaction, afterOutput)
	if err != nil {
		return nil, err
	}
	pr, err := fdb.PrefixRange([]byte{})
	if err != nil {
		return nil, err
	}
	fullBeginingIndex := []byte{}
	fullBeginingIndex = append(fullBeginingIndex, commonConst.UtxoIndexPrefix...)
	fullBeginingIndex = append(fullBeginingIndex, readingRange.beginingKey...)

	fullEndingIndex := []byte{}
	fullEndingIndex = append(fullEndingIndex, commonConst.UtxoIndexPrefix...)
	fullEndingIndex = append(fullEndingIndex, readingRange.endingKey...)

	pr.Begin = fdb.Key(fullBeginingIndex)
	pr.End = fdb.Key(fullEndingIndex)
	ret, err := r.db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		values, err := tr.GetRange(pr, options).GetSliceWithError()
		if err != nil {
			return nil, err
		}
		return values, nil
	})
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return nil, errors.New("Could not read utxos")
	}
	values := ret.([]fdb.KeyValue)
	toReturn := [][transaction.UTXOIndexLength]byte{}
	expenctedKeyLength := len(commonConst.UtxoIndexPrefix) + transaction.UTXOIndexLength
	toCutFromKey := len(commonConst.UtxoIndexPrefix)
	for _, kv := range values {
		key := kv.Key
		value := kv.Value
		if len(value) != 1 || value[0] != commonConst.UTXOisReadyForSpending || len(key) != expenctedKeyLength {
			continue
		}
		index := [transaction.UTXOIndexLength]byte{}
		keyMeaningfulPart := key[toCutFromKey:]
		copy(index[:], keyMeaningfulPart)
		toReturn = append(toReturn, index)
	}
	return toReturn, nil
}
