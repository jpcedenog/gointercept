package interceptors

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jpcedenog/gointercept"
	"strings"
)

var exceptionsMap = getExceptionsMap([]string{"ALPN", "C-PEP", "C-PEP-Info", "CalDAV-Timezones", "Content-ID",
	"Content-MD5", "DASL", "DAV", "DNT", "ETag", "GetProfile", "HTTP2-Settings", "Last-Event-ID", "MIME-Version",
	"Optional-WWW-Authenticate", "Sec-WebSocket-Accept", "Sec-WebSocket-Extensions", "Sec-WebSocket-Key",
	"Sec-WebSocket-Protocol", "Sec-WebSocket-Version", "SLUG", "TCN", "TE", "TTL", "WWW-Authenticate",
	"X-ATT-DeviceId", "X-DNSPrefetch-Control", "X-UIDH"})

// NormalizeHTTPRequestHeaders captures the headers (single and multi-value) sent in the API Gateway (HTTP) request and
// normalizes them to either an all-lowercase form or to their canonical form (content-type as opposed to Content-Type)
// based on the value of the given 'canonical' parameter.
func NormalizeHTTPRequestHeaders(canonical bool) gointercept.Interceptor {
	return gointercept.Interceptor{
		Before: func(context context.Context, payload interface{}) (interface{}, error) {
			if apiGatewayRequest, ok := payload.(events.APIGatewayProxyRequest); ok {
				if apiGatewayRequest.Headers != nil {
					for key, value := range apiGatewayRequest.Headers {
						apiGatewayRequest.Headers[normalizeKey(key, canonical)] = value
					}
				}

				if apiGatewayRequest.MultiValueHeaders != nil {
					for key, values := range apiGatewayRequest.MultiValueHeaders {
						apiGatewayRequest.MultiValueHeaders[normalizeKey(key, canonical)] = values
					}
				}

				return apiGatewayRequest, nil
			}

			return payload, nil
		},
	}
}

func getExceptionsMap(exceptions []string) map[string]string {
	exceptionsMap := make(map[string]string)
	for _, e := range exceptions {
		exceptionsMap[strings.ToLower(e)] = e
	}

	return exceptionsMap
}

func normalizeKey(key string, canonical bool) string {
	lowerKey := strings.ToLower(key)
	if canonical {
		if value, ok := exceptionsMap[lowerKey]; ok {
			return value
		}
		return strings.Title(lowerKey)
	}
	return lowerKey
}
