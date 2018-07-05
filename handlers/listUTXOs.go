package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bankex/go-plasma/transaction"
	"github.com/valyala/fasthttp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bankex/go-plasma/foundationdb"
	common "github.com/ethereum/go-ethereum/common"
)

type listUTXOsRequest struct {
	For               string `json:"for"`
	BlockNumber       int    `json:"blockNumber"`
	TransactionNumber int    `json:"transactionNumber"`
	OutputNumber      int    `json:"outputNumber"`
	Limit             int    `json:"limit,omitempty"`
}

type singleUTXOdetails struct {
	BlockNumber       int    `json:"blockNumber"`
	TransactionNumber int    `json:"transactionNumber"`
	OutputNumber      int    `json:"outputNumber"`
	Value             string `json:"value"`
}

type listUTXOsResponse struct {
	Error bool                `json:"error"`
	UTXOs []singleUTXOdetails `json:"utxos"`
}

type ListUTXOsHandler struct {
	db         *fdb.Database
	utxoLister *foundationdb.UTXOlister
}

func NewListUTXOsHandler(db *fdb.Database) *ListUTXOsHandler {
	lister := foundationdb.NewUTXOlister(db)
	handler := &ListUTXOsHandler{db, lister}
	return handler
}

func (h *ListUTXOsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var requestJSON listUTXOsRequest
	err := json.NewDecoder(r.Body).Decode(&requestJSON)
	if err != nil {
		writeEmptyResponse(w)
		return
	}

	forBytes := common.FromHex(requestJSON.For)
	address := common.Address{}
	copy(address[:], forBytes)
	blockNumber := uint32(requestJSON.BlockNumber)
	transactionNumber := uint32(requestJSON.TransactionNumber)
	outputNumber := uint8(requestJSON.OutputNumber)
	limit := 50
	if requestJSON.Limit != 0 {
		limit = requestJSON.Limit
	}
	if limit > 100 {
		limit = 100
	}
	utxos, err := h.utxoLister.GetUTXOsForAddress(address, blockNumber, transactionNumber, outputNumber, limit)
	if err != nil {
		writeEmptyResponse(w)
		return
	}
	details := make([]singleUTXOdetails, len(utxos))
	for i, utxo := range utxos {
		detail := transaction.ParseIndexIntoUTXOdetails(utxo)
		responseDetails := singleUTXOdetails{int(detail.BlockNumber), int(detail.TransactionNumber),
			int(detail.OutputNumber), detail.Value}
		details[i] = responseDetails
	}
	writeResponse(w, details)
	return
}

func (h *ListUTXOsHandler) HandlerFunc(ctx *fasthttp.RequestCtx) {
	var requestJSON listUTXOsRequest
	err := json.Unmarshal(ctx.PostBody(), &requestJSON)
	if err != nil {
		writeEmptyFasthttpResponse(ctx)
		return
	}

	forBytes := common.FromHex(requestJSON.For)
	address := common.Address{}
	copy(address[:], forBytes)
	blockNumber := uint32(requestJSON.BlockNumber)
	transactionNumber := uint32(requestJSON.TransactionNumber)
	outputNumber := uint8(requestJSON.OutputNumber)
	limit := 50
	if requestJSON.Limit != 0 {
		limit = requestJSON.Limit
	}
	if limit > 100 {
		limit = 100
	}
	utxos, err := h.utxoLister.GetUTXOsForAddress(address, blockNumber, transactionNumber, outputNumber, limit)
	if err != nil {
		writeEmptyFasthttpResponse(ctx)
		return
	}
	details := make([]singleUTXOdetails, len(utxos))
	for i, utxo := range utxos {
		detail := transaction.ParseIndexIntoUTXOdetails(utxo)
		responseDetails := singleUTXOdetails{int(detail.BlockNumber), int(detail.TransactionNumber),
			int(detail.OutputNumber), detail.Value}
		details[i] = responseDetails
	}
	writeFasthttpResponse(ctx, details)
	return
}

func writeEmptyResponse(w http.ResponseWriter) {
	response := listUTXOsResponse{false, []singleUTXOdetails{}}
	json.NewEncoder(w).Encode(response)

}

func writeResponse(w http.ResponseWriter, details []singleUTXOdetails) {
	response := listUTXOsResponse{false, details}
	json.NewEncoder(w).Encode(response)
}

func writeEmptyFasthttpResponse(ctx *fasthttp.RequestCtx) {
	response := listUTXOsResponse{false, []singleUTXOdetails{}}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)

}

func writeFasthttpResponse(ctx *fasthttp.RequestCtx, details []singleUTXOdetails) {
	response := listUTXOsResponse{false, details}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}
