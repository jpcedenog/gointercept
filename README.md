<div align="center">
  <img alt="Go Intercept" src="img/GoIntercept.png"/>
  <p><strong>Elegant and modular middleware for AWS Lambdas with Golang</strong></p>
</div>

[![Total alerts](https://img.shields.io/lgtm/alerts/g/jpcedenog/gointercept.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/jpcedenog/gointercept/alerts/)

### About GoIntercept

GoIntercept is a simple but powerful middleware engine that allows you to simplify your AWS Lambda functions when using
 Go.

Web frameworks such as Echo, Spiral, Gin, among others, provide middleware functionality to wrap HTTP requests and
 provide additional features without polluting the core logic in the HTTP handler. GoIntercept aims to provide
  similar behavior but for AWS Lambda functions.

In other words, a middleware layer allows developers to focus on the business logic within the Lambda functions and then
 enhance their behavior by providing additional functionality like authentication/authorization, input validation
 , serialization, etc. in a modular and reusable way.

### Quick Example

The simple example below shows the power of GoIntercept:

```go
package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
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

func SampleFunction(context context.Context, input Input) (*Output, error) {
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
	lambda.Start(gointercept.This(SampleFunction).With(
		interceptors.Notify("Start -1-", "End -1-"),
		interceptors.CreateAPIGatewayProxyResponse(&interceptors.DefaultStatusCodes{Success: 200, Error: 400}),
		interceptors.AddHeaders(map[string]string{"Content-Type": "application/json", "company-header1": "foo1", "company-header2": "foo2"}),
		interceptors.ParseInput(&Input{}, false),
	))
}
```

### Usage

### Dissecting GoIntercept

#### Execution Order

#### Stop and Exit

#### Error Handling

### Custom Middlewares

#### Middlewares Are Just Functions!

#### Inline Middlewares

### Available Middlewares

### Contributing

### License
