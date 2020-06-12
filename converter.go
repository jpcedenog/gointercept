package gointercept

import (
	"bytes"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"strings"
)

// Converts the value pointed to by the response parameter into an APIGatewayResponse instance. If the given parameter
// is already an APIGatewayResponse, it is returned as is. Otherwise, a new instance is created and the given parameter
// is attached as part of the body field
func ConvertToAPIGatewayResponse(response interface{}) (events.APIGatewayProxyResponse, error) {
	apiGatewayResponse, ok := response.(events.APIGatewayProxyResponse)

	if !ok {
		apiGatewayResponse = events.APIGatewayProxyResponse{}

		body, ok := response.([]byte)
		if !ok {
			b, err := json.Marshal(response)
			if err != nil {
				return apiGatewayResponse, err
			}
			body = b
		}

		var buf bytes.Buffer
		json.HTMLEscape(&buf, body)
		apiGatewayResponse.Body = buf.String()
	}

	return apiGatewayResponse, nil
}

type input struct {
	Body string `json:"body"`
}

// Returns the contents of the Body field from the given parameter
func GetBody(request interface{}) (string, error) {
	apiGatewayRequest, ok := request.(events.APIGatewayProxyRequest)
	if ok {
		return apiGatewayRequest.Body, nil

	} else {
		bodyBytes, err := GetBytes(request)
		if err != nil {
			return "", err
		}
		var input input
		decoder := json.NewDecoder(strings.NewReader(string(bodyBytes)))
		if err := decoder.Decode(&input); err != nil {
			return "", err
		}
		return input.Body, nil
	}
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(key)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
