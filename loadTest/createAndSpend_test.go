package loadtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/bankex/go-plasma/transaction"
	"github.com/bankex/go-plasma/types"
	"github.com/ethereum/go-ethereum/common"
)

var txToCreate = 100000
var blockNumber = int(rand.Uint32())
var testAccount = "0xf62803ffaddda373d44b10bf6bb404909be0e66b"
var testAccountBinary = common.FromHex(testAccount)
var testPrivateKey = common.FromHex("0x7e2abf9c3bcd5c08c6d2156f0d55764602aed7b584c4e95fa01578e605d4cd32")
var amountAsString = "1000000000000000000"
var serverAddress = "http://127.0.0.1:3001"
var concurrencyLimit = 2000
var timeout = time.Duration(60 * time.Second)
var timesToRun = 10
var connLimit = 200

var httpClient *http.Client

type responseStruct struct {
	Error bool `json:"error"`
}

type txCreatingResult struct {
	txNumber int
	success  bool
}

func TestServer(t *testing.T) {
	for i := 0; i < timesToRun; i++ {
		run()
	}
}

func create(txNumber int, results chan txCreatingResult, wg *sync.WaitGroup) {
	defer wg.Done()
	data := map[string]interface{}{"for": testAccount, "blockNumber": blockNumber, "transactionNumber": txNumber, "outputNumber": 0, "value": amountAsString}
	body, err := json.Marshal(data)
	req, err := http.NewRequest("POST", serverAddress+"/createUTXO", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		results <- txCreatingResult{0, false}
		return
	}
	client := httpClient
	resp, err := client.Do(req)
	if err != nil {
		// fmt.Print(err)
		results <- txCreatingResult{0, false}
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// fmt.Print(err)
		results <- txCreatingResult{0, false}
		return
	}
	var response responseStruct
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		// fmt.Print(err)
		results <- txCreatingResult{0, false}
		return
	}
	if response.Error {
		// fmt.Print(err)
		results <- txCreatingResult{0, false}
		return
	}
	results <- txCreatingResult{txNumber, true}
}

func spend(str string, results chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	data := map[string]interface{}{"tx": str}
	body, err := json.Marshal(data)
	req, err := http.NewRequest("POST", serverAddress+"/sendRawTX", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		results <- false
		return
	}
	client := httpClient
	resp, err := client.Do(req)
	if err != nil {
		results <- false
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		results <- false
		return
	}
	var response responseStruct
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		results <- false
		return
	}
	results <- !response.Error
}

func createTransferTransaction(blockNumber int, txNumberInBlock int, outputNumberInTransaction int, value string) ([]byte, error) {
	bn := types.NewBigInt(int64(blockNumber))
	tn := types.NewBigInt(int64(txNumberInBlock))
	in := types.NewBigInt(int64(outputNumberInTransaction))
	v := types.NewBigInt(0)
	v.SetString(value, 10)
	input := &transaction.TransactionInput{}
	err := input.SetFields(bn, tn, in, v)
	if err != nil {
		return nil, err
	}
	bn = types.NewBigInt(0)
	to := common.Address{}
	copy(to[:], testAccountBinary)
	output := &transaction.TransactionOutput{}
	err = output.SetFields(bn, to, v)
	if err != nil {
		return nil, err
	}

	inputs := []*transaction.TransactionInput{input}
	outputs := []*transaction.TransactionOutput{output}
	txType := transaction.TransactionTypeSplit
	tx, err := transaction.NewUnsignedTransaction(txType, inputs, outputs)
	emptyBytes := [32]byte{}
	signed, err := transaction.NewSignedTransaction(tx, []byte{0x00}, emptyBytes[:], emptyBytes[:])
	signed.Sign(testPrivateKey)
	var b bytes.Buffer
	i := io.Writer(&b)
	err = signed.EncodeRLP(i)
	if err != nil {
		return nil, err
	}
	a := b.Bytes()
	return a, nil
}

func run() {
	defaultRoundTripper := http.DefaultTransport
	defaultTransportPointer, ok := defaultRoundTripper.(*http.Transport)
	if !ok {
		panic(fmt.Sprintf("defaultRoundTripper not an *http.Transport"))
	}
	defaultTransport := *defaultTransportPointer
	defaultTransport.MaxIdleConns = connLimit
	defaultTransport.MaxIdleConnsPerHost = connLimit
	httpClient = &http.Client{Transport: &defaultTransport, Timeout: timeout}
	// httpClient = &http.Client{Timeout: timeout}
	rand.Seed(time.Now().UnixNano())
	blockNumber = int(rand.Uint32())
	fmt.Println("Creating " + strconv.Itoa(txToCreate) + " UTXOS")
	chanForConcurrency := make(chan bool, concurrencyLimit)
	var wg sync.WaitGroup
	chanForCreate := make(chan txCreatingResult, 10000000)
	for i := 0; i < txToCreate; i++ {
		wg.Add(1)
		tmp := i
		chanForConcurrency <- true
		go func() {
			create(tmp, chanForCreate, &wg)
			<-chanForConcurrency
		}()
		if i%1000 == 0 {
			fmt.Println("Created " + strconv.Itoa(i))
		}
	}
	wg.Wait()
	fmt.Println("Created UTXOs")
	chanForResults := make(chan bool, 10000000)
	allTXes := []string{}
	validTxNumbers := []int{}
	close(chanForCreate)
	for res := range chanForCreate {
		if res.success {
			validTxNumbers = append(validTxNumbers, res.txNumber)
		}
	}

	for _, txNumber := range validTxNumbers {
		a, err := createTransferTransaction(blockNumber, txNumber, 0, amountAsString)
		if err != nil {
			continue
		}
		str := common.ToHex(a)
		allTXes = append(allTXes, str)
	}
	fmt.Println("Spending " + strconv.Itoa(len(allTXes)) + " UTXOS")
	start := time.Now()
	for i := 0; i < len(allTXes); i++ {
		wg.Add(1)
		chanForConcurrency <- true
		tx := allTXes[i]
		go func() {
			spend(tx, chanForResults, &wg)
			<-chanForConcurrency
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	close(chanForResults)
	counter := 0
	for res := range chanForResults {
		if res {
			counter++
		}
	}

	fmt.Println("Sent " + strconv.Itoa(counter) + " succesfully")
	txSpeed := float64(counter) / elapsed.Seconds()
	fmt.Println("TX speed = " + fmt.Sprintf("%f", txSpeed))

}

func main() {
	run()
}
