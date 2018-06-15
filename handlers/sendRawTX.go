package handlers

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	foundationdb "github.com/bankex/go-plasma/foundationdb"
	transaction "github.com/bankex/go-plasma/transaction"
	common "github.com/ethereum/go-ethereum/common"
	rlp "github.com/ethereum/go-ethereum/rlp"
	redis "github.com/go-redis/redis"
)

type SendRawTXHandler struct {
	db          *fdb.Database
	redisClient *redis.Client
	utxoReader  *foundationdb.UTXOReader
	utxoWriter  *foundationdb.UTXOWriter
	ops         uint64
}

func NewSendRawTXHandler(db *fdb.Database, redisClient *redis.Client) *SendRawTXHandler {
	reader := foundationdb.NewUTXOReader(db)
	writer := foundationdb.NewUTXOWriter(db)
	handler := &SendRawTXHandler{db, redisClient, reader, writer, 0}
	return handler
}

func (h *SendRawTXHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var requestJSON sendRawRLPTXRequest
	err := json.NewDecoder(r.Body).Decode(&requestJSON)
	if err != nil {
		// log.Println("Failed to decode JSON")
		// log.Printf("%+v\n", err)
		// writeDebugResponse(w, "Cound't decode JSON")
		writeErrorResponse(w)
		return
	}
	bytes := common.FromHex(requestJSON.TX)
	if bytes == nil || len(bytes) == 0 {
		// log.Println("Failed to decode hex string")
		// writeDebugResponse(w, "Cound't decode hex string")
		writeErrorResponse(w)
		return
	}
	tx := &(transaction.SignedTransaction{})
	err = rlp.DecodeBytes(bytes, tx)
	if err != nil {
		// log.Println("Failed to decode transaction")
		// log.Printf("%+v\n", err)
		// writeDebugResponse(w, "Cound't decode transaction")
		writeErrorResponse(w)
		return
	}
	err = tx.Validate()
	if err != nil {
		// log.Println("Transaction is invalid")
		// log.Printf("%+v\n", err)
		// writeDebugResponse(w, "Cound't validate transaction")
		writeErrorResponse(w)
		return
	}
	tx.RawValue = bytes
	err = h.utxoReader.CheckIfUTXOsExist(tx)
	if err != nil {
		// log.Println("UTXO doesn't exist")
		// log.Printf("%+v\n", err)
		// writeDebugResponse(w, "UTXO doesn't exist")
		writeErrorResponse(w)
		return
	}
	atomic.AddUint64(&h.ops, 1)
	counter := atomic.LoadUint64(&h.ops)
	// counter, err := h.redisClient.Incr("ctr").Result()
	if err != nil {
		// log.Println("Failed to get counter")
		// log.Printf("%+v\n", err)
		// writeDebugResponse(w, "Cound't get counter")
		writeErrorResponse(w)
		return
	}
	err = h.utxoWriter.WriteSpending(tx, uint64(counter))
	if err != nil {
		// log.Println("Cound't write transaction")
		// log.Printf("%+v\n", err)
		// writeDebugResponse(w, "Cound't write transaction")
		writeErrorResponse(w)
		return
	}
	writeSuccessResponse(w)
	return
}

// func writeErrorResponse(w http.ResponseWriter) {
// 	response := sendRawRLPTXResponse{Error: true, Reason: "invalid transaction"}
// 	json.NewEncoder(w).Encode(response)
// }

// func writeDebugResponse(w http.ResponseWriter, reason string) {
// 	response := sendRawRLPTXResponse{Error: true, Reason: reason}
// 	json.NewEncoder(w).Encode(response)
// }

// func writeSuccessResponse(w http.ResponseWriter) {
// 	response := sendRawRLPTXResponse{Error: false, Accepted: true}
// 	json.NewEncoder(w).Encode(response)
// }
