// Package interceptors provides the building blocks of the functionality provided by GoIntercept
// All interceptors, native and custom, should be found under this package
package interceptors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jpcedenog/gointercept"
	"github.com/jpcedenog/gointercept/internal"
	"net/http"
	"strings"
)

// ParseInput parses the Lambda function's payload into the value pointed to by the input parameter
func ParseInput(input interface{}, allowUnknownFields bool) gointercept.Interceptor {
	var localPayload interface{}
	return gointercept.Interceptor{
		Before: func(ctx context.Context, payload interface{}) (interface{}, error) {
			body, err := internal.GetBody(payload)
			if err != nil {
				return payload, err
			}
			localPayload = body
			decoder := json.NewDecoder(strings.NewReader(body))
			if !allowUnknownFields {
				decoder.DisallowUnknownFields()
			}
			if err := decoder.Decode(input); err != nil {
				return payload, fmt.Errorf("can't parse %#v - %w", localPayload, err)
			}

			return input, nil
		},
		OnError: func(ctx context.Context, payload interface{}, err error) (interface{}, error) {
			return payload, &HTTPError{http.StatusUnprocessableEntity, err.Error()}
		},
	}
}
