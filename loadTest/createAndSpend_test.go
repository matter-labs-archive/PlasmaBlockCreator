package loadtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/bankex/go-plasma/crypto"
	"github.com/bankex/go-plasma/crypto/secp256k1"
	"github.com/bankex/go-plasma/transaction"
	"github.com/bankex/go-plasma/types"
	"github.com/caarlos0/env"
	"github.com/ethereum/go-ethereum/common"
	"github.com/valyala/fasthttp"
)

var txToCreate = 1000000

// var txToCreate = 100000
var blockNumber = int(rand.Uint32())

var doubleSpendProb = 0

// var testAccount = "0xf62803ffaddda373d44b10bf6bb404909be0e66b"
// var testPrivateKey = common.FromHex("0x7e2abf9c3bcd5c08c6d2156f0d55764602aed7b584c4e95fa01578e605d4cd32")

var testAccountBinary = common.FromHex("0xf62803ffaddda373d44b10bf6bb404909be0e66b")
var amountAsString = "1000000000000000000"
var testAccounts = []string{"0x627306090abab3a6e1400e9345bc60c78a8bef57",
	"0xf17f52151ebef6c7334fad080c5704d77216b732",
	"0xc5fdf4076b8f3a5357c5e395ab970b5b54098fef",
	"0x821aea9a577a9b44299b9c15c88cf3087f3b5544",
	"0x0d1d4e623d10f9fba5db95830f7d3839406c6af2",
	"0x2932b7a2355d6fecc4b5c0b6bd44cc31df247a2e",
	"0x2191ef87e392377ec08e7c08eb105ef5448eced5",
	"0x0f4f2ac550a1b4e2280d04c21cea7ebd822934b5",
	"0x6330a553fc93768f612722bb8c2ec78ac90b3bbc",
	"0x5aeda56215b167893e80b4fe645ba6d5bab767de"}

var testPrivateKeys = [][]byte{
	common.FromHex("0xc87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3"),
	common.FromHex("0xae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f"),
	common.FromHex("0x0dbbe8e4ae425a6d2687f1a7e3ba17bc98c673636790f1b8ad91193c05875ef1"),
	common.FromHex("0xc88b703fb08cbea894b6aeff5a544fb92e78a18e19814cd85da83b71f772aa6c"),
	common.FromHex("0x388c684f0ba1ef5017716adb5d21a053ea8e90277d0868337519f97bede61418"),
	common.FromHex("0x659cbb0e2411a44db63778987b1e22153c086a95eb6b18bdf89de078917abc63"),
	common.FromHex("0x82d052c865f5763aad42add438569276c00d3d88a2d062d36b2bae914d58b8c8"),
	common.FromHex("0xaa3680d5d48a8283413f7a108367c7299ca73f553735860a87b08f39395618b7"),
	common.FromHex("0x0f62d96d6675f32685bbdb8ac13cda7c23436f63efbb9d07700d8669ff12b7c4"),
	common.FromHex("0x8d5366123cb560bb606379f90a0bfd4769eecc0557f1b362dcae9012b548b1e5")}

var privateKeysTemp = [][]byte{}

var serverAddress = "127.0.0.1:3001"

type config struct {
	ServerAddr string `env:"TEST_SERVER" envDefault:"127.0.0.1:3001"`
}

var concurrencyLimit = 100000

// var concurrencyLimit = 10000
var timeout = time.Duration(60 * time.Second)
var timesToRun = 10

var connLimit = 30000

// var connLimit = 50
var fastClient *fasthttp.PipelineClient

// var httpClient *http.Client

type responseStruct struct {
	Error bool `json:"error"`
}

type txCreatingResult struct {
	txNumber   int
	success    bool
	privateKey []byte
}

func TestServer(t *testing.T) {
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("%+v\n", err)
	}
	fmt.Printf("%+v\n", cfg)
	// serverAddress = cfg.ServerAddr
	for i := 0; i < timesToRun; i++ {
		run()
	}
}

func create(txNumber int, results chan txCreatingResult, wg *sync.WaitGroup) {
	defer wg.Done()
	// randomAccountIndex := rand.Intn(len(testAccounts))
	// testAccount := testAccounts[randomAccountIndex]
	// testPrivateKey := testPrivateKeys[randomAccountIndex]
	randomHash := make([]byte, 32)
	rand.Read(randomHash)
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	sig, err := secp256k1.Sign(randomHash, randomBytes)
	if err != nil {
		results <- txCreatingResult{0, false, []byte{}}
		return
	}
	sender, err := secp256k1.RecoverPubkey(randomHash, sig)
	if err != nil {
		results <- txCreatingResult{0, false, []byte{}}
		return
	}
	senderAddr := crypto.PubkeyToAddress(sender)
	data := map[string]interface{}{"for": common.ToHex(senderAddr[:]), "blockNumber": blockNumber, "transactionNumber": txNumber, "outputNumber": 0, "value": amountAsString}
	body, err := json.Marshal(data)
	// req, err := http.NewRequest("POST", serverAddress+"/createUTXO", bytes.NewBuffer(body))
	// req.Header.Set("Content-Type", "application/json")
	// if err != nil {
	// 	results <- txCreatingResult{0, false, []byte{}}
	// 	return
	// }
	// client := httpClient
	// resp, err := client.Do(req)
	req := fasthttp.AcquireRequest()
	req.Header.SetRequestURI("http://" + serverAddress + "/createUTXO")
	req.Header.SetMethod("POST")
	req.Header.SetContentLength(len(body))
	req.Header.Set("Content-Type", "application/json")
	// req.Header.SetHost("http://" + serverAddress)
	req.SetBody(body)
	resp := fasthttp.AcquireResponse()
	err = fastClient.Do(req, resp)
	if err != nil {
		// fmt.Print(err)
		results <- txCreatingResult{0, false, []byte{}}
		return
	}
	// defer resp.Body.Close()
	// respBody, err := ioutil.ReadAll(resp.Body)
	respBody := resp.Body()
	if err != nil {
		// fmt.Print(err)
		results <- txCreatingResult{0, false, []byte{}}
		return
	}
	var response responseStruct
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		// fmt.Print(err)
		results <- txCreatingResult{0, false, []byte{}}
		return
	}
	if response.Error {
		// fmt.Print(err)
		results <- txCreatingResult{0, false, []byte{}}
		return
	}
	results <- txCreatingResult{txNumber, true, randomBytes}
}

func spend(str string, results chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	data := map[string]interface{}{"tx": str}
	body, err := json.Marshal(data)
	// req, err := http.NewRequest("POST", serverAddress+"/sendRawTX", bytes.NewBuffer(body))
	// req.Header.Set("Content-Type", "application/json")
	// if err != nil {
	// 	results <- false
	// 	return
	// }
	// client := httpClient
	// resp, err := client.Do(req)
	// if err != nil {
	// 	results <- false
	// 	return
	// }
	// defer resp.Body.Close()
	// respBody, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	results <- false
	// 	return
	// }

	req := fasthttp.AcquireRequest()
	req.Header.SetRequestURI("http://" + serverAddress + "/sendRawTX")
	req.Header.SetMethod("POST")
	req.Header.SetContentLength(len(body))
	req.SetBody(body)
	req.Header.Set("Content-Type", "application/json")
	// req.Header.SetHost("http://" + serverAddress)
	resp := fasthttp.AcquireResponse()
	err = fastClient.Do(req, resp)
	if err != nil {
		// fmt.Print(err)
		results <- false
		return
	}
	// defer resp.Body.Close()
	// respBody, err := ioutil.ReadAll(resp.Body)
	respBody := resp.Body()
	var response responseStruct
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		results <- false
		return
	}
	results <- !response.Error
}

func createTransferTransaction(blockNumber int, txNumberInBlock int, outputNumberInTransaction int, value string, testPrivateKey []byte) ([]byte, error) {
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
	// defaultRoundTripper := http.DefaultTransport
	// defaultTransportPointer, ok := defaultRoundTripper.(*http.Transport)
	// if !ok {
	// 	panic(fmt.Sprintf("defaultRoundTripper not an *http.Transport"))
	// }
	// defaultTransport := *defaultTransportPointer
	// defaultTransport.MaxIdleConns = connLimit
	// defaultTransport.MaxIdleConnsPerHost = connLimit
	// httpClient = &http.Client{Transport: &defaultTransport, Timeout: timeout}
	fastClient = &fasthttp.PipelineClient{
		Addr:     serverAddress,
		MaxConns: connLimit,
	}
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
		if i%10000 == 0 {
			fmt.Println("Created " + strconv.Itoa(i))
		}
	}
	wg.Wait()
	fmt.Println("Created UTXOs")
	chanForResults := make(chan bool, 10000000)
	allTXes := []string{}
	validTxNumbers := []txCreatingResult{}
	close(chanForCreate)
	for res := range chanForCreate {
		if res.success {
			validTxNumbers = append(validTxNumbers, res)
		}
	}
	validTXes := 0
	for _, res := range validTxNumbers {
		a, err := createTransferTransaction(blockNumber, res.txNumber, 0, amountAsString, res.privateKey)
		if err != nil {
			continue
		}
		str := common.ToHex(a)
		allTXes = append(allTXes, str)
		validTXes++
		randomInt := rand.Intn(10)
		if randomInt < doubleSpendProb {
			allTXes = append(allTXes, str)
		}
	}
	fmt.Println("Spending " + strconv.Itoa(len(allTXes)) + " UTXOS (including double spends)")
	fmt.Println("Valid transactions = " + strconv.Itoa(validTXes))
	start := time.Now()
	for i := 0; i < len(allTXes); i++ {
		wg.Add(1)
		chanForConcurrency <- true
		tx := allTXes[i]
		go func() {
			spend(tx, chanForResults, &wg)
			<-chanForConcurrency
		}()
		if i%10000 == 0 {
			fmt.Println("Spent " + strconv.Itoa(i))
		}
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
