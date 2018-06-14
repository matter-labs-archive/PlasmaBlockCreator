package handlers

import (
	"encoding/json"
	"net/http"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bankex/go-plasma/foundationdb"
)

type LastBlockHandler struct {
	db *fdb.Database
}

type lastBlockResponse struct {
	Error       bool `json:"error"`
	BlockNumber int  `json:"blockNumber,omitempty"`
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
