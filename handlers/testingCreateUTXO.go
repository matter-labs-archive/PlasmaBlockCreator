package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bankex/go-plasma/foundationdb"
	"github.com/bankex/go-plasma/types"
	common "github.com/ethereum/go-ethereum/common"
)

type createUTXOrequest struct {
	For               string `json:"for,omitempty"`
	BlockNumber       int    `json:"blockNumber,omitempty"`
	TransactionNumber int    `json:"transactionNumber,omitempty"`
	OutputNumber      int    `json:"outputNumber,omitempty"`
	Value             string `json:"value,omitempty"`
}

type createUTXOResponse struct {
	Error bool `json:"error"`
}

type CreateUTXOHandler struct {
	db          *fdb.Database
	utxoCreator *foundationdb.TestUTXOcreator
}

func NewCreateUTXOHandler(db *fdb.Database) *CreateUTXOHandler {
	creator := foundationdb.NewTestUTXOcreator(db)
	handler := &CreateUTXOHandler{db, creator}
	return handler
}

func (h *CreateUTXOHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var requestJSON createUTXOrequest
	err := json.NewDecoder(r.Body).Decode(&requestJSON)
	if err != nil {
		log.Println("Failed to decode JSON")
		log.Printf("%+v\n", err)
		writeDebugResponse(w, "Cound't decode JSON")
		// writeErrorResponse(w)
		return
	}

	forBytes := common.FromHex(requestJSON.For)
	address := common.Address{}
	copy(address[:], forBytes)
	bigint := types.NewBigInt(0)
	bigint.SetString(requestJSON.Value, 10)
	blockNumber := uint32(requestJSON.BlockNumber)
	transactionNumber := uint32(requestJSON.TransactionNumber)
	outputNumber := uint8(requestJSON.OutputNumber)
	err = h.utxoCreator.InsertUTXO(address, blockNumber, transactionNumber, outputNumber, bigint)
	if err != nil {
		log.Println("Failed to write transaction")
		log.Printf("%+v\n", err)
		writeDebugResponse(w, "Cound't write transaction")
		// writeErrorResponse(w)
		return
	}
	writeSuccessResponse(w)
	return
}
