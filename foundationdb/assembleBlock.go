package foundationdb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/rlp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bankex/go-plasma/block"
	commonConst "github.com/bankex/go-plasma/common"
	transaction "github.com/bankex/go-plasma/transaction"
	hashmap "github.com/cornelk/hashmap"
	"github.com/go-redis/redis"
)

type BlockAssembler struct {
	db          *fdb.Database
	redisClient *redis.Client
}

func NewBlockAssembler(db *fdb.Database, redisClient *redis.Client) *BlockAssembler {
	reader := &BlockAssembler{db: db, redisClient: redisClient}
	return reader
}

func (r *BlockAssembler) getRecordsForBlock(blockNumber uint32) ([]*transaction.SpendingRecord, error) {
	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)

	nextBlockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(nextBlockNumberBuffer, blockNumber+1)

	pr, err := fdb.PrefixRange([]byte{})
	if err != nil {
		return nil, err
	}

	txNumberPadding := make([]byte, transaction.TransactionNumberLength)

	fullBeginingIndex := []byte{}
	fullBeginingIndex = append(fullBeginingIndex, commonConst.TransactionIndexPrefix...)
	fullBeginingIndex = append(fullBeginingIndex, blockNumberBuffer...)
	fullBeginingIndex = append(fullBeginingIndex, txNumberPadding...)

	fullEndingIndex := []byte{}
	fullEndingIndex = append(fullEndingIndex, commonConst.TransactionIndexPrefix...)
	fullEndingIndex = append(fullEndingIndex, nextBlockNumberBuffer...)
	fullEndingIndex = append(fullEndingIndex, txNumberPadding...)

	pr.Begin = fdb.Key(fullBeginingIndex)
	pr.End = fdb.Key(fullEndingIndex)

	options := fdb.RangeOptions{}
	options.Mode = fdb.StreamingMode(fdb.StreamingModeWantAll)

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
		return nil, errors.New("Could not write a transaction")
	}
	values := ret.([]fdb.KeyValue)
	toReturn := []*transaction.SpendingRecord{}
	expectedKeyLength := len(commonConst.TransactionIndexPrefix) + transaction.BlockNumberLength + transaction.TransactionNumberLength
	for _, kv := range values {
		key := kv.Key
		value := kv.Value
		if len(key) != expectedKeyLength {
			continue
		}
		var newSpendingRecord transaction.SpendingRecord
		err := rlp.DecodeBytes(value, &newSpendingRecord)
		if err != nil {
			return nil, errors.New("Failed to deserialize spending record")
		}
		toReturn = append(toReturn, &newSpendingRecord)
	}
	return toReturn, nil
}

func (r *BlockAssembler) AssembleBlock(newBlockNumber uint32, previousHash []byte, startNext bool) (*block.Block, error) {
	spendingRecords, err := r.getRecordsForBlock(newBlockNumber)
	start := time.Now()
	if err != nil {
		return nil, err
	}
	if !startNext {
		if len(spendingRecords) == 0 {
			return nil, nil
		}
	}
	if len(spendingRecords) == 0 {
		return nil, nil
	}
	newCounterToSet := (uint64(newBlockNumber+1) << (transaction.TransactionNumberLength * 8)) - 1
	keys := make([]string, 1)
	keys[0] = "ctr"
	values := make([]string, 1)
	values[0] = strconv.FormatUint(newCounterToSet, 10)
	// fmt.Println("Setting new value to redis = " + values[0])
	_, err = r.redisClient.EvalSha("8b071016ecfd75b7cce1c7d76591b4a4219b43cd", keys, values).Result()
	if err != nil {
		return nil, err
	}
	counterCheck, err := r.redisClient.Get(keys[0]).Uint64()
	if err != nil {
		return nil, err
		// fmt.Println(err)
		// os.Exit(1)
	}
	if counterCheck < newCounterToSet {
		return nil, errors.New("New counter is less than expected")
		// fmt.Println("Counter didn't increment")
		// os.Exit(1)
	}
	spendingRecords, err = r.getRecordsForBlock(newBlockNumber)
	if err != nil {
		return nil, err
	}
	spendingTXes := []*transaction.SignedTransaction{}
	inputLookupHashmap := &hashmap.HashMap{}
	for _, spendingRec := range spendingRecords {
		for _, utxoIndex := range spendingRec.OutputIndexes {
			val, _ := inputLookupHashmap.Get(utxoIndex[:])
			if val == nil {
				inputLookupHashmap.Set(utxoIndex[:], []byte{0x01})
			} else {
				return nil, errors.New("Potential doublespend")
			}
		}
		spendingTXes = append(spendingTXes, spendingRec.SpendingTransaction)
	}

	newBlock, err := block.NewBlock(newBlockNumber, spendingTXes, previousHash)
	if err != nil {
		return nil, err
	}
	elapsed := time.Since(start)
	fmt.Println("Block assembling taken " + fmt.Sprintf("%d", elapsed.Nanoseconds()/1000000) + " ms")

	return newBlock, nil
}
