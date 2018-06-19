package foundationdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	commonConst "github.com/bankex/go-plasma/common"
	transaction "github.com/bankex/go-plasma/transaction"
	types "github.com/bankex/go-plasma/types"
	common "github.com/ethereum/go-ethereum/common"
)

type TestUTXOcreator struct {
	db *fdb.Database
}

func NewTestUTXOcreator(db *fdb.Database) *TestUTXOcreator {
	reader := &TestUTXOcreator{db: db}
	return reader
}

// func (r *TestUTXOcreator) InsertUTXO(address common.Address, blockNumber uint32, transactionNumber uint32, outputNumber uint8, value *types.BigInt) error {
// 	utxoIndexes := make([]subspace.Subspace, 1)
// 	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
// 	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)
// 	transactionNumberBuffer := make([]byte, transaction.TransactionNumberLength)
// 	binary.BigEndian.PutUint32(transactionNumberBuffer, transactionNumber)
// 	outputNumberBuffer := make([]byte, transaction.OutputNumberLength)
// 	outputNumberBuffer[0] = outputNumber
// 	valueBuffer, err := value.GetLeftPaddedBytes(transaction.ValueLength)
// 	addressDirectory, err := directory.CreateOrOpen(r.db, []string{"utxo"}, nil)
// 	// fmt.Println(common.ToHex(addressDirectory.Bytes()))
// 	fullSubspace := addressDirectory.Sub(tuple.Tuple{address[:]})
// 	// fmt.Println(common.ToHex(fullSubspace.Bytes()))
// 	fullSubspace = fullSubspace.Sub(tuple.Tuple{blockNumberBuffer})
// 	// fmt.Println(common.ToHex(fullSubspace.Bytes()))
// 	fullSubspace = fullSubspace.Sub(tuple.Tuple{transactionNumberBuffer})
// 	// fmt.Println(common.ToHex(fullSubspace.Bytes()))
// 	fullSubspace = fullSubspace.Sub(tuple.Tuple{outputNumberBuffer})
// 	// fmt.Println(common.ToHex(fullSubspace.Bytes()))
// 	fullSubspace = fullSubspace.Sub(tuple.Tuple{valueBuffer})
// 	// fmt.Println(common.ToHex(fullSubspace.Bytes()))
// 	utxoIndexes[0] = fullSubspace
// 	_, err = r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
// 		for _, index := range utxoIndexes {
// 			existing, err := tr.Get(index).Get()
// 			if err != nil {
// 				return nil, err
// 			}
// 			if len(existing) != 0 {
// 				return nil, errors.New("Record already exists")
// 			}
// 		}
// 		for _, index := range utxoIndexes {
// 			tr.Set(index, []byte{UTXOisReadyForSpending})
// 		}
// 		for _, index := range utxoIndexes {
// 			existing, err := tr.Get(index).Get()
// 			if err != nil {
// 				tr.Reset()
// 				return nil, err
// 			}
// 			if len(existing) != 1 || bytes.Compare(existing, []byte{UTXOisReadyForSpending}) != 0 {
// 				tr.Reset()
// 				return nil, errors.New("Reading mismatch")
// 			}
// 		}
// 		return nil, nil
// 	})
// 	if err != nil {
// 		fmt.Print(err)
// 		fmt.Println(common.ToHex(fullSubspace.Bytes()))
// 		return err
// 	}
// 	return nil
// }

func (r *TestUTXOcreator) InsertUTXO(address common.Address, blockNumber uint32, transactionNumber uint32, outputNumber uint8, value *types.BigInt) error {
	utxoIndexes := make([][]byte, 1)
	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)
	transactionNumberBuffer := make([]byte, transaction.TransactionNumberLength)
	binary.BigEndian.PutUint32(transactionNumberBuffer, transactionNumber)
	outputNumberBuffer := make([]byte, transaction.OutputNumberLength)
	outputNumberBuffer[0] = outputNumber
	valueBuffer, err := value.GetLeftPaddedBytes(transaction.ValueLength)
	key := []byte{}
	key = append(key, commonConst.UtxoIndexPrefix...)
	key = append(key, address[:]...)
	key = append(key, blockNumberBuffer...)
	key = append(key, transactionNumberBuffer...)
	key = append(key, outputNumberBuffer...)
	key = append(key, valueBuffer...)
	utxoIndexes[0] = key
	fmt.Println(common.ToHex(key))
	_, err = r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		for _, index := range utxoIndexes {
			existing, err := tr.Get(fdb.Key(index)).Get()
			if err != nil {
				return nil, err
			}
			if len(existing) != 0 {
				return nil, errors.New("Record already exists")
			}
		}
		for _, index := range utxoIndexes {
			tr.Set(fdb.Key(index), []byte{commonConst.UTXOisReadyForSpending})
		}
		for _, index := range utxoIndexes {
			existing, err := tr.Get(fdb.Key(index)).Get()
			if err != nil {
				tr.Reset()
				return nil, err
			}
			if len(existing) != 1 || bytes.Compare(existing, []byte{commonConst.UTXOisReadyForSpending}) != 0 {
				tr.Reset()
				return nil, errors.New("Reading mismatch")
			}
		}
		return nil, nil
	})
	if err != nil {
		fmt.Print(err)
		fmt.Println(common.ToHex(key))
		return err
	}
	return nil
}
