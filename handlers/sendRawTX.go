package handlers

import (
	"encoding/json"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	foundationdb "github.com/bankex/go-plasma/foundationdb"
	transaction "github.com/bankex/go-plasma/transaction"
	common "github.com/ethereum/go-ethereum/common"
	redis "github.com/go-redis/redis"
	"github.com/valyala/fasthttp"
)

type SendRawTXHandler struct {
	db          *fdb.Database
	redisClient *redis.Client
	utxoReader  *foundationdb.UTXOReader
	utxoWriter  *foundationdb.UTXOWriter
	parser      *transaction.TransactionParser
}

func NewSendRawTXHandler(db *fdb.Database, redisClient *redis.Client, parser *transaction.TransactionParser, writerConcurrency int) *SendRawTXHandler {
	reader := foundationdb.NewUTXOReader(db)
	writer := foundationdb.NewUTXOWriter(db, writerConcurrency)
	handler := &SendRawTXHandler{db, redisClient, reader, writer, parser}
	return handler
}

// func (h *SendRawTXHandler) Handle(w http.ResponseWriter, r *http.Request) {
// 	var requestJSON sendRawRLPTXRequest
// 	err := json.NewDecoder(r.Body).Decode(&requestJSON)
// 	if err != nil {
// 		writeErrorResponse(w)
// 		return
// 	}
// 	bytes := common.FromHex(requestJSON.TX)
// 	if bytes == nil || len(bytes) == 0 {
// 		writeErrorResponse(w)
// 		return
// 	}
// 	tx := &(transaction.SignedTransaction{})
// 	err = rlp.DecodeBytes(bytes, tx)
// 	if err != nil {
// 		writeErrorResponse(w)
// 		return
// 	}
// 	err = tx.Validate()
// 	if err != nil {
// 		writeErrorResponse(w)
// 		return
// 	}
// 	tx.RawValue = bytes
// 	// err = h.utxoReader.CheckIfUTXOsExist(tx)
// 	// if err != nil {
// 	// 	writeErrorResponse(w)
// 	// 	return
// 	// }
// 	counter, err := h.redisClient.Incr("ctr").Result()
// 	if err != nil {
// 		writeErrorResponse(w)
// 		return
// 	}
// 	err = h.utxoWriter.WriteSpending(tx, uint64(counter))
// 	if err != nil {
// 		writeErrorResponse(w)
// 		return
// 	}
// 	writeSuccessResponse(w)
// 	return
// }

func (h *SendRawTXHandler) HandlerFunc(ctx *fasthttp.RequestCtx) {
	var requestJSON sendRawRLPTXRequest
	err := json.Unmarshal(ctx.PostBody(), &requestJSON)
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	bytes := common.FromHex(requestJSON.TX)
	if bytes == nil || len(bytes) == 0 {
		writeFasthttpErrorResponse(ctx)
		return
	}
	parsedRes, err := h.parser.Parse(bytes)
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	// err = h.utxoReader.CheckIfUTXOsExist(tx)
	// if err != nil {
	// 	writeFasthttpErrorResponse(ctx)
	// 	return
	// }
	counter, err := h.redisClient.Incr("ctr").Result()
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	writeFasthttpSuccessResponse(ctx)
	return

	err = h.utxoWriter.WriteSpending(parsedRes, uint64(counter))
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	writeFasthttpSuccessResponse(ctx)
	return
}

func writeFasthttpErrorResponse(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	response := sendRawRLPTXResponse{Error: true, Reason: "invalid transaction"}
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}

func writeFasthttpSuccessResponse(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	response := sendRawRLPTXResponse{Error: false, Accepted: true}
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}
