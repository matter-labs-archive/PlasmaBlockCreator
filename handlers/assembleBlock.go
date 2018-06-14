package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bankex/go-plasma/block"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bankex/go-plasma/foundationdb"
	common "github.com/ethereum/go-ethereum/common"
	redis "github.com/go-redis/redis"
)

type assmebleBlockRequest struct {
	// BlockNumber       int    `json:"blockNumber"`
	BlockNumber       string `json:"blockNumber"`
	PreviousBlockHash string `json:"previousBlockHash"`
	StartNext         bool   `json:"startNext"`
}

type assembleBlockResponse struct {
	Error           bool   `json:"error"`
	SerializedBlock string `json:"previousBlockHash,omitempty"`
}

type AssembleBlockHandler struct {
	db             *fdb.Database
	redisClient    *redis.Client
	blockAssembler *foundationdb.BlockAssembler
	signingKey     []byte
}

func NewAssembleBlockHandler(db *fdb.Database, redisClient *redis.Client, signingKey []byte) *AssembleBlockHandler {
	creator := foundationdb.NewBlockAssembler(db, redisClient)
	handler := &AssembleBlockHandler{db, redisClient, creator, signingKey}
	return handler
}

func (h *AssembleBlockHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var requestJSON assmebleBlockRequest
	err := json.NewDecoder(r.Body).Decode(&requestJSON)
	if err != nil {
		writeErrorResponse(w)
		return
	}
	previousHash := common.FromHex(requestJSON.PreviousBlockHash)
	if len(previousHash) != block.PreviousBlockHashLength {
		writeErrorResponse(w)
		return
	}
	// newBlockNumber := uint32(requestJSON.BlockNumber)
	bn, _ := strconv.ParseUint(requestJSON.BlockNumber, 10, 32)
	newBlockNumber := uint32(bn)
	startNext := requestJSON.StartNext
	block, err := h.blockAssembler.AssembleBlock(newBlockNumber, previousHash, startNext)
	if err != nil || block == nil {
		writeErrorResponse(w)
		return
	}
	err = block.Sign(h.signingKey)
	if err != nil {
		writeErrorResponse(w)
		return
	}
	rawBlock, err := block.Serialize()
	if err != nil || rawBlock == nil {
		writeErrorResponse(w)
		return
	}
	response := assembleBlockResponse{Error: false, SerializedBlock: common.ToHex(rawBlock)}
	json.NewEncoder(w).Encode(response)
	return
}
