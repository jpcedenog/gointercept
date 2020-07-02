package interceptors

import (
	"context"
	"github.com/jpcedenog/gointercept"
	"log"
)

// Logs the given string parameters before and after the execution of the Lambda function
func Notify(beforeMessage, afterMessage string) gointercept.Interceptor {
	return gointercept.Interceptor{
		Before: func(ctx context.Context, payload interface{}) (interface{}, error) {
			log.Println(beforeMessage)
			return payload, nil
		},
		After: func(ctx context.Context, payload interface{}) (interface{}, error) {
			log.Println(afterMessage)
			return payload, nil
		},
	}
}
