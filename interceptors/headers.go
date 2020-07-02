package interceptors

import (
	"context"
	"github.com/jpcedenog/gointercept"
	"github.com/jpcedenog/gointercept/internal"
)

// Attaches the given headers to the output returned by the Lambda function. It does so
// by wrapping this output with an APIGatewayProxyResponse if necessary
func AddHeaders(headers map[string]string) gointercept.Interceptor {
	return gointercept.Interceptor{
		After: func(ctx context.Context, payload interface{}) (interface{}, error) {
			apiGatewayResponse, e := internal.ConvertToAPIGatewayResponse(payload)
			if e == nil {
				if apiGatewayResponse.Headers == nil {
					apiGatewayResponse.Headers = make(map[string]string)
				}

				if headers != nil && len(headers) > 0 {
					for k, v := range headers {
						apiGatewayResponse.Headers[k] = v
					}
				}
			}
			return apiGatewayResponse, e
		},
	}
}
