package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bankex/go-plasma/block"
	"github.com/bankex/go-plasma/transaction"
	"github.com/bankex/go-plasma/types"

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

func (h *CreateFundingTXHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

func writeBlockResponse(block *block.Block, w http.ResponseWriter) {

}
