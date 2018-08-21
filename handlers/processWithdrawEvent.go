package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/shamatar/go-plasma/transaction"
	"github.com/shamatar/go-plasma/types"
	"github.com/valyala/fasthttp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	common "github.com/ethereum/go-ethereum/common"
	"github.com/shamatar/go-plasma/foundationdb"
)

type withdrawTXrequest struct {
	For   string `json:"_from"`
	Index string `json:"_index"`
}

type withdrawTXresponse struct {
	Error                   bool   `json:"error"`
	Action                  string `json:"action,omitempty"`
	BlockForChallenge       string `json:"blockForChallenge,omitempty"`
	TransactionForChallenge string `json:"transactionForChallenge,omitempty"`
	InputForChallenge       string `json:"inputForChallenge,omitempty"`
}

type WithdrawTXHandler struct {
	db               *fdb.Database
	txWithdrawMarker *foundationdb.WithdrawTXMarker
}

func NewWithdrawTXHandler(db *fdb.Database) *WithdrawTXHandler {
	marker := foundationdb.NewWithdrawTXMarker(db)
	handler := &WithdrawTXHandler{db, marker}
	return handler
}

func (h *WithdrawTXHandler) HandlerFunc(ctx *fasthttp.RequestCtx) {
	var requestJSON withdrawTXrequest
	err := json.Unmarshal(ctx.PostBody(), &requestJSON)
	if err != nil {
		writeWithdrawResponse(ctx, false)
		return
	}
	to := common.Address{}
	toBytes := common.FromHex(requestJSON.For)
	if len(toBytes) != transaction.AddressLength {
		writeWithdrawResponse(ctx, false)
		return
	}
	copy(to[:], toBytes)
	utxoIndex := types.NewBigInt(0)
	utxoIndex.SetString(requestJSON.Index, 10)
	success, err := h.txWithdrawMarker.MarkTX(to, utxoIndex)
	if err != nil {
		writeWithdrawResponse(ctx, false)
		return
	}
	if success != true {
		lookup, err := foundationdb.LookupSpendingIndex(h.db, utxoIndex)
		if err != nil {
			writeWithdrawResponse(ctx, false)
			return
		}
		writeWithdrawChallengeRequiredResponse(ctx, lookup)
		return
	}
	writeWithdrawResponse(ctx, true)
	return
}

func writeWithdrawResponse(ctx *fasthttp.RequestCtx, result bool) {
	response := withdrawTXresponse{!result, "", "", "", ""}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}

func writeWithdrawChallengeRequiredResponse(ctx *fasthttp.RequestCtx, lookup *foundationdb.SpendingLookupResult) {
	response := withdrawTXresponse{false,
		"sendChallenge",
		strconv.Itoa(lookup.BlockNumber),
		strconv.Itoa(lookup.TransactionNumber),
		strconv.Itoa(lookup.InputNumber)}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}
