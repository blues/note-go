// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package api

import (
	"io"
	"net/http"
)

// ErrorResponse v1
//
// The structure returned from HTTPS API calls when there is an error.
type ErrorResponse struct {
	// Error represents the human readable error message.
	//
	// required: true
	// type: string
	Error string `json:"err"`

	// Code represents the standard status code
	//
	// required: true
	// type: int
	Code int `json:"code"`

	// Status is the machine readable string representation of the error code.
	//
	// required: true
	// type: string
	Status string `json:"status"`

	// Request is the request that was made that resulted in error. The url path would be sufficient.
	//
	// required: false
	// type: string
	Request string `json:"request,omitempty"`

	// Details are any additional information about the request that would be nice to in the response.
	// The request body would be nice especially if there are a lot of parameters.
	//
	// required: false
	// type: object
	Details map[string]interface{} `json:"details,omitempty"`

	// Debug is any other nice to have information to aid in debugging.
	//
	// required: false
	// type: string
	Debug string `json:"debug,omitempty"`
}

var SuspendedBillingAccountResponse = ErrorResponse{
	Code:   403,
	Status: "Forbidden",
	Error:  "this billing account is suspended",
}

// WithRequest is a an easy way to add http.Request information to an error.
// It takes a http.Request object, parses the URI string into response.Request
// and adds the request Body (if it exists) into the response.Details["body"] as a string
func (e ErrorResponse) WithRequest(r *http.Request) ErrorResponse {
	e.Request = r.RequestURI
	var bodyBytes []byte
	if r.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			return e
		}
	}
	if len(e.Details) == 0 {
		e.Details = make(map[string]interface{})
	}
	e.Details["body"] = string(bodyBytes)
	return e
}

// WithError adds an error string from an error object into the response.
func (e ErrorResponse) WithError(err error) ErrorResponse {
	e.Error = err.Error()
	return e
}
