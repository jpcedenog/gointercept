package interceptors

import (
	"context"
	"encoding/json"
	"github.com/jpcedenog/gointercept"
	"github.com/jpcedenog/gointercept/internal"
	"github.com/qri-io/jsonschema"
	"net/http"
)

// ValidateBodyJSONSchema validates the given payload (in JSON format) against the given JSON schema.
//
// For more information check: https://github.com/qri-io/jsonschema
func ValidateBodyJSONSchema(schema string) gointercept.Interceptor {
	return gointercept.Interceptor{
		Before: func(ctx context.Context, payload interface{}) (interface{}, error) {
			body, err := internal.GetBody(payload)
			if err != nil {
				return payload, err
			}

			rs := &jsonschema.Schema{}
			if err := json.Unmarshal([]byte(schema), rs); err != nil {
				return payload, err
			}

			errs, err := rs.ValidateBytes(ctx, []byte(body))
			if err != nil {
				return payload, err
			}

			if len(errs) > 0 {
				return payload, errs[0]
			}

			return payload, nil
		},
		OnError: func(ctx context.Context, payload interface{}, err error) (interface{}, error) {
			return payload, &HTTPError{http.StatusUnprocessableEntity, err.Error()}
		},
	}
}
