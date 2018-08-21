package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func writeFasthttpErrorResponse(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	response := sendRawRLPTXResponse{Error: true, Reason: "invalid transaction"}
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}

func writeFasthttpSuccessResponse(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	response := sendRawRLPTXResponse{Error: false, Accepted: true}
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}
