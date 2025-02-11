// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package api

import "net/http"

// ErrNotFound returns the default for an HTTP 404 NotFound
func ErrNotFound() ErrorResponse {
	return ErrorResponse{
		Status: http.StatusText(http.StatusNotFound),
		Error:  "The requested resource could not be found",
		Code:   http.StatusNotFound,
	}
}

// ErrUnauthorized returns the default for an HTTP 401 Unauthorized
func ErrUnauthorized() ErrorResponse {
	return ErrorResponse{
		Status: http.StatusText(http.StatusUnauthorized),
		Error:  "The request could not be authorized",
		Code:   http.StatusUnauthorized,
	}
}

// ErrForbidden returns the default for an HTTP 403 Forbidden
func ErrForbidden() ErrorResponse {
	return ErrorResponse{
		Status: http.StatusText(http.StatusForbidden),
		Error:  "The requested action was forbidden",
		Code:   http.StatusForbidden,
	}
}

// ErrMethodNotAllowed returns the default for an HTTP 405 Method Not Allowed
func ErrMethodNotAllowed() ErrorResponse {
	return ErrorResponse{
		Status: http.StatusText(http.StatusMethodNotAllowed),
		Error:  "Method not allowed on this endpoint",
		Code:   http.StatusMethodNotAllowed,
	}
}

// ErrInternalServerError returns the default for an HTTP 500 InternalServerError
func ErrInternalServerError() ErrorResponse {
	return ErrorResponse{
		Status: http.StatusText(http.StatusInternalServerError),
		Error:  "An internal server error occurred",
		Code:   http.StatusInternalServerError,
	}
}

// ErrEventsQueryTimeout returns the default for a GetEvents (and related) request that took too long
func ErrEventsQueryTimeout() ErrorResponse {
	return ErrorResponse{
		Status: "Took too long",
		Error:  "Events query took too long to complete",
		Code:   http.StatusGatewayTimeout,
	}
}

// ErrBadRequest returns the default for an HTTP 400 BadRequest
func ErrBadRequest() ErrorResponse {
	return ErrorResponse{
		Status: http.StatusText(http.StatusBadRequest),
		Error:  "The request was malformed or contained invalid parameters",
		Code:   http.StatusBadRequest,
	}
}

// ErrUnsupportedMediaType returns the default for an HTTP 415 UnsupportedMediaType
func ErrUnsupportedMediaType() ErrorResponse {
	return ErrorResponse{
		Status: http.StatusText(http.StatusUnsupportedMediaType),
		Error:  "The request is using an unknown content type",
		Code:   http.StatusUnsupportedMediaType,
	}
}

// ErrConflict returns the default for an HTTP 409 Conflict
func ErrConflict() ErrorResponse {
	return ErrorResponse{
		Status: http.StatusText(http.StatusConflict),
		Error:  "The resource could not be created due to a conflict",
		Code:   http.StatusConflict,
	}
}
