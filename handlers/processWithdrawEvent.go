package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bankex/go-plasma/transaction"
	"github.com/bankex/go-plasma/types"
	"github.com/valyala/fasthttp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bankex/go-plasma/foundationdb"
	common "github.com/ethereum/go-ethereum/common"
)

type withdrawTXrequest struct {
	For          string `json:"address"`
	DepositIndex string `json:"depositIndex"`
	Value        string `json:"amount"`
}

type withdrawTXresponse struct {
	Error bool `json:"error"`
}

type WithdrawTXHandler struct {
	db        *fdb.Database
	txCreator *foundationdb.FundingTXcreator
}

func NewWithdrawTXHandler(db *fdb.Database) *WithdrawTXHandler {
	creator := foundationdb.NewFundingTXcreator(db)
	handler := &WithdrawTXHandler{db, creator}
	return handler
}

func (h *WithdrawTXHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var requestJSON createFundingTXrequest
	err := json.NewDecoder(r.Body).Decode(&requestJSON)
	if err != nil {
		writeErrorResponse(w)
		return
	}
	to := common.Address{}
	toBytes := common.FromHex(requestJSON.For)
	if len(toBytes) != transaction.AddressLength {
		writeErrorResponse(w)
		return
	}
	copy(to[:], toBytes)
	depositIndex := types.NewBigInt(0)
	depositIndex.SetString(requestJSON.DepositIndex, 10)
	value := types.NewBigInt(0)
	value.SetString(requestJSON.Value, 10)
	counter, err := h.redisClient.Incr("ctr").Result()
	if err != nil {
		writeErrorResponse(w)
		return
	}
	err = h.txCreator.CreateFundingTX(to, value, uint64(counter), depositIndex)
	if err != nil {
		writeErrorResponse(w)
		return
	}
	writeSuccessResponse(w)
	return
}

func (h *WithdrawTXHandler) HandlerFunc(ctx *fasthttp.RequestCtx) {
	var requestJSON createFundingTXrequest
	err := json.Unmarshal(ctx.PostBody(), &requestJSON)
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	to := common.Address{}
	toBytes := common.FromHex(requestJSON.For)
	if len(toBytes) != transaction.AddressLength {
		writeFasthttpErrorResponse(ctx)
		return
	}
	copy(to[:], toBytes)
	depositIndex := types.NewBigInt(0)
	depositIndex.SetString(requestJSON.DepositIndex, 10)
	value := types.NewBigInt(0)
	value.SetString(requestJSON.Value, 10)
	counter, err := h.redisClient.Incr("ctr").Result()
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	err = h.txCreator.CreateFundingTX(to, value, uint64(counter), depositIndex)
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	writeFasthttpSuccessResponse(ctx)
	return
}
