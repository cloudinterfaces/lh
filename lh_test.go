package lh

import (
	"encoding/json"
	"net"
	"net/http"
	"net/rpc"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/aws/aws-lambda-go/lambda/messages"
)

func TestInvoke(t *testing.T) {
	s := &service{http: http.DefaultServeMux}
	req := new(messages.InvokeRequest)
	req.Payload = []byte(payload)
	res := new(messages.InvokeResponse)
	err := s.Invoke(req, res)
	if err != nil {
		t.Fatal(err)
	}
	response := events.APIGatewayProxyResponse{}
	if err := json.Unmarshal(res.Payload, &response); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 404 {
		t.Fatalf("Expected 404 StatusCode, got %v", response.StatusCode)
	}
	if !response.IsBase64Encoded {
		t.Fatal("Response not base64 encoded.")
	}
}

func TestRoundTrip(t *testing.T) {
	s := &service{http: http.DefaultServeMux}
	err := rpc.RegisterName("Function", s)
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	go rpc.Accept(l)
	c, err := rpc.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	req := new(messages.InvokeRequest)
	req.Payload = []byte(payload)
	res := new(messages.InvokeResponse)
	err = c.Call("Function.Invoke", req, res)
	if err != nil {
		t.Fatal(err)
	}
	response := events.APIGatewayProxyResponse{}
	if err := json.Unmarshal(res.Payload, &response); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 404 {
		t.Fatalf("Expected 404 StatusCode, got %v", response.StatusCode)
	}
	if !response.IsBase64Encoded {
		t.Fatal("Response not base64 encoded.")
	}
}

const payload = `{
	"body": "{\"test\":\"body\"}",
	"resource": "/{proxy+}",
	"requestContext": {
	  "resourceId": "123456",
	  "apiId": "1234567890",
	  "resourcePath": "/{proxy+}",
	  "httpMethod": "POST",
	  "requestId": "c6af9ac6-7b61-11e6-9a41-93e8deadbeef",
	  "accountId": "123456789012",
	  "identity": {
		"apiKey": null,
		"userArn": null,
		"cognitoAuthenticationType": null,
		"caller": null,
		"userAgent": "Custom User Agent String",
		"user": null,
		"cognitoIdentityPoolId": null,
		"cognitoIdentityId": null,
		"cognitoAuthenticationProvider": null,
		"sourceIp": "127.0.0.1",
		"accountId": null
	  },
	  "stage": "prod"
	},
	"queryStringParameters": {
	  "foo": "bar"
	},
	"headers": {
	  "Via": "1.1 08f323deadbeefa7af34d5feb414ce27.cloudfront.net (CloudFront)",
	  "Accept-Language": "en-US,en;q=0.8",
	  "CloudFront-Is-Desktop-Viewer": "true",
	  "CloudFront-Is-SmartTV-Viewer": "false",
	  "CloudFront-Is-Mobile-Viewer": "false",
	  "X-Forwarded-For": "127.0.0.1, 127.0.0.2",
	  "CloudFront-Viewer-Country": "US",
	  "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	  "Upgrade-Insecure-Requests": "1",
	  "X-Forwarded-Port": "443",
	  "Host": "1234567890.execute-api.us-east-1.amazonaws.com",
	  "X-Forwarded-Proto": "https",
	  "X-Amz-Cf-Id": "cDehVQoZnx43VYQb9j2-nvCh-9z396Uhbp027Y2JvkCPNLmGJHqlaA==",
	  "CloudFront-Is-Tablet-Viewer": "false",
	  "Cache-Control": "max-age=0",
	  "User-Agent": "Custom User Agent String",
	  "CloudFront-Forwarded-Proto": "https",
	  "Accept-Encoding": "gzip, deflate, sdch"
	},
	"pathParameters": {
	  "proxy": "path/to/resource"
	},
	"httpMethod": "POST",
	"stageVariables": {
	  "baz": "qux"
	},
	"path": "/path/to/resource"
  }`
