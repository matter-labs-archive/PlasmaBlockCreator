package handlers

import (
	"encoding/json"

	"github.com/bankex/go-plasma/transaction"
	"github.com/bankex/go-plasma/types"
	"github.com/valyala/fasthttp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bankex/go-plasma/foundationdb"
	common "github.com/ethereum/go-ethereum/common"
	redis "github.com/go-redis/redis"
)

type createFundingTXrequest struct {
	For          string `json:"address"`
	DepositIndex string `json:"depositIndex"`
	Value        string `json:"amount"`
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
