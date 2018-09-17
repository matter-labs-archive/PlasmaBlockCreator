package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/matterinc/PlasmaCommons/block"
	"github.com/valyala/fasthttp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	common "github.com/ethereum/go-ethereum/common"
	redis "github.com/go-redis/redis"
	"github.com/shamatar/go-plasma/foundationdb"
)

type assmebleBlockRequest struct {
	// BlockNumber       int    `json:"blockNumber"`
	BlockNumber       string `json:"blockNumber"`
	PreviousBlockHash string `json:"previousBlockHash"`
	StartNext         bool   `json:"startNext"`
}

type assembleBlockResponse struct {
	Error           bool   `json:"error"`
	SerializedBlock string `json:"serializedBlock,omitempty"`
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

func (h *AssembleBlockHandler) HandlerFunc(ctx *fasthttp.RequestCtx) {
	var requestJSON assmebleBlockRequest
	err := json.Unmarshal(ctx.PostBody(), &requestJSON)
	if err != nil {
		writeBlockAssemblyResponse(ctx, true, []byte{})
		return
	}
	previousHash := common.FromHex(requestJSON.PreviousBlockHash)
	if len(previousHash) != block.PreviousBlockHashLength {
		writeBlockAssemblyResponse(ctx, true, []byte{})
		return
	}
	// newBlockNumber := uint32(requestJSON.BlockNumber)
	bn, _ := strconv.ParseUint(requestJSON.BlockNumber, 10, 32)
	newBlockNumber := uint32(bn)
	startNext := requestJSON.StartNext
	block, err := h.blockAssembler.AssembleBlock(newBlockNumber, previousHash, startNext)
	if err != nil || block == nil {
		writeBlockAssemblyResponse(ctx, true, []byte{})
		return
	}
	err = block.Sign(h.signingKey)
	if err != nil {
		writeBlockAssemblyResponse(ctx, true, []byte{})
		return
	}
	rawBlock, err := block.Serialize()
	if err != nil || rawBlock == nil {
		writeBlockAssemblyResponse(ctx, true, []byte{})
		return
	}
	writeBlockAssemblyResponse(ctx, false, rawBlock)
	// response := assembleBlockResponse{Error: false, SerializedBlock: common.ToHex(rawBlock)}
	// ctx.SetContentType("application/json")
	// ctx.SetStatusCode(fasthttp.StatusOK)
	// body, _ := json.Marshal(response)
	// ctx.SetBody(body)
	return
}

func writeBlockAssemblyResponse(ctx *fasthttp.RequestCtx, errorResult bool, rawBlock []byte) {
	response := assembleBlockResponse{Error: errorResult, SerializedBlock: common.ToHex(rawBlock)}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}
