package handlers

import (
	"encoding/json"

	"github.com/matterinc/PlasmaCommons/transaction"
	"github.com/matterinc/PlasmaCommons/types"
	"github.com/valyala/fasthttp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/shamatar/go-plasma/foundationdb"
	common "github.com/ethereum/go-ethereum/common"
	redis "github.com/go-redis/redis"
)

type createFundingTXrequest struct {
	For          string `json:"_from"`
	DepositIndex string `json:"_depositIndex"`
	Value        string `json:"_amount"`
}

type createFundingTXresponse struct {
	Error bool `json:"error"`
}

type CreateFundingTXHandler struct {
	db          *fdb.Database
	redisClient *redis.Client
	txCreator   *foundationdb.FundingTXcreator
}

func NewCreateFundingTXHandler(db *fdb.Database, redisClient *redis.Client, signingKey []byte) *CreateFundingTXHandler {
	creator := foundationdb.NewFundingTXcreator(db, signingKey)
	handler := &CreateFundingTXHandler{db, redisClient, creator}
	return handler
}

func (h *CreateFundingTXHandler) HandlerFunc(ctx *fasthttp.RequestCtx) {
	var requestJSON createFundingTXrequest
	err := json.Unmarshal(ctx.PostBody(), &requestJSON)
	if err != nil {
		writeDepositResponse(ctx, true)
		return
	}
	to := common.Address{}
	toBytes := common.FromHex(requestJSON.For)
	if len(toBytes) != transaction.AddressLength {
		writeDepositResponse(ctx, true)
		return
	}
	copy(to[:], toBytes)
	depositIndex := types.NewBigInt(0)
	depositIndex.SetString(requestJSON.DepositIndex, 10)
	value := types.NewBigInt(0)
	value.SetString(requestJSON.Value, 10)
	counter, err := h.redisClient.Incr("ctr").Result()
	if err != nil {
		writeDepositResponse(ctx, true)
		return
	}
	err = h.txCreator.CreateFundingTX(to, value, uint64(counter), depositIndex)
	if err != nil {
		if err.Error() == "Duplicate funding transaction" {
			writeDepositResponse(ctx, false)
			return
		}
		writeDepositResponse(ctx, true)
		return
	}
	writeDepositResponse(ctx, false)
	return
}

func writeDepositResponse(ctx *fasthttp.RequestCtx, errorResult bool) {
	response := createFundingTXresponse{errorResult}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}
