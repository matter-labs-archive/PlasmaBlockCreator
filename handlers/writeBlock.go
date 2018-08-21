package handlers

import (
	"github.com/shamatar/go-plasma/block"
	"github.com/valyala/fasthttp"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/shamatar/go-plasma/foundationdb"
)

type WriteBlockHandler struct {
	db     *fdb.Database
	writer *foundationdb.BlockWriter
}

func NewWriteBlockHandler(db *fdb.Database) *WriteBlockHandler {
	writer := foundationdb.NewBlockWriter(db)
	handler := &WriteBlockHandler{db, writer}
	return handler
}

func (h *WriteBlockHandler) HandlerFunc(ctx *fasthttp.RequestCtx) {
	rawBlock := ctx.PostBody()
	if len(rawBlock) == 0 {
		writeFasthttpErrorResponse(ctx)
		return
	}
	block, err := block.NewBlockFromBytes(rawBlock)
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	err = h.writer.WriteBlock(*block)
	if err != nil {
		writeFasthttpErrorResponse(ctx)
		return
	}
	writeFasthttpSuccessResponse(ctx)
	return
}
