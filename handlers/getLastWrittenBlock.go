package handlers

import (
	"encoding/json"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/shamatar/go-plasma/foundationdb"
	"github.com/valyala/fasthttp"
)

type LastBlockHandler struct {
	db *fdb.Database
}

type lastBlockResponse struct {
	Error       bool `json:"error"`
	BlockNumber int  `json:"blockNumber"`
}

func NewLastBlockHandler(db *fdb.Database) *LastBlockHandler {
	handler := &LastBlockHandler{db}
	return handler
}

func (h *LastBlockHandler) HandlerFunc(ctx *fasthttp.RequestCtx) {
	lastBlock, err := foundationdb.GetLastWrittenBlock(h.db)
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	response := lastBlockResponse{Error: false, BlockNumber: int(lastBlock)}
	body, _ := json.Marshal(response)
	ctx.SetContentType("application/json")
	ctx.SetBody(body)
	ctx.SetStatusCode(fasthttp.StatusOK)
	return
}
