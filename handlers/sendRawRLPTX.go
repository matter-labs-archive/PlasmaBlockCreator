package handlers

import (
	"encoding/json"
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
	err := json.NewDecoder(r.Body).Decode(&requestJSON)
	if err != nil {
		log.Println("Failed to decode JSON")
		log.Printf("%+v\n", err)
		writeErrorResponse(w)
		return
	}
	bytes := common.FromHex(requestJSON.TX)
	if bytes == nil || len(bytes) == 0 {
		log.Println("Failed to decode hex string")
		writeErrorResponse(w)
		return
	}
	tx := &(transaction.SignedTransaction{})
	err = rlp.DecodeBytes(bytes, tx)
	if err != nil {
		log.Println("Failed to decode transaction")
		log.Printf("%+v\n", err)
		writeDebugResponse(w, "Cound't decode transaction")
		// writeErrorResponse(w)
		return
	}
	err = tx.Validate()
	if err != nil {
		log.Println("Transaction is invalid")
		log.Printf("%+v\n", err)
		writeDebugResponse(w, "Cound't validate transaction")
		// writeErrorResponse(w)
		return
	}
	tx.RawValue = bytes
	exists, err := h.utxoReader.CheckIfUTXOsExist(tx)
	if err != nil || !exists {
		log.Println("UTXO doesn't exist")
		log.Printf("%+v\n", err)
		writeDebugResponse(w, "UTXO doesn't exist")
		// writeErrorResponse(w)
		return
	}
	counter, err := h.redisClient.Incr("ctr").Result()
	if err != nil {
		log.Println("Failed to get counter")
		log.Printf("%+v\n", err)
		writeErrorResponse(w)
		return
	}
	writen, err := h.utxoWriter.WriteSpending(tx, counter)
	if err != nil || !writen {
		log.Println("Cound't write transaction")
		log.Printf("%+v\n", err)
		writeDebugResponse(w, "Cound't write transaction")
		// writeErrorResponse(w)
		return
	}
	writeSuccessResponse(w)
	return
}

func writeErrorResponse(w http.ResponseWriter) {
	response := sendRawRLPTXResponse{Error: true, Reason: "invalid transaction"}
	json.NewEncoder(w).Encode(response)
}

func writeDebugResponse(w http.ResponseWriter, reason string) {
	response := sendRawRLPTXResponse{Error: true, Reason: reason}
	json.NewEncoder(w).Encode(response)
}

func writeSuccessResponse(w http.ResponseWriter) {
	response := sendRawRLPTXResponse{Error: false, Accepted: true}
	json.NewEncoder(w).Encode(response)
}
