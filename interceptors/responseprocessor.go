package interceptors

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jpcedenog/gointercept"
)

type DefaultStatusCodes struct {
	Success int
	Error   int
}

// The CreateAPIGatewayProxyResponse wraps the output of the Lambda function with an APIGatewayProxyResponse
// instance
func CreateAPIGatewayProxyResponse(defaultStatusCode *DefaultStatusCodes) gointercept.Interceptor {
	return gointercept.Interceptor{
		After: func(ctx context.Context, payload interface{}) (interface{}, error) {
			response, err := gointercept.ConvertToAPIGatewayResponse(payload)
			if err != nil {
				return payload, err
			}
			if response.StatusCode == 0 && defaultStatusCode != nil {
				response.StatusCode = defaultStatusCode.Success
			}

			return response, nil
		},
		OnError: func(ctx context.Context, payload interface{}, err error) (interface{}, error) {
			response, e := gointercept.ConvertToAPIGatewayResponse(payload)
			if e != nil {
				response = events.APIGatewayProxyResponse{}
				response.Body = e.Error()
				return payload, nil
			}
			if response.StatusCode == 0 && defaultStatusCode != nil {
				response.StatusCode = defaultStatusCode.Error
			}

			response.Body = err.Error()
			return response, nil
		},
	}
}
