package foundationdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	hashmap "github.com/cornelk/hashmap"
	"github.com/matterinc/PlasmaCommons/block"
	commonConst "github.com/matterinc/PlasmaCommons/common"
	transaction "github.com/matterinc/PlasmaCommons/transaction"
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

	fmt.Println("Writing block number " + strconv.Itoa(int(blockNumber)))

	numberOfTransactionInBlock := len(block.Transactions)
	utxosToWrite := make([][][]byte, numberOfTransactionInBlock)                // [numTxes][someOutputsPerTX][outputBytes]
	spendingHistoriesToWrite := make([][][2][]byte, numberOfTransactionInBlock) //[numTxes][someInputsPerTX][originating, spending][data]
	inputLookupHashmap := hashmap.New(uintptr(numberOfTransactionInBlock))
	outputLookupHashmap := hashmap.New(uintptr(numberOfTransactionInBlock))

	start := time.Now()

	for i, tx := range block.Transactions {
		if tx.UnsignedTransaction.TransactionType[0] != transaction.TransactionTypeFund {
			for _, input := range tx.UnsignedTransaction.Inputs {
				key := input.GetReferedUTXO().GetBytes()
				val, _ := inputLookupHashmap.Get(key)
				if val == nil {
					inputLookupHashmap.Set(key, 1)
				} else {
					return errors.New("Potential doublespend")
				}
			}
		}

		if tx.UnsignedTransaction.TransactionType[0] != transaction.TransactionTypeFund {
			transactionSpendingHistory := [][2][]byte{}
			for k := range tx.UnsignedTransaction.Inputs {
				originatingKey, err := transaction.CreateShortUTXOIndexForInput(tx, k)
				if err != nil {
					return errors.New("Transaction numbering is incorrect")
				}
				spendingKey := transaction.PackUTXOnumber(blockNumber, uint32(i), uint8(k))
				tuple := [2][]byte{}
				tuple[0] = append(commonConst.SpendingIndexKey, originatingKey...)
				tuple[1] = spendingKey
				transactionSpendingHistory = append(transactionSpendingHistory, tuple)
			}
			spendingHistoriesToWrite[i] = transactionSpendingHistory
		}

		transactionNewUTXOs := [][]byte{}
		for j := range tx.UnsignedTransaction.Outputs {
			key, err := transaction.CreateShortUTXOIndexForOutput(tx, blockNumber, uint32(i), j)
			if err != nil {
				return errors.New("Transaction numbering is incorrect")
			}
			val, _ := outputLookupHashmap.Get(key)
			if val == nil {
				outputLookupHashmap.Set(key, 1)
			} else {
				return errors.New("Transaction numbering is incorrect")
			}
			fullIndex, err := transaction.CreateUTXOIndexForOutput(tx, blockNumber, uint32(i), j)
			if err != nil {
				return err
			}
			prefixedIndex := []byte{}
			prefixedIndex = append(prefixedIndex, commonConst.UtxoIndexPrefix...)
			prefixedIndex = append(prefixedIndex, fullIndex[:]...)
			transactionNewUTXOs = append(transactionNewUTXOs, prefixedIndex)
		}
		utxosToWrite[i] = transactionNewUTXOs

	}
	elapsed := time.Since(start)
	fmt.Println("Block writing preparation taken " + fmt.Sprintf("%d", elapsed.Nanoseconds()/1000000) + " ms")

	fmt.Println("Writing " + strconv.Itoa(numberOfTransactionInBlock) + " transactions")
	start = time.Now()

	totalWritten := 0
	numSlices := numberOfTransactionInBlock / blockSliceLengthToWrite
	if numberOfTransactionInBlock%blockSliceLengthToWrite != 0 {
		numSlices++
	}
	for i := 0; i < numSlices; i++ {
		minTxNumber := uint32(0)
		maxTxNumber := uint32(0)
		if (i+1)*blockSliceLengthToWrite < len(utxosToWrite) {
			minTxNumber = uint32(i * blockSliceLengthToWrite)
			maxTxNumber = uint32((i+1)*blockSliceLengthToWrite) - 1
		} else {
			minTxNumber = uint32(i * blockSliceLengthToWrite)
			maxTxNumber = uint32(numberOfTransactionInBlock) - 1
		}
		currentUTXOSlice := utxosToWrite[minTxNumber : maxTxNumber+1]
		currentHistorySlice := spendingHistoriesToWrite[minTxNumber : maxTxNumber+1]
		err := r.writeSlice(currentUTXOSlice, currentHistorySlice, blockNumber, minTxNumber, maxTxNumber)
		if err != nil {
			return err
		}
		totalWritten += int(maxTxNumber - minTxNumber + 1)
	}
	fmt.Println("Has written " + strconv.Itoa(totalWritten) + " transaction for outputs and histories")

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
	elapsed = time.Since(start)
	fmt.Println("Block writing taken " + fmt.Sprintf("%d", elapsed.Nanoseconds()/1000000) + " ms")

	return nil
}

func (r *BlockWriter) writeSlice(utxoSlice [][][]byte, historySlice [][][2][]byte, blockNumber uint32, minTxNumber uint32, maxTxNumber uint32) error {
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

	futureUTXOSlices := []fdb.FutureByteSlice{}
	futureHistorySlices := []fdb.FutureByteSlice{}
	newLastTxIndex := []byte{}
	newLastTxIndex = append(newLastTxIndex, blockNumberBuffer...)
	newLastTxIndex = append(newLastTxIndex, transactionNumberBuffer...)
	// i := 0
	// j := 0

	_, err = r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		for _, transactionUTXOs := range utxoSlice {
			for _, utxo := range transactionUTXOs {
				tr.Set(fdb.Key(utxo), []byte{commonConst.UTXOisReadyForSpending})
				futureUTXOSlices = append(futureUTXOSlices, tr.Get(fdb.Key(utxo)))
			}
		}

		for _, transactionHistories := range historySlice {
			for _, history := range transactionHistories {
				tr.Set(fdb.Key(history[0]), history[1])
				futureHistorySlices = append(futureHistorySlices, tr.Get(fdb.Key(history[0])))
			}
		}

		tr.Set(fdb.Key(commonConst.TransactionNumberKey), newLastTxIndex)
		futureIndexRec := tr.Get(fdb.Key(commonConst.TransactionNumberKey))

		// for _, transactionUTXOs := range utxoSlice {
		// 	for range transactionUTXOs {
		// 		if bytes.Compare(futureUTXOSlices[i].MustGet(), []byte{commonConst.UTXOisReadyForSpending}) != 0 {
		// 			return nil, errors.New("Failed to write UTXO from block")
		// 		}
		// 		i++
		// 	}
		// }

		// for _, transactionHistories := range historySlice {
		// 	for _, history := range transactionHistories {
		// 		if bytes.Compare(futureHistorySlices[j].MustGet(), history[1]) != 0 {
		// 			return nil, errors.New("Failed to write TX history from block")
		// 		}
		// 		j++
		// 	}
		// }

		if bytes.Compare(futureIndexRec.MustGet(), newLastTxIndex) != 0 {
			return nil, errors.New("Failed to set last written TX number")
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	// fmt.Println("Has written " + strconv.Itoa(i) + " outputs")
	// fmt.Println("Has written " + strconv.Itoa(j) + " histories")
	return nil
}
