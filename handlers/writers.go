package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type generalErrorResponse struct {
	Error  bool   `json:"error"`
	Reason string `json:"reason,omitempty"`
}

func writeGeneralErrorResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set"Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set"Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	ctx.Response.Header.Set"Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	response := generalErrorResponse{Error: true}
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}

func writeFasthttpErrorResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set"Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set"Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	ctx.Response.Header.Set"Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	response := sendRawRLPTXResponse{Error: true, Reason: "invalid transaction"}
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}

func writeFasthttpSuccessResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set"Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set"Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	ctx.Response.Header.Set"Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	response := sendRawRLPTXResponse{Error: false, Accepted: true}
	body, _ := json.Marshal(response)
	ctx.SetBody(body)
}
