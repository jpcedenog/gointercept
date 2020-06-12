package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jpcedenog/gointercept"
	"github.com/jpcedenog/gointercept/interceptors"
	"log"
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
	log.Printf("Executing main content with %#v\n", input)
	if input.Value%2 != 0 {
		return nil, errors.New("passed incorrect value")
	} else {
		return &Output{
			Status:  "Function ran successfully!",
			Content: input.Content,
		}, nil
	}
}

func main() {
	handler := gointercept.This(simpleFunction).With(
		interceptors.Notify("Start -1-", "End -1-"),
		interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: 200, Error: 400}),
		interceptors.AddHeaders(map[string]string{"Content-Type": "application/json", "company-header1": "foo1", "company-header2": "foo2"}),
		interceptors.ParseInput(&Input{}, false),
	)

	request := events.APIGatewayProxyRequest{Body: "{\"content\": \"Random content\", \"value\": 1 }"}
	//request := events.APIGatewayProxyRequest{Body: "{\"foo\": 1}"}

	response, err := handler(context.TODO(), request)
	if err != nil {
		panic(err)
	}

	prettyJson, _ := json.MarshalIndent(response, "", "  ")
	fmt.Printf("\nResponse:\n%s\n", string(prettyJson))
}

