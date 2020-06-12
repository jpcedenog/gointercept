package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jpcedenog/gointercept"
	"github.com/jpcedenog/gointercept/interceptors"
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

func simpleFunction(context context.Context, input Input) (*Output, error) {
	if input.Value%2 != 0 {
		return nil, errors.New("incorrect parameter")
	} else {
		return &Output{
			Status:  "Function ran successfully!",
			Content: input.Content,
		}, nil
	}
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
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: 200, Error: 400}),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `{"Status":"Function ran successfully!","Content":"Random content"}`,
			expectedStatus: 200,
		},
		{
			scenario: "Successful parsing and headers in output",
			request:  events.APIGatewayProxyRequest{Body: "{\"content\": \"Random content\", \"value\": 2 }"},
			handler: gointercept.This(simpleFunction).With(
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: 200, Error: 400}),
				interceptors.AddHeaders(map[string]string{"Content-Type": "application/json", "company-header1": "foo1", "company-header2": "foo2"}),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:    `{"Status":"Function ran successfully!","Content":"Random content"}`,
			expectedStatus:  200,
			expectedHeaders: &map[string]string{"Content-Type": "application/json", "company-header1": "foo1", "company-header2": "foo2"},
		},
		{
			scenario: "Successful parsing and error in output",
			request:  events.APIGatewayProxyRequest{Body: "{\"content\": \"Random content\", \"value\": 1 }"},
			handler: gointercept.This(simpleFunction).With(
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: 200, Error: 400}),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `incorrect parameter`,
			expectedStatus: 400,
		},
		{
			scenario: "Parsing error",
			request:  events.APIGatewayProxyRequest{Body: "{\"foo\": 1}"},
			handler: gointercept.This(simpleFunction).With(
				interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: 200, Error: 400}),
				interceptors.ParseInput(&Input{}, false)),
			expectedBody:   `can't parse "{\"foo\": 1}" - json: unknown field "foo"`,
			expectedStatus: 400,
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
