package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	sql "database/sql"

	sqlfunctions "github.com/bankex/go-plasma/sqlfunctions"
	transaction "github.com/bankex/go-plasma/transaction"
	common "github.com/ethereum/go-ethereum/common"
	rlp "github.com/ethereum/go-ethereum/rlp"
	redis "github.com/go-redis/redis"
)

type sendRawRLPTXRequest struct {
	TX string `json:"tx,omitempty"`
}

type sendRawRLPTXResponse struct {
	Error    bool   `json:"error"`
	Accepted bool   `json:"accepted,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type SendRawRLPTXHandler struct {
	db          *sql.DB
	redisClient *redis.Client
	utxoReader  *sqlfunctions.UTXOreader
	utxoWriter  *sqlfunctions.TransactionSpendingWriter
}

func NewSendRawRLPTXHandler(db *sql.DB, redisClient *redis.Client) *SendRawRLPTXHandler {
	reader := sqlfunctions.NewUTXOReader(db)
	writer := sqlfunctions.NewTransactionSpendingWriter(db)
	handler := &SendRawRLPTXHandler{db, redisClient, reader, writer}
	return handler
}

func (h *SendRawRLPTXHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var requestJSON sendRawRLPTXRequest
	log.Fatal("Got request")
	err := json.NewDecoder(r.Body).Decode(&requestJSON)
	if err != nil {
		go log.Println("Failed to decode JSON")
		go log.Printf("%+v\n", err)
		writeErrorResponse(w)
		return
	}
	bytes := common.FromHex(requestJSON.TX)
	if bytes == nil || len(bytes) == 0 {
		fmt.Println("Failed to decode hex string")
		writeErrorResponse(w)
		return
	}
	tx := &(transaction.SignedTransaction{})
	err = rlp.DecodeBytes(bytes, tx)
	if err != nil {
		fmt.Println("Failed to decode transaction")
		fmt.Printf("%+v\n", err)
		writeErrorResponse(w)
		return
	}
	err = tx.Validate()
	if err != nil {
		fmt.Println("Transaction is invalid")
		fmt.Printf("%+v\n", err)
		writeErrorResponse(w)
		return
	}
	tx.RawValue = bytes
	exists, err := h.utxoReader.CheckIfUTXOsExist(tx)
	if err != nil || !exists {
		fmt.Println("UTXO doesn't exist")
		fmt.Printf("%+v\n", err)
		writeErrorResponse(w)
		return
	}
	counter, err := h.redisClient.Incr("ctr").Result()
	if err != nil {
		writeErrorResponse(w)
		return
	}
	writen, err := h.utxoWriter.WriteSpending(tx, counter)
	if err != nil || !writen {
		fmt.Println("Cound't write transaction")
		fmt.Printf("%+v\n", err)
		writeErrorResponse(w)
		return
	}
	writeSuccessResponse(w)
	return
}

func writeErrorResponse(w http.ResponseWriter) {
	response := sendRawRLPTXResponse{Error: true, Reason: "invalid transaction"}
	json.NewEncoder(w).Encode(response)
}

func writeSuccessResponse(w http.ResponseWriter) {
	response := sendRawRLPTXResponse{Error: false, Accepted: true}
	json.NewEncoder(w).Encode(response)
}
