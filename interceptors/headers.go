// Package interceptors provides the building blocks of the functionality provided by GoIntercept
// All interceptors, native and custom, should be found under this package
package interceptors

import (
	"context"
	"github.com/jpcedenog/gointercept"
	"github.com/jpcedenog/gointercept/internal"
	"strconv"
	"strings"
)

type hsts struct {
	maxAge            int
	includeSubDomains bool
	preLoad           bool
}

type securityHeaders struct {
	activateDNSPrefetchControl bool
	frameGuardAction           string
	hidePoweredByWith          string
	hsts                       hsts
	activateIENoOpen           bool
	activateNoSniff            bool
	referrerPolicy             string
}

func getDefaults() securityHeaders {
	securityHeaders := securityHeaders{}

	securityHeaders.activateDNSPrefetchControl = true
	securityHeaders.frameGuardAction = "DENY"
	securityHeaders.hidePoweredByWith = ""
	securityHeaders.hsts.maxAge = 180 * 24 * 60 * 60
	securityHeaders.hsts.includeSubDomains = true
	securityHeaders.hsts.preLoad = true
	securityHeaders.activateIENoOpen = true
	securityHeaders.activateNoSniff = true
	securityHeaders.referrerPolicy = "no-referrer"

	return securityHeaders
}

// Option represents a configuration option for the Headers interceptor
type Option func(*securityHeaders)

// DNSPrefetchControl controls browser DNS prefetching
func DNSPrefetchControl(activate bool) Option {
	return func(f *securityHeaders) {
		f.activateDNSPrefetchControl = activate
	}
}

// FrameGuard controls option to prevent clickjacking
func FrameGuard(action string) Option {
	return func(f *securityHeaders) {
		f.frameGuardAction = action
	}
}

// HidePoweredBy controls option to remove the Server/X-Powered-By header
func HidePoweredBy(with string) Option {
	return func(f *securityHeaders) {
		f.hidePoweredByWith = with
	}
}

// HTTPStrictTransportSecurity controls option to set HTTPStrictTransportSecurity
func HTTPStrictTransportSecurity(maxAge int, includeSubDomains bool, preLoad bool) Option {
	return func(f *securityHeaders) {
		f.hsts.maxAge = maxAge
		f.hsts.includeSubDomains = includeSubDomains
		f.hsts.preLoad = preLoad
	}
}

// IENoOpen controls option to set X-Download-Options for IE8+
func IENoOpen(activate bool) Option {
	return func(f *securityHeaders) {
		f.activateIENoOpen = activate
	}
}

// NoSniff controls option to keep clients from sniffing the MIME type
func NoSniff(activate bool) Option {
	return func(f *securityHeaders) {
		f.activateNoSniff = activate
	}
}

// ReferrerPolicy controls option to to hide the Referer header
func ReferrerPolicy(policy string) Option {
	return func(f *securityHeaders) {
		f.referrerPolicy = policy
	}
}

// AddHeaders attaches the given key-value mappings as HTTP headers to the output returned by the Lambda function. It does so
// by wrapping this output with an APIGatewayProxyResponse if necessary
func AddHeaders(headers map[string]string) gointercept.Interceptor {
	return gointercept.Interceptor{
		After: func(ctx context.Context, payload interface{}) (interface{}, error) {
			apiGatewayResponse, e := internal.ConvertToAPIGatewayResponse(payload)

			if e != nil {
				return payload, e
			}

			if apiGatewayResponse.Headers == nil {
				apiGatewayResponse.Headers = make(map[string]string)
			}

			if headers != nil && len(headers) > 0 {
				for k, v := range headers {
					apiGatewayResponse.Headers[k] = v
				}
			}

			return apiGatewayResponse, e
		},
	}
}

// AddSecurityHeaders attaches default HTTP security headers to the output returned by the Lambda function. This is similar to the
// functionality offered by HelmetJS. For more information on the headers added by this interceptor check
// (https://helmetjs.github.io/)
//
// Optionally, this interceptor's behavior can be customized by passing functions to activate, deactivate, or
// modify the functionality of the default headers. These functions include: DNSPrefetchControl, FrameGuard,
// HidePoweredBy, HTTPStrictTransportSecurity, IENoOpen, NoSniff, and ReferrerPolicy.
func AddSecurityHeaders(options ...Option) gointercept.Interceptor {
	securityHeaders := getDefaults()
	for _, opt := range options {
		opt(&securityHeaders)
	}

	headers := make(map[string]string)
	if securityHeaders.activateDNSPrefetchControl {
		headers["X-DNS-Prefetch-Control"] = "on"
	} else {
		headers["X-DNS-Prefetch-Control"] = "off"
	}

	headers["X-Frame-Options"] = securityHeaders.frameGuardAction
	headers["X-Powered-By"] = securityHeaders.hidePoweredByWith

	strictTransportSecurity := []string{strconv.Itoa(securityHeaders.hsts.maxAge)}
	if securityHeaders.hsts.includeSubDomains {
		strictTransportSecurity = append(strictTransportSecurity, "includeSubDomains")
	}
	if securityHeaders.hsts.preLoad {
		strictTransportSecurity = append(strictTransportSecurity, "preLoad")
	}
	headers["Strict-Transport-Security"] = strings.Join(strictTransportSecurity, "; ")

	if securityHeaders.activateIENoOpen {
		headers["X-Download-Options"] = "noopen"
	}

	if securityHeaders.activateNoSniff {
		headers["X-Content-Type-Options"] = "nosniff"
	}

	headers["Referrer-Policy"] = securityHeaders.referrerPolicy

	return AddHeaders(headers)
}
