package access

import "github.com/valyala/fasthttp"

func Entry(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		h(ctx)

		// process ctx
	}
}
