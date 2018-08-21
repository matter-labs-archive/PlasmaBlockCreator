package dbtest_update

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/shamatar/go-plasma/transaction"
	"github.com/ethereum/go-ethereum/common"
)

var txToCreate = uint32(1000000)
var testAccount = "0xf62803ffaddda373d44b10bf6bb404909be0e66b"
var testAccountBinary = common.FromHex(testAccount)
var testPrivateKey = common.FromHex("0x7e2abf9c3bcd5c08c6d2156f0d55764602aed7b584c4e95fa01578e605d4cd32")
var amountAsString = "1000000000000000000"
var serverAddress = "http://127.0.0.1:3001"
var concurrencyLimit = 29000
var timeout = time.Duration(30 * time.Second)
var timesToRun = 10

type txCreatingResult struct {
	txNumber uint32
	success  bool
}

func TestServer(t *testing.T) {
	for i := 0; i < timesToRun; i++ {
		run()
	}
}

func insert(db *fdb.Database, blockNumber uint32, transactionNumber uint32, outputNumber uint8, results chan txCreatingResult, wg *sync.WaitGroup) {
	defer wg.Done()
	utxoIndexes := make([][]byte, 1)
	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)
	transactionNumberBuffer := make([]byte, transaction.TransactionNumberLength)
	binary.BigEndian.PutUint32(transactionNumberBuffer, transactionNumber)
	outputNumberBuffer := make([]byte, transaction.OutputNumberLength)
	outputNumberBuffer[0] = outputNumber
	// valueBuffer, err := value.GetLeftPaddedBytes(transaction.ValueLength)
	key := []byte{}
	// key = append(key, utxoIndexPrefix...)
	// key = append(key, address[:]...)
	key = append(key, blockNumberBuffer...)
	key = append(key, transactionNumberBuffer...)
	key = append(key, outputNumberBuffer...)
	// key = append(key, valueBuffer...)
	utxoIndexes[0] = key
	_, err := db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		// for _, index := range utxoIndexes {
		// 	existing, err := tr.Get(fdb.Key(index)).Get()
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	if len(existing) != 0 {
		// 		return nil, errors.New("Record already exists")
		// 	}
		// }
		for _, index := range utxoIndexes {
			tr.Set(fdb.Key(index), []byte{0x01})
		}
		// for _, index := range utxoIndexes {
		// 	existing, err := tr.Get(fdb.Key(index)).Get()
		// 	if err != nil {
		// 		tr.Reset()
		// 		return nil, err
		// 	}
		// 	if len(existing) != 1 || bytes.Compare(existing, []byte{UTXOisReadyForSpending}) != 0 {
		// 		tr.Reset()
		// 		return nil, errors.New("Reading mismatch")
		// 	}
		// }
		return nil, nil
	})
	if err != nil {
		results <- txCreatingResult{0, false}
		return
	}
	results <- txCreatingResult{transactionNumber, true}
}

func update(db *fdb.Database, blockNumber uint32, transactionNumber uint32, outputNumber uint8, results chan txCreatingResult, wg *sync.WaitGroup) {
	defer wg.Done()
	utxoIndexes := make([][]byte, 1)
	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)
	transactionNumberBuffer := make([]byte, transaction.TransactionNumberLength)
	binary.BigEndian.PutUint32(transactionNumberBuffer, transactionNumber)
	outputNumberBuffer := make([]byte, transaction.OutputNumberLength)
	outputNumberBuffer[0] = outputNumber
	// valueBuffer, err := value.GetLeftPaddedBytes(transaction.ValueLength)
	key := []byte{}
	// key = append(key, utxoIndexPrefix...)
	// key = append(key, address[:]...)
	key = append(key, blockNumberBuffer...)
	key = append(key, transactionNumberBuffer...)
	key = append(key, outputNumberBuffer...)
	// key = append(key, valueBuffer...)
	utxoIndexes[0] = key
	_, err := db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		// for _, index := range utxoIndexes {
		// 	existing, err := tr.Get(fdb.Key(index)).Get()
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	if len(existing) != 0 {
		// 		return nil, errors.New("Record already exists")
		// 	}
		// }
		for _, index := range utxoIndexes {
			tr.Clear(fdb.Key(index))
		}
		// for _, index := range utxoIndexes {
		// 	existing, err := tr.Get(fdb.Key(index)).Get()
		// 	if err != nil {
		// 		tr.Reset()
		// 		return nil, err
		// 	}
		// 	if len(existing) != 1 || bytes.Compare(existing, []byte{UTXOisReadyForSpending}) != 0 {
		// 		tr.Reset()
		// 		return nil, errors.New("Reading mismatch")
		// 	}
		// }
		return nil, nil
	})
	if err != nil {
		results <- txCreatingResult{0, false}
		return
	}
	results <- txCreatingResult{transactionNumber, true}
}

func run() {
	fdb.MustAPIVersion(510)
	// foundDB := fdb.MustOpenDefault()
	rand.Seed(time.Now().UnixNano())
	blockNumber := rand.Uint32()
	fmt.Println("Inserting " + strconv.Itoa(int(txToCreate)) + " records")
	chanForConcurrency := make(chan bool, concurrencyLimit)
	var wg sync.WaitGroup
	chanForCreate := make(chan txCreatingResult, 1000000000)
	start := time.Now()
	for i := uint32(0); i < txToCreate; i++ {
		wg.Add(1)
		bn := blockNumber + i
		tmp := i
		outputNum := uint8(0)
		chanForConcurrency <- true
		go func() {
			foundDB := fdb.MustOpenDefault()
			insert(&foundDB, bn, tmp, outputNum, chanForCreate, &wg)
			<-chanForConcurrency
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println("Created records")
	close(chanForCreate)

	counter := 0
	for res := range chanForCreate {
		if res.success {
			counter++
		}
	}

	fmt.Println("Inserted " + strconv.Itoa(counter) + " succesfully")
	txSpeed := float64(counter) / elapsed.Seconds()
	fmt.Println("TX insert speed = " + fmt.Sprintf("%f", txSpeed))

	fmt.Println("Updating " + strconv.Itoa(int(txToCreate)) + " records")
	chanForUpdate := make(chan txCreatingResult, 1000000000)
	start = time.Now()
	for i := uint32(0); i < txToCreate; i++ {
		wg.Add(1)
		bn := blockNumber + i
		tmp := i
		outputNum := uint8(0)
		chanForConcurrency <- true
		go func() {
			foundDB := fdb.MustOpenDefault()
			update(&foundDB, bn, tmp, outputNum, chanForUpdate, &wg)
			<-chanForConcurrency
		}()
	}
	wg.Wait()
	elapsed = time.Since(start)
	fmt.Println("Updated records")
	close(chanForUpdate)

	counter = 0
	for res := range chanForUpdate {
		if res.success {
			counter++
		}
	}

	fmt.Println("Updated " + strconv.Itoa(counter) + " succesfully")
	txSpeed = float64(counter) / elapsed.Seconds()
	fmt.Println("TX update speed = " + fmt.Sprintf("%f", txSpeed))
}

func main() {
	run()
}
