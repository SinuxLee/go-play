package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var (
	readTimeout, _         = time.ParseDuration("500ms")
	writeTimeout, _        = time.ParseDuration("500ms")
	maxIdleConnDuration, _ = time.ParseDuration("1h")
	client                 = &fasthttp.Client{
		ReadTimeout:                   readTimeout,
		WriteTimeout:                  writeTimeout,
		MaxIdleConnDuration:           maxIdleConnDuration,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		// increase DNS cache time to an hour instead of default minute
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}
)

func Index(ctx *fasthttp.RequestCtx) {
	// buf := bytes.NewBufferString("")
	// buf.Write(ctx.Method())
	// buf.WriteByte(' ')

	// buf.Write(ctx.Path())
	// buf.WriteByte(' ')

	// buf.Write(ctx.URI().QueryString())
	// buf.WriteByte('\n')

	// buf.Write(ctx.Request.Header.Header())
	// buf.WriteByte('\n')

	// buf.Write(ctx.Request.Body())
	// buf.WriteByte('\n')

	// ctx.Write(buf.Bytes())

	data := `{"name": "libz"}`
	reqTimeout := time.Duration(2000) * time.Millisecond
	req := fasthttp.AcquireRequest()
	req.SetRequestURI("http://10.0.84.166:8060/anything/haha")
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBodyRaw([]byte(data))
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	client.DoTimeout(req, resp, reqTimeout)
	defer fasthttp.ReleaseResponse(resp)

	statusCode := resp.StatusCode()
	respBody := resp.Body()

	if statusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "ERR invalid HTTP response code: %d\n", statusCode)
		return
	}

	resp.Header.CopyTo(&ctx.Response.Header)
	ctx.SetContentType("application/json")
	ctx.Write(respBody)
}

func main() {
	r := router.New()
	r.ANY("/{path:*}", Index)

	compressHandler := fasthttp.CompressHandlerLevel(r.Handler, fasthttp.CompressDefaultCompression)
	timeoutHandler := fasthttp.TimeoutHandler(compressHandler, 5*time.Second, "Request timed out")

	log.Printf("%+v", r.List())

	if err := fasthttp.ListenAndServe(":8080", timeoutHandler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
