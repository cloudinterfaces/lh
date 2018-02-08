# lh
A package to serve http.Handlers in the Amazon Lambda Go runtime.

This is a replacement for the [lambda package](github.com/aws/aws-lambda-go/lambda) as published in the aws/aws-lambda-go repository. This package allows serving (many or most) http.Handlers.

## Usage
Call lh.ServeHTTP(handler) from a main package. If handler is nil, http.DefaultServeMux will be served.

## Deployment
The [lago](https://github.com/cloudinterfaces/lago) commandline tool is designed to make deploying Go handlers to Lambda about as easy as it can possibly be.

## Notes
This package only works with API Gateway requests. Undefined behavior may occur if the Lambda function is called with other request types. API Gateway must be configured to treat all content types as binary (```*/*``` is in "Binary Media Types" under the Settings tab). No notion of streaming or connections exists in the API Gateway/Lambda environment, therefore the ResponseWriter may not be asserted to http.Flusher nor http.Hijacker, the CONNECT method cannot be used in a useful way, and other behavior (like Expect headers) are handled upstream. Request and response size is limited by the Lambda environment; it's unlikely more than 10k of request headers, 6 megabytes of request body, or 10 megabytes of response will succeed. Finally, it's unreasonable to expect goroutines created from ServeHTTP methods to continue running after ServeHTTP exits; if something makes assumptions about long-running processes it's probably not a candidate for the Lambda environment.

That said, many packages that implement http.Handler should work fine. For example [Gorilla's RPC](https://github.com/gorilla/rpc) works, and [gin](https://github.com/gin-gonic/gin) works OK when used as an http.Handler. [Revel](https://revel.github.io/) won't but you can't do most of the stuff it does in the Lambda environment anyway.