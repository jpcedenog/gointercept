package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jpcedenog/gointercept"
	"github.com/jpcedenog/gointercept/interceptors"
	"net/http"
	"strings"
	"testing"
)

type Input struct {
	Content string `json:"content"`
	Value   int    `json:"value"`
}

type Output struct {
	Status  string
	Content string
}

const (
	schema = `{
    "$id": "https://qri.io/schema/",
    "$comment" : "sample comment",
    "title": "Input",
    "type": "object",
    "properties": {
        "content": {
            "type": "string"
        },
        "value": {
            "description": "The Value",
            "type": "integer",
            "minimum": 0,
            "maximum": 2
        }
    },
    "required": ["content", "value"]
  }`
)

func simpleFunction(context context.Context, input Input) (*Output, error) {
	if input.Value%2 != 0 {
		return nil, errors.New("Value is not even")
	}

	return &Output{
		Status:  "Function ran successfully!",
		Content: input.Content,
	}, nil
}

func TestParseCustomInput(t *testing.T) {
	request := events.APIGatewayProxyRequest{Body: "{\"content\": \"Random content\", \"value\": 2 }"}

	handler := gointercept.This(simpleFunction).With(
		interceptors.ParseInput(&Input{}, false),
	)

	var response Output
	if err := executeHandler(handler, request, &response); err != nil {
		panic(err)
	}

	if response.Content != "Random content" {
		t.Errorf("Unexpected content '%s' in response", response.Content)
		t.Fail()
	}
}

func TestAPIGatewayRequestResponse(t *testing.T) {
	cases := []struct {
		scenario        string
		request         interface{}
		handler         gointercept.LambdaHandler
		expectedBody    string
		expectedStatus  int
		expectedHeaders *map[string]string
	}{
		{
			scenario: "Successful parsing and output",
			request:  events.APIGatewayProxyRequest{Body: "{\"content\": \"Random content\", \"value\": 2 }"},
			handler: gointercept.This(simpleFunction).With(
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{http.StatusOK, http.StatusBadRequest}),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `{"Status":"Function ran successfully!","Content":"Random content"}`,
			expectedStatus: http.StatusOK,
		},
		{
			scenario: "Successful parsing and headers in output",
			request:  events.APIGatewayProxyRequest{Body: "{\"content\": \"Random content\", \"value\": 2 }"},
			handler: gointercept.This(simpleFunction).With(
				interceptors.AddHeaders(map[string]string{"Content-Type": "application/json", "company-header1": "foo1", "company-header2": "foo2"}),
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: http.StatusOK, Error: http.StatusBadRequest}),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:    `{"Status":"Function ran successfully!","Content":"Random content"}`,
			expectedStatus:  http.StatusOK,
			expectedHeaders: &map[string]string{"Content-Type": "application/json", "company-header1": "foo1", "company-header2": "foo2"},
		},
		{
			scenario: "Successful parsing and error in output",
			request:  events.APIGatewayProxyRequest{Body: "{\"content\": \"Random content\", \"value\": 1 }"},
			handler: gointercept.This(simpleFunction).With(
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: http.StatusOK, Error: http.StatusBadRequest}),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `Value is not even`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			scenario: "Parsing error",
			request:  events.APIGatewayProxyRequest{Body: "{\"foo\": 1}"},
			handler: gointercept.This(simpleFunction).With(
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: http.StatusOK, Error: http.StatusBadRequest}),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `can't parse "{\"foo\": 1}" - json: unknown field "foo"`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			scenario: "JSON Schema validation success",
			request:  events.APIGatewayProxyRequest{Body: `{ "content": "Random content", "value": 2 }`},
			handler: gointercept.This(simpleFunction).With(
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: http.StatusOK, Error: http.StatusBadRequest}),
				interceptors.ValidateJSONSchema(schema),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `{"Status":"Function ran successfully!","Content":"Random content"}`,
			expectedStatus: http.StatusOK,
		},
		{
			scenario: "JSON Schema validation error (Missing attribute)",
			request:  events.APIGatewayProxyRequest{Body: `{ "content": "Random content" }`},
			handler: gointercept.This(simpleFunction).With(
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: http.StatusOK, Error: http.StatusBadRequest}),
				interceptors.ValidateJSONSchema(schema),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `/: {"content":"Random c... "value" value is required`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			scenario: "JSON Schema validation error (Wrong attribute type)",
			request:  events.APIGatewayProxyRequest{Body: `{ "content": "Random content", "value": "30" }`},
			handler: gointercept.This(simpleFunction).With(
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: http.StatusOK, Error: http.StatusBadRequest}),
				interceptors.ValidateJSONSchema(schema),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `/value: "30" type should be integer, got string`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			scenario: "JSON Schema validation error (Value out of range)",
			request:  events.APIGatewayProxyRequest{Body: `{ "content": "Random content", "value": 20 }`},
			handler: gointercept.This(simpleFunction).With(
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: http.StatusOK, Error: http.StatusBadRequest}),
				interceptors.ValidateJSONSchema(schema),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `/value: 20 must be less than or equal to 2.000000`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			scenario: "Successful parsing and default security headers in output",
			request:  events.APIGatewayProxyRequest{Body: `{"content": "Random content", "value": 2 }`},
			handler: gointercept.This(simpleFunction).With(
				interceptors.AddSecurityHeaders(),
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: http.StatusOK, Error: http.StatusBadRequest}),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `{"Status":"Function ran successfully!","Content":"Random content"}`,
			expectedStatus: http.StatusOK,
			expectedHeaders: &map[string]string{"X-DNS-Prefetch-Control": "on",
				"X-Frame-Options":           "DENY",
				"X-Powered-By":              "",
				"Strict-Transport-Security": "15552000; includeSubDomains; preLoad",
				"X-Download-Options":        "noopen",
				"X-Content-Type-Options":    "nosniff",
				"Referrer-Policy":           "no-referrer",
			},
		},
		{
			scenario: "Successful parsing and custom security headers in output",
			request:  events.APIGatewayProxyRequest{Body: `{"content": "Random content", "value": 2 }`},
			handler: gointercept.This(simpleFunction).With(
				interceptors.AddSecurityHeaders(
					interceptors.DNSPrefetchControl(false),
					interceptors.FrameGuard("SAMEORIGIN"),
					interceptors.HidePoweredBy("PHP 4.2.0"),
					interceptors.HTTPStrictTransportSecurity(2, false, false),
					interceptors.IENoOpen(false),
					interceptors.NoSniff(false),
					interceptors.ReferrerPolicy("same-origin"),
				),
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: http.StatusOK, Error: http.StatusBadRequest}),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `{"Status":"Function ran successfully!","Content":"Random content"}`,
			expectedStatus: http.StatusOK,
			expectedHeaders: &map[string]string{"X-DNS-Prefetch-Control": "off",
				"X-Frame-Options":           "SAMEORIGIN",
				"X-Powered-By":              "PHP 4.2.0",
				"Strict-Transport-Security": "2",
				"Referrer-Policy":           "same-origin",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.scenario, func(t *testing.T) {
			var response events.APIGatewayProxyResponse
			if err := executeHandler(c.handler, c.request, &response); err != nil {
				panic(err)
			}

			if response.Body != c.expectedBody {
				t.Errorf("Unexpected content '%s' in response's body", response.Body)
				t.Fail()
			}
			if response.StatusCode != c.expectedStatus {
				t.Errorf("Unexpected status '%d' in response", response.StatusCode)
				t.Fail()
			}
			if c.expectedHeaders != nil {
				if response.Headers == nil {
					t.Errorf("Response was expected to have headers but it has none")
					t.Fail()
				}
				for key, value := range *c.expectedHeaders {
					if response.Headers[key] != value {
						t.Errorf("Expected header '%s: %s' in response not found", key, value)
						t.Fail()
					}
				}
			}
		})
	}
}

func executeHandler(handler gointercept.LambdaHandler, request interface{}, response interface{}) error {
	resp, err := handler(context.TODO(), request)
	if err != nil {
		return err
	}
	if err := decode(resp, &response); err != nil {
		return err
	}

	return nil
}

func decode(input interface{}, response interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(input); err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(string(buf.Bytes())))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		return err
	}
	return nil
}
