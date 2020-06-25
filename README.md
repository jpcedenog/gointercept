<div align="center">
  <img alt="Go Intercept" src="img/GoIntercept.png"/>
  <p><strong>Elegant and modular middleware for AWS Lambdas with Golang</strong></p>
</div>

[![Total alerts](https://img.shields.io/lgtm/alerts/g/jpcedenog/gointercept.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/jpcedenog/gointercept/alerts/)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjpcedenog%2Fgointercept.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjpcedenog%2Fgointercept?ref=badge_shield)

### About GoIntercept

GoIntercept is a simple but powerful middleware engine that allows you to simplify your AWS Lambda functions when using
 Go.

Web frameworks such as Echo, Spiral, Gin, among others, provide middleware functionality to wrap HTTP requests and
 provide additional features without polluting the core logic in the HTTP handler. GoIntercept aims to provide
  similar behavior but for AWS Lambda functions.

In other words, a middleware layer allows developers to focus on the business logic within the Lambda functions and then
 enhance their behavior by providing additional functionality like authentication/authorization, input validation
 , serialization, etc. in a modular and reusable way.

### Install

To use GoIntercept simply get it from its repository:

```shell script
go get github.com/jpcedenog/gointercept
```

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

From the quick example above, you may already seen that GoIntercept is extremely easy to wrap around your existing Lambda Handlers. It is designed to get out of your way and remove all the boilerplate related to trivial operations.
 
The steps below describe the process to use GoIntercept:

1. Implement your Lambda Handler.
2. Import the "gointercept" and "gointercept/interceptors" packages.
3. In the main() function, wrap your Lambda handler with the gointercept.This() function.
4. Add all the required interceptors with the .With() method. More interceptors are coming soon! Stay tuned!

### Dissecting GoIntercept

Just like [Middy](https://middy.js.org/), GoIntercept is based on the onion middleware pattern. This means that each interceptor specified in the With() method wraps around the following interceptor on the list or the Lambda Handler itself when the last interceptor is reached.
  
#### Execution Order

The sequence of interceptors, passed to the .With() method, specifies the order in which they are executed. This means that the last interceptor on the list runs just before the Lambda handler is executed . Additionally, each interceptor specifies at least one of three possible execution phases: Before, After, and OnError.

The Before phase runs before the following interceptor on the list, or the Lambda handler itself, runs. Note that in this phase the Lambda handler's response has not been created yet, so you will have access only to the request. 

The After phase runs after the following interceptor on the list, or the Lambda handler itself, has run. Note that in this phase the Lambda handler's response has already been created and is fully available.

As an example, if three middlewares have been specified and each has a Before and After phases, the steps below present the expected execution order:

1. middleware1 (before)
2. middleware2 (before)
3. middleware3 (before)
4. Lambda Handler
5. middleware3 (after)
6. middleware2 (after)
7. middleware1 (after)

#### Error Handling

### Custom Middlewares

#### Inline Middlewares

### Available Middlewares

### Contributing

In the spirit of Open Source Software, everyone is very welcome to contribute to this repository. Feel free to raise issues or to submit Pull Requests.

### License


[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjpcedenog%2Fgointercept.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjpcedenog%2Fgointercept?ref=badge_large)