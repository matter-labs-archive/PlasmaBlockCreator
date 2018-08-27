package loadtest

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/caarlos0/env"
	"github.com/ethereum/go-ethereum/common"
	"github.com/matterinc/PlasmaCommons/crypto"
	"github.com/matterinc/PlasmaCommons/crypto/secp256k1"
	"github.com/valyala/fasthttp"
)

func TestServerConcurrently(t *testing.T) {
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("%+v\n", err)
	}
	fmt.Printf("%+v\n", cfg)
	// serverAddress = cfg.ServerAddr
	for i := 0; i < timesToRun; i++ {
		runConcurrently()
	}
}

func startCreationWorker(txNumbersChannel chan int) chan txCreatingResult {
	results := make(chan txCreatingResult)
	go func() {
		defer close(results)
		for txNumber := range txNumbersChannel {
			randomHash := make([]byte, 32)
			rand.Read(randomHash)
			randomBytes := make([]byte, 32)
			rand.Read(randomBytes)
			sig, err := secp256k1.Sign(randomHash, randomBytes)
			if err != nil {
				results <- txCreatingResult{0, false, []byte{}}
				continue
			}
			sender, err := secp256k1.RecoverPubkey(randomHash, sig)
			if err != nil {
				results <- txCreatingResult{0, false, []byte{}}
				continue
			}
			senderAddr := crypto.PubkeyToAddress(sender)
			data := map[string]interface{}{"for": common.ToHex(senderAddr[:]), "blockNumber": blockNumber, "transactionNumber": txNumber, "outputNumber": 0, "value": amountAsString}
			body, err := json.Marshal(data)
			req := fasthttp.AcquireRequest()
			req.Header.SetRequestURI("http://" + serverAddress + "/createUTXO")
			req.Header.SetMethod("POST")
			req.Header.SetContentLength(len(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetBody(body)
			resp := fasthttp.AcquireResponse()
			err = fastClient.Do(req, resp)
			if err != nil {
				results <- txCreatingResult{0, false, []byte{}}
				continue
			}
			respBody := resp.Body()
			if err != nil {
				results <- txCreatingResult{0, false, []byte{}}
				continue
			}
			var response responseStruct
			err = json.Unmarshal(respBody, &response)
			if err != nil {
				results <- txCreatingResult{0, false, []byte{}}
				continue
			}
			if response.Error {
				results <- txCreatingResult{0, false, []byte{}}
				continue
			}
			results <- txCreatingResult{txNumber, true, randomBytes}
		}
	}()
	return results
}

func startSpendingWorker(txesChannel chan string, wg *sync.WaitGroup, maxCapacity int) chan bool {
	results := make(chan bool, maxCapacity)
	go func() {
		defer close(results)
		defer wg.Done()
		for txHex := range txesChannel {
			data := map[string]interface{}{"tx": txHex}
			body, err := json.Marshal(data)
			req := fasthttp.AcquireRequest()
			req.Header.SetRequestURI("http://" + serverAddress + "/sendRawTX")
			req.Header.SetMethod("POST")
			req.Header.SetContentLength(len(body))
			req.SetBody(body)
			req.Header.Set("Content-Type", "application/json")
			resp := fasthttp.AcquireResponse()
			err = fastClient.Do(req, resp)
			if err != nil {
				results <- false
				continue
			}
			respBody := resp.Body()
			var response responseStruct
			err = json.Unmarshal(respBody, &response)
			if err != nil {
				results <- false
				continue
			}
			results <- !response.Error
		}
	}()
	return results
}

func mergeResults(cs []<-chan txCreatingResult) <-chan txCreatingResult {
	var wg sync.WaitGroup
	out := make(chan txCreatingResult)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan txCreatingResult) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func mergeBools(cs []<-chan bool) <-chan bool {
	var wg sync.WaitGroup
	out := make(chan bool)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan bool) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func runConcurrently() {
	var wg sync.WaitGroup

	fastClient = &fasthttp.PipelineClient{
		Addr:     serverAddress,
		MaxConns: connLimit,
	}
	rand.Seed(time.Now().UnixNano())
	blockNumber = int(rand.Uint32())
	fmt.Println("Creating " + strconv.Itoa(txToCreate) + " UTXOS")
	txNumbersChannel := make(chan int, txToCreate)
	for i := 0; i < txToCreate; i++ {
		txNumbersChannel <- i
	}
	close(txNumbersChannel)
	resultsChannels := make([]<-chan txCreatingResult, concurrencyLimit)
	for i := 0; i < concurrencyLimit; i++ {
		resultsChannels[i] = startCreationWorker(txNumbersChannel)
	}
	chanToCreate := mergeResults(resultsChannels)
	allTXes := []string{}
	validTxNumbers := []txCreatingResult{}
	for res := range chanToCreate {
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

	chanForSubmissions := make(chan string, len(allTXes))
	for i := 0; i < len(allTXes); i++ {
		chanForSubmissions <- allTXes[i]
	}
	close(chanForSubmissions)
	submissionResultsChannels := make([]<-chan bool, concurrencyLimit)

	// we are ready, start timing
	start := time.Now()
	for i := 0; i < concurrencyLimit; i++ {
		submissionResultsChannels[i] = startSpendingWorker(chanForSubmissions, &wg, len(allTXes))
		wg.Add(1)
	}
	wg.Wait()
	elapsed := time.Since(start)
	chanForResults := mergeBools(submissionResultsChannels)
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
