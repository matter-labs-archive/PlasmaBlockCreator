package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/matterinc/PlasmaCommons/types"
	"github.com/valyala/fasthttp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/matterinc/PlasmaBlockCreator/foundationdb"
)

type depositWithdrawTXrequest struct {
	Index string `json:"_depositIndex"`
}

type depositWithdrawTXresponse struct {
	Error                   bool   `json:"error"`
	Action                  string `json:"action,omitempty"`
	BlockForChallenge       string `json:"blockForChallenge,omitempty"`
	TransactionForChallenge string `json:"transactionForChallenge,omitempty"`
}

type DepositWithdrawTXHandler struct {
	db *fdb.Database
}

func NewDepositWithdrawTXHandler(db *fdb.Database) *DepositWithdrawTXHandler {
	handler := &DepositWithdrawTXHandler{db}
	return handler
}

func (h *DepositWithdrawTXHandler) HandlerFunc(ctx *fasthttp.RequestCtx) {
	var requestJSON depositWithdrawTXrequest
	err := json.Unmarshal(ctx.PostBody(), &requestJSON)
	if err != nil {
		writeDepositWithdrawResponse(ctx, false)
		return
	}
	depositIndex := types.NewBigInt(0)
	depositIndex.SetString(requestJSON.Index, 10)
	information, err := foundationdb.LookupDepositIndex(h.db, depositIndex)
	if err != nil {
		writeDepositWithdrawResponse(ctx, false)
		return
	}
	writeDepositWithdrawChallengeRequiredResponse(ctx, information)
	return
}

func writeDepositWithdrawResponse(ctx *fasthttp.RequestCtx, result bool) {
	response := depositWithdrawTXresponse{!result, "", "", ""}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}

func writeDepositWithdrawChallengeRequiredResponse(ctx *fasthttp.RequestCtx, lookup *foundationdb.DepositLookupResult) {
	response := depositWithdrawTXresponse{false,
		"sendChallenge",
		strconv.Itoa(lookup.BlockNumber),
		strconv.Itoa(lookup.TransactionNumber)}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}
