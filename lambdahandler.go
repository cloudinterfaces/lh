// Package lh enables http.Handlers to be served
// in the AWS Lambda Go runtime.
/*
ServeHTTP(h) should be called from a main package. If
h is nil, http.DefaultServeMux will be used.

Note this package is intended for Lambda functions
as API Gateway endpoints; other payloads will have undefined behavior.
Also note API Gateway must be configured so all media types are
treated as binary. That is, "Binary Media Types" under "Settings" includes:*/
//	*/*
/*
API Gateway can remap hostnames and paths. For example using an API Gateway domain
(not a custom domain), the stage component of the request URL is
elided from the path provided to Lambda handlers. This package
inspects for Location headers that would be otherwise incorrect
and attempts to correct them if FixRelativeRedirect is left
at its default of true.

The Lambda environment can be inspected by asserting
http.Responsewriter to the Lambda interface. The ResponseWriter
may not be asserted to an http.Hijacker nor Flusher, as the Lambda
environment does not support any form of streaming (including but not
limited to websockets and EventSources).
*/
package lh

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/rpc"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

// FixRelativeRedirect attemtps to "fix"
// relative redirects when API gateway
// is not configured with a custom domain.
// Setting this to false prevents
// modification of Location headers.
var FixRelativeRedirect = true

// PanicMessage is returned as 
// the body of the 500 HTTP response
// when unrecovered panics occur.
var PanicMessage = "Function panic"

// Lambda is an interface to retrieve Lambda-specific
// types. A Lambda http.ResponseWriter supports
// this type assertion.
type Lambda interface {
	InvokeRequest() messages.InvokeRequest
	ClientContext() lambdacontext.ClientContext
	GatewayRequest() events.APIGatewayProxyRequest
}

var _ = Lambda(&responsewriter{})

type responsewriter struct {
	statusCode int
	io.Writer
	header http.Header
	req    messages.InvokeRequest
	prox   events.APIGatewayProxyRequest
}

func (r *responsewriter) Close() error {
	return nil
}

func (r *responsewriter) InvokeRequest() messages.InvokeRequest {
	return r.req
}

func (r *responsewriter) GatewayRequest() events.APIGatewayProxyRequest {
	return r.prox
}

func (r *responsewriter) ClientContext() lambdacontext.ClientContext {
	ctx := lambdacontext.ClientContext{}
	json.Unmarshal(r.req.ClientContext, &ctx)
	return ctx
}

func (r *responsewriter) Header() http.Header {
	return r.header
}

func (r *responsewriter) WriteHeader(code int) {
	r.statusCode = code
}

// Service implements the Lambda Function API.
type Service struct {
	http http.Handler
}

// Ping is the Lambda keepalive.
func (s *Service) Ping(req *messages.PingRequest, response *messages.PingResponse) error {
	*response = messages.PingResponse{}
	return nil
}

func (s *Service) invokehttp(req *messages.InvokeRequest, res *messages.InvokeResponse) error {
	defer func() {
		if err := recover(); err != nil {
			pr := events.APIGatewayProxyResponse{
				StatusCode:      500,
				Headers:         map[string]string{"Content-Type": "text/plain"},
				Body:            base64.StdEncoding.EncodeToString([]byte(PanicMessage)),
				IsBase64Encoded: true,
			}
			res.Payload, _ = json.Marshal(pr)
			log.Printf("panic: %v\n", err)
			log.Println(string(debug.Stack()))
		}
	}()
	ctx, nonsense := context.WithDeadline(context.Background(), time.Unix(req.Deadline.Seconds, req.Deadline.Nanos).UTC())
	defer nonsense()
	e := events.APIGatewayProxyRequest{}
	err := json.Unmarshal(req.Payload, &e)
	if err != nil {
		return err
	}
	sr := ioutil.NopCloser(strings.NewReader(e.Body))
	if e.IsBase64Encoded {
		sr = ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, sr))
	}
	var qs string
	if len(e.QueryStringParameters) > 0 {
		vals := url.Values{}
		for k, v := range e.QueryStringParameters {
			vals.Set(k, v)
		}
		qs = fmt.Sprintf("?%s", vals.Encode())
	}
	u := fmt.Sprintf("%s%s", e.Path, qs)
	r := httptest.NewRequest(e.HTTPMethod, u, sr).WithContext(ctx)
	for k, v := range e.Headers {
		r.Header.Set(k, v)
	}
	buf := new(bytes.Buffer)
	rw := &responsewriter{
		Writer: base64.NewEncoder(base64.StdEncoding, buf),
		header: http.Header{},
		req:    *req,
		prox:   e,
	}
	s.http.ServeHTTP(rw, r)
	rw.Writer.(io.Closer).Close()
	c := rw.statusCode
	if c == 0 {
		c = http.StatusOK
	}
	if FixRelativeRedirect && c > 300 && c < 400 && strings.HasSuffix(e.Headers["Host"], ".amazonaws.com") {
		loc := rw.header.Get("Location")
		if strings.HasPrefix(loc, "/") {
			loc = fmt.Sprintf("/%s%s", e.RequestContext.Stage, loc)
			rw.header.Set("Location", loc)
		}
	}
	headers := map[string]string{}
	for k := range rw.header {
		headers[k] = rw.header.Get(k)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "text/plain"
	}
	pr := events.APIGatewayProxyResponse{
		StatusCode:      c,
		Headers:         headers,
		Body:            buf.String(),
		IsBase64Encoded: true,
	}
	res.Payload, err = json.Marshal(pr)
	return err
}

// Invoke is the Lambda RPC call.
func (s *Service) Invoke(req *messages.InvokeRequest, response *messages.InvokeResponse) error {
	return s.invokehttp(req, response)
}

// ServeHTTP serves h for the AWS Lambda Go
// runtime. If h is nil, serves http.DefaultServeMux.
// If called outside the Lambda environment,
// starts a normal http.Server on localhost on
// a random port.
func ServeHTTP(h http.Handler) {
	if h == nil {
		h = http.DefaultServeMux
	}
	port := os.Getenv("_LAMBDA_SERVER_PORT")
	if len(port) == 0 {
		port = "0"
	}
	l, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal(err)
	}
	if port == "0" {
		log.Printf("Starting test server on %s", l.Addr().String())
		http.Serve(l, h)
		return
	}
	err = rpc.RegisterName("Function", &Service{http: h})
	if err != nil {
		log.Fatal("failed to register handler function")
	}
	rpc.Accept(l)
	log.Fatal("accept should not have returned")
}
