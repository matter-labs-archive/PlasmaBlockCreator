package foundationdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bankex/go-plasma/block"
	commonConst "github.com/bankex/go-plasma/common"
	transaction "github.com/bankex/go-plasma/transaction"
	hashmap "github.com/cornelk/hashmap"
)

const blockSliceLengthToWrite = 10000

type BlockWriter struct {
	db *fdb.Database
}

func NewBlockWriter(db *fdb.Database) *BlockWriter {
	reader := &BlockWriter{db: db}
	return reader
}

func (r *BlockWriter) WriteBlock(block block.Block) error {
	blockNumber := binary.BigEndian.Uint32(block.BlockHeader.BlockNumber[:])
	if blockNumber == 0 {
		return errors.New("Invalid block number")
	}
	lastBlockNumber, err := GetLastWrittenBlock(r.db)
	if err != nil {
		return errors.New("Failed to get last written block")
	}
	if lastBlockNumber >= blockNumber {
		return nil
	}
	if blockNumber > lastBlockNumber+1 {
		return errors.New("Writing out of order")
	}
	utxosToWrite := [][]byte{}
	inputLookupHashmap := &hashmap.HashMap{}
	outputLookupHashmap := &hashmap.HashMap{}
	for _, tx := range block.Transactions {
		if tx.SignedTransaction.UnsignedTransaction.TransactionType[0] != transaction.TransactionTypeFund {
			for _, input := range tx.SignedTransaction.UnsignedTransaction.Inputs {
				key := input.GetReferedUTXO().GetBytes()
				val, _ := inputLookupHashmap.Get(key)
				if val == nil {
					inputLookupHashmap.Set(key, []byte{0x01})
				} else {
					return errors.New("Potential doublespend")
				}
			}
		}
		for j := range tx.SignedTransaction.UnsignedTransaction.Outputs {
			key, err := transaction.CreateShortUTXOIndexForOutput(tx, j, blockNumber)
			if err != nil {
				return errors.New("Transaction numbering is incorrect")
			}
			val, _ := outputLookupHashmap.Get(key)
			if val == nil {
				outputLookupHashmap.Set(key, []byte{0x01})
			} else {
				return errors.New("Transaction numbering is incorrect")
			}
			fullIndex, err := transaction.CreateUTXOIndexForOutput(tx, j, blockNumber)
			if err != nil {
				return err
			}
			// fmt.Println(common.ToHex(fullIndex[:]))
			prefixedIndex := []byte{}
			prefixedIndex = append(prefixedIndex, commonConst.UtxoIndexPrefix...)
			prefixedIndex = append(prefixedIndex, fullIndex[:]...)
			utxosToWrite = append(utxosToWrite, prefixedIndex)
		}
	}
	fmt.Println("Writing " + strconv.Itoa(len(utxosToWrite)) + " transactions")
	totalWritten := 0
	for i := 0; i <= len(utxosToWrite)/blockSliceLengthToWrite; i++ {
		currentSlice := [][]byte{}
		minTxNumber := uint32(0)
		maxTxNumber := uint32(0)
		if (i+1)*blockSliceLengthToWrite < len(utxosToWrite) {
			currentSlice = utxosToWrite[i*blockSliceLengthToWrite : (i+1)*blockSliceLengthToWrite]
			minTxNumber = uint32(i * blockSliceLengthToWrite)
			maxTxNumber = uint32((i+1)*blockSliceLengthToWrite) - 1
		} else {
			currentSlice = utxosToWrite[i*blockSliceLengthToWrite:]
			minTxNumber = uint32(i * blockSliceLengthToWrite)
			maxTxNumber = uint32(len(utxosToWrite)) - 1
		}
		err := r.writeSlice(currentSlice, blockNumber, minTxNumber, maxTxNumber)
		if err != nil {
			return err
		}
		totalWritten += len(currentSlice)
	}
	fmt.Println("Has written " + strconv.Itoa(totalWritten) + " transactions")
	_, err = r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		tr.Set(fdb.Key(commonConst.BlockNumberKey), block.BlockHeader.BlockNumber[:])
		updateValue, err := tr.Get(fdb.Key(commonConst.BlockNumberKey)).Get()
		if err != nil {
			return nil, err
		}
		if bytes.Compare(updateValue, block.BlockHeader.BlockNumber[:]) != 0 {
			return nil, errors.New("Failed to write new block number")
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *BlockWriter) writeSlice(slice [][]byte, blockNumber uint32, minTxNumber uint32, maxTxNumber uint32) error {
	bn, txn, err := GetLastWrittenTransactionAndBlock(r.db)
	if err != nil {
		return err
	}
	if !(bn == blockNumber || bn+1 == blockNumber) {
		return errors.New("Writing invalid block number")
	}
	if bn == blockNumber && maxTxNumber < txn {
		return nil
	}
	canWrite := false
	if minTxNumber == 0 && bn+1 == blockNumber {
		canWrite = true
	} else if minTxNumber == txn+1 && bn == blockNumber {
		canWrite = true
	}
	if !canWrite {
		return errors.New("Can not write")
	}

	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)

	transactionNumberBuffer := make([]byte, transaction.TransactionNumberLength)
	binary.BigEndian.PutUint32(transactionNumberBuffer, maxTxNumber)

	futureSlices := make([]fdb.FutureByteSlice, len(slice))
	newLastTxIndex := []byte{}
	newLastTxIndex = append(newLastTxIndex, blockNumberBuffer...)
	newLastTxIndex = append(newLastTxIndex, transactionNumberBuffer...)
	_, err = r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		for _, utxo := range slice {
			tr.Set(fdb.Key(utxo), []byte{commonConst.UTXOisReadyForSpending})
		}
		tr.Set(fdb.Key(commonConst.TransactionNumberKey), newLastTxIndex)

		for i, utxo := range slice {
			futureSlices[i] = tr.Get(fdb.Key(utxo))
		}
		futureIndexRec := tr.Get(fdb.Key(commonConst.TransactionNumberKey))
		for i := range slice {
			if bytes.Compare(futureSlices[i].MustGet(), []byte{commonConst.UTXOisReadyForSpending}) != 0 {
				return nil, errors.New("Failed to write TX from block")
			}
		}
		if bytes.Compare(futureIndexRec.MustGet(), newLastTxIndex) != 0 {
			return nil, errors.New("Failed to set last written TX number")
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}
