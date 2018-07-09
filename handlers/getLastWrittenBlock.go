package handlers

import (
	"encoding/json"
	"net/http"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bankex/go-plasma/foundationdb"
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

func (h *LastBlockHandler) Handle(w http.ResponseWriter, r *http.Request) {
	lastBlock, err := foundationdb.GetLastWrittenBlock(h.db)
	if err != nil {
		writeErrorResponse(w)
		return
	}
	response := lastBlockResponse{Error: false, BlockNumber: int(lastBlock)}
	json.NewEncoder(w).Encode(response)
	return
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
