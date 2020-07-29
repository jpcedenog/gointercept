package interceptors

import (
	"context"
	"github.com/jpcedenog/gointercept"
	"github.com/jpcedenog/gointercept/internal"
)

// DefaultStatusCodes specifies the default return codes that will be used for successful and
// unsuccessful responses
type DefaultStatusCodes struct {
	Success int
	Error   int
}

// HTTPError represents a generic HTTP error. Specifies the error's code and status text that are used
// when creating and sending proper HTTP responses to the API consumer.
//
// Optionally, an interceptor can throw this type of error with the corresponding code and status text.
// Then, the CreateAPIGatewayProxyResponse interceptor will create the appropriate API Gateway response
// used the information provided by it
type HTTPError struct {
	StatusCode int
	StatusText string
}

func (e *HTTPError) Error() string {
	return e.StatusText
}

// CreateAPIGatewayProxyResponse wraps the output of the Lambda function with an APIGatewayProxyResponse instance
func CreateAPIGatewayProxyResponse(defaultStatusCode *DefaultStatusCodes) gointercept.Interceptor {
	return gointercept.Interceptor{
		After: func(ctx context.Context, payload interface{}) (interface{}, error) {
			response, err := internal.ConvertToAPIGatewayResponse(payload)
			if err != nil {
				return payload, err
			}
			if response.StatusCode == 0 && defaultStatusCode != nil {
				response.StatusCode = defaultStatusCode.Success
			}

			return response, nil
		},
		OnError: func(ctx context.Context, payload interface{}, err error) (interface{}, error) {
			response, e := internal.ConvertToAPIGatewayResponse(payload)
			if e != nil {
				return payload, e
			}
			if httpError, ok := err.(*HTTPError); ok {
				response.Body = httpError.StatusText
				response.StatusCode = httpError.StatusCode
				return response, nil
			}

			response.Body = err.Error()
			response.StatusCode = defaultStatusCode.Error

			return payload, err
		},
	}
}
