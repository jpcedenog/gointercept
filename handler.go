// Package gointercept provides primitives to create a middleware layer around AWS Lambda functions
package gointercept

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
)

// The LambdaHandler type represents the signature of the AWS Lambda functions that GoIntercept handles
type LambdaHandler func(context.Context, interface{}) (interface{}, error)

// The ErrorHandler type represents a local function signature used to handle and escalate errors
type ErrorHandler func(context.Context, interface{}, error) (interface{}, error)

// The Interceptor type contains the three potential handlers that can be applied during the Lambda function
// lifecycle. That is, a handler to be executed before, after, an on error of the Lambda function
type Interceptor struct {
	Before  LambdaHandler
	After   LambdaHandler
	OnError ErrorHandler
}

func (interceptor Interceptor) handle(handler LambdaHandler) LambdaHandler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		response := request
		var err error

		if interceptor.Before != nil {
			response, err = interceptor.Before(ctx, request)
			if err != nil {
				return processError(response, interceptor, ctx, err)
			}
		}

		response, err = handler(ctx, response)
		if err != nil {
			return processError(response, interceptor, ctx, err)
		}

		if interceptor.After != nil {
			response, err = interceptor.After(ctx, response)
			if err != nil {
				return processError(response, interceptor, ctx, err)
			}
		}

		return response, err
	}
}

func processError(response interface{}, interceptor Interceptor, ctx context.Context, err error) (interface{}, error) {
	if interceptor.OnError != nil {
		return interceptor.OnError(ctx, response, err)
	} else {
		return response, err
	}
}

// The InterceptedHandler type wraps a LambdaHandler so interceptors can be applied to it
type InterceptedHandler struct {
	handler LambdaHandler
}

// The With method wraps the given handler with the provided interceptors. Interceptors are wrapped in the order
// provided.
//
//That is, the first interceptor's 'Before' handler (if any) is executed first and before everything else.
// The last provided interceptor's 'Before' handler (if any) is executed right before the Lambda function is executed.
// 'After' handlers are executed after the Lambda function execution, in a similar fashion.
func (a *InterceptedHandler) With(adapters ...Interceptor) LambdaHandler {
	handler := a.handler
	last := len(adapters) - 1
	for i := range adapters {
		adapter := adapters[last-i]
		handler = adapter.handle(handler)
	}
	return handler
}

// The This function converts the given Lambda function into an InterceptedHandler
func This(handler interface{}) *InterceptedHandler {
	return &InterceptedHandler{handler: newHandler(handler)}
}

func errorHandler(e error) LambdaHandler {
	return func(ctx context.Context, event interface{}) (interface{}, error) {
		return nil, e
	}
}

func newHandler(handlerFunc interface{}) LambdaHandler {
	if handlerFunc == nil {
		return errorHandler(fmt.Errorf("handler is nil"))
	}
	handler := reflect.ValueOf(handlerFunc)
	handlerType := reflect.TypeOf(handlerFunc)

	if handlerType.Kind() != reflect.Func {
		return errorHandler(fmt.Errorf("handler kind %s is not %s", handlerType.Kind(), reflect.Func))
	}

	takesContext, err := validateArguments(handlerType)
	if err != nil {
		return errorHandler(err)
	}

	if err := validateReturns(handlerType); err != nil {
		return errorHandler(err)
	}

	return func(ctx context.Context, payload interface{}) (interface{}, error) {
		var args []reflect.Value
		if takesContext {
			args = append(args, reflect.ValueOf(ctx))
		}
		if (handlerType.NumIn() == 1 && !takesContext) || handlerType.NumIn() == 2 {
			eventType := handlerType.In(handlerType.NumIn() - 1)
			event := reflect.New(eventType)

			payloadBytes, err := GetBytes(payload)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(payloadBytes, event.Interface()); err != nil {
				if e, ok := err.(*json.SyntaxError); ok {
					log.Printf("Syntax error at byte offset %d\n", e.Offset)
				}
				fmt.Println("Error unmarshalling payload bytes")
				return nil, err
			}
			args = append(args, event.Elem())
		}

		response := handler.Call(args)

		var err error
		if len(response) > 0 {
			if errVal, ok := response[len(response)-1].Interface().(error); ok {
				err = errVal
			}
		}
		var val interface{}
		if len(response) > 1 {
			val = response[0].Interface()
		}

		return val, err
	}
}

func validateArguments(handler reflect.Type) (bool, error) {
	handlerTakesContext := false
	if handler.NumIn() > 2 {
		return false, fmt.Errorf("handlers may not take more than two arguments, but handler takes %d", handler.NumIn())
	} else if handler.NumIn() > 0 {
		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
		argumentType := handler.In(0)
		handlerTakesContext = argumentType.Implements(contextType)
		if handler.NumIn() > 1 && !handlerTakesContext {
			return false, fmt.Errorf("handler takes two arguments, but the first is not Context. got %s", argumentType.Kind())
		}
	}

	return handlerTakesContext, nil
}

func validateReturns(handler reflect.Type) error {
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	switch n := handler.NumOut(); {
	case n > 2:
		return fmt.Errorf("handler may not return more than two values")
	case n > 1:
		if !handler.Out(1).Implements(errorType) {
			return fmt.Errorf("handler returns two values, but the second does not implement error")
		}
	case n == 1:
		if !handler.Out(0).Implements(errorType) {
			return fmt.Errorf("handler returns a single value, but it does not implement error")
		}
	}

	return nil
}
