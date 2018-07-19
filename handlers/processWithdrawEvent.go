package handlers

import (
	"encoding/json"

	"github.com/bankex/go-plasma/transaction"
	"github.com/bankex/go-plasma/types"
	"github.com/valyala/fasthttp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bankex/go-plasma/foundationdb"
	common "github.com/ethereum/go-ethereum/common"
)

type withdrawTXrequest struct {
	For   string `json:"address"`
	Index string `json:"index"`
}

type withdrawTXresponse struct {
	Error bool `json:"error"`
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
		writeWithdrawResponse(ctx, false)
		return
	}
	writeWithdrawResponse(ctx, true)
	return
}

func writeWithdrawResponse(ctx *fasthttp.RequestCtx, result bool) {
	response := withdrawTXresponse{result}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}
