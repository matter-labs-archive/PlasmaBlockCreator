package handlers

import (
	"encoding/json"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	common "github.com/ethereum/go-ethereum/common"
	redis "github.com/go-redis/redis"
	commonTools "github.com/shamatar/go-plasma/common"
	foundationdb "github.com/shamatar/go-plasma/foundationdb"
	transaction "github.com/shamatar/go-plasma/transaction"
	"github.com/valyala/fasthttp"
)

type sendRawRLPTXRequest struct {
	TX string `json:"tx,omitempty"`
}

type sendRawRLPTXResponse struct {
	Error    bool   `json:"error"`
	Accepted bool   `json:"accepted,omitempty"`
	Reason   string `json:"reason,omitempty"`
}
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
	err = h.utxoReader.CheckIfUTXOsExist(&parsedRes.TX)
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	// counter, err := h.redisClient.Incr("ctr").Result()
	// if err != nil {
	// 	writeFasthttpErrorResponse(ctx)
	// 	return
	// }
	counter := commonTools.GetCounter()
	err = h.utxoWriter.WriteSpending(parsedRes, uint64(counter))
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	writeFasthttpSuccessResponse(ctx)
	return
}
