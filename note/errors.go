// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

// Package note errors.go contains programmatically-testable error strings
package note

import (
	"fmt"
	"net/http"
	"strings"
)

// ErrTimeout (golint)
const ErrTimeout = "{timeout}"

var _ = defineError(ErrTimeout, http.StatusRequestTimeout)

// ErrInternalTimeout of a notehub-to-notehub transaction (golint)
const ErrInternalTimeout = "{internal-timeout}"

var _ = defineError(ErrInternalTimeout, http.StatusGatewayTimeout)

// ErrRouteTimeout of a notehub-to-customer-service transaction (golint)
const ErrRouteTimeout = "{route-timeout}"

var _ = defineError(ErrRouteTimeout, http.StatusRequestTimeout)

// ErrClosed (golint)
const ErrClosed = "{closed}"

var _ = defineError(ErrClosed, http.StatusGone)

// ErrFileNoExist (golint)
const ErrFileNoExist = "{file-noexist}"

var _ = defineError(ErrFileNoExist, http.StatusNotFound)

// ErrNotefileName (golint)
const ErrNotefileName = "{notefile-bad-name}"

var _ = defineError(ErrNotefileName, http.StatusBadRequest)

// ErrNotefileInUse (golint)
const ErrNotefileInUse = "{notefile-in-use}"

var _ = defineError(ErrNotefileInUse, http.StatusConflict)

// ErrNotefileExists (golint)
const ErrNotefileExists = "{notefile-exists}"

var _ = defineError(ErrNotefileExists, http.StatusConflict)

// ErrNotefileNoExist (golint)
const ErrNotefileNoExist = "{notefile-noexist}"

var _ = defineError(ErrNotefileNoExist, http.StatusNotFound)

// ErrNotefileQueueDisallowed (golint)
const ErrNotefileQueueDisallowed = "{notefile-queue-disallowed}"

var _ = defineError(ErrNotefileQueueDisallowed, http.StatusBadRequest)

// ErrNoteNoExist (golint)
const ErrNoteNoExist = "{note-noexist}"

var _ = defineError(ErrNoteNoExist, http.StatusNotFound)

// ErrNoteExists (golint)
const ErrNoteExists = "{note-exists}"

var _ = defineError(ErrNoteExists, http.StatusConflict)

// ErrTooManyNotes (golint)
const ErrTooManyNotes = "{too-many-notes}"

var _ = defineError(ErrTooManyNotes, http.StatusBadRequest)

// ErrTrackerNoExist (golint)
const ErrTrackerNoExist = "{tracker-noexist}"

var _ = defineError(ErrTrackerNoExist, http.StatusNotFound)

// ErrTrackerExists (golint)
const ErrTrackerExists = "{tracker-exists}"

var _ = defineError(ErrTrackerExists, http.StatusConflict)

// ErrNetwork (golint)
const ErrNetwork = "{network}"

var _ = defineError(ErrNetwork, http.StatusServiceUnavailable)

// ErrRegistrationFailure (golint)
const ErrRegistrationFailure = "{registration-failure}"

var _ = defineError(ErrRegistrationFailure, http.StatusServiceUnavailable)

// ErrExtendedNetworkFailure (golint)
const ErrExtendedNetworkFailure = "{extended-network-failure}"

var _ = defineError(ErrExtendedNetworkFailure, http.StatusServiceUnavailable)

// ErrExtendedServiceFailure (golint)
const ErrExtendedServiceFailure = "{extended-service-failure}"

var _ = defineError(ErrExtendedServiceFailure, http.StatusServiceUnavailable)

// ErrHostUnreachable (golint)
const ErrHostUnreachable = "{host-unreachable}"

var _ = defineError(ErrHostUnreachable, http.StatusServiceUnavailable)

// ErrDFUNotReady (golint)
const ErrDFUNotReady = "{dfu-not-ready}"

var _ = defineError(ErrDFUNotReady, http.StatusServiceUnavailable)

// ErrDFUInProgress (golint)
const ErrDFUInProgress = "{dfu-in-progress}"

var _ = defineError(ErrDFUInProgress, http.StatusServiceUnavailable)

// ErrAuth (golint)
const ErrAuth = "{auth}"

var _ = defineError(ErrAuth, http.StatusUnauthorized)

// ErrTicket (golint)
const ErrTicket = "{ticket}"

var _ = defineError(ErrTicket, http.StatusUnauthorized)

// ErrHubNoHandler (golint)
const ErrHubNoHandler = "{no-handler}"

var _ = defineError(ErrHubNoHandler, http.StatusInternalServerError)

// ErrDeviceNotFound (golint)
const ErrDeviceNotFound = "{device-noexist}"

var _ = defineError(ErrDeviceNotFound, http.StatusNotFound)

// ErrDeviceNotSpecified (golint)
const ErrDeviceNotSpecified = "{device-none}"

var _ = defineError(ErrDeviceNotSpecified, http.StatusBadRequest)

// ErrDeviceId (golint)
const ErrDeviceId = "{device-id-invalid}"

var _ = defineError(ErrDeviceId, http.StatusBadRequest)

// ErrDeviceDisabled (golint)
const ErrDeviceDisabled = "{device-disabled}"

var _ = defineError(ErrDeviceDisabled, http.StatusBadRequest)

// ErrProductNotFound (golint)
const ErrProductNotFound = "{product-noexist}"

var _ = defineError(ErrProductNotFound, http.StatusNotFound)

// ErrProductNotSpecified (golint)
const ErrProductNotSpecified = "{product-none}"

var _ = defineError(ErrProductNotSpecified, http.StatusBadRequest)

// ErrAppNotFound (golint)
const ErrAppNotFound = "{app-noexist}"

var _ = defineError(ErrAppNotFound, http.StatusNotFound)

// ErrAppNotSpecified (golint)
const ErrAppNotSpecified = "{app-none}"

var _ = defineError(ErrAppNotSpecified, http.StatusBadRequest)

// ErrAppDeleted (golint)
const ErrAppDeleted = "{app-deleted}"

var _ = defineError(ErrAppDeleted, http.StatusGone)

// ErrAppExists (golint)
const ErrAppExists = "{app-exists}"

var _ = defineError(ErrAppExists, http.StatusConflict)

// ErrFleetNotFound (golint)
const ErrFleetNotFound = "{fleet-noexist}"

var _ = defineError(ErrFleetNotFound, http.StatusNotFound)

// ErrCardIo (golint)
const ErrCardIo = "{io}"

var _ = defineError(ErrCardIo, http.StatusBadGateway)

// ErrCardHeartbeat (golint) Doesn't seem to be used as a request error
const ErrCardHeartbeat = "{heartbeat}"

// ErrAccessDenied (golint)
const ErrAccessDenied = "{access-denied}"

var _ = defineError(ErrAccessDenied, http.StatusForbidden)

// ErrWebPayload (golint)
const ErrWebPayload = "{web-payload}"

var _ = defineError(ErrWebPayload, http.StatusBadRequest)

// ErrHubMode (golint)  Unused
const ErrHubMode = "{hub-mode}"

// ErrTemplateIncompatible (golint)
const ErrTemplateIncompatible = "{template-incompatible}"

var _ = defineError(ErrTemplateIncompatible, http.StatusBadRequest)

// ErrSyntax (golint)
const ErrSyntax = "{syntax}"

var _ = defineError(ErrSyntax, http.StatusBadRequest)

// ErrIncompatible (golint)
const ErrIncompatible = "{incompatible}"

var _ = defineError(ErrIncompatible, http.StatusNotAcceptable)

// ErrReqNotSupported (golint)
const ErrReqNotSupported = "{not-supported}"

var _ = defineError(ErrReqNotSupported, http.StatusNotImplemented)

// ErrTooBig (golint)
const ErrTooBig = "{too-big}"

var _ = defineError(ErrTooBig, http.StatusRequestEntityTooLarge)

// ErrJson (golint)
const ErrJson = "{not-json}"

var _ = defineError(ErrJson, http.StatusBadRequest)

// Status messages returned by the notecard in request.Status
const StatusIdle = "{idle}"
const StatusNtnIdle = "{ntn-idle}"
const StatusTransportConnected = "{connected}"
const StatusTransportDisconnected = "{disconnected}"
const StatusTransportConnecting = "{connecting}"
const StatusTransportConnectFailure = "{connect-failure}"
const StatusTransportConnectedClosed = "{connected-closed}"
const StatusTransportWaitService = "{wait-service}"
const StatusTransportWaitData = "{wait-data}"
const StatusTransportWaitGateway = "{wait-gateway}"
const StatusTransportWaitModule = "{wait-module}"
const StatusGPSInactive = "{gps-inactive}"

// These are returned from JSONata transforms as special strings to indicate the given behavior
// Used by Smart Fleets and during routing
const ErrAddToFleet = "{add-to-fleet}"
const ErrRemoveFromFleet = "{remove-from-fleet}"
const ErrLeaveFleetAlone = "{leave-fleet-alone}"
const ErrDoNotRoute = "{do-not-route}"

// These can be sent from Notehub to the notecard to indicate it should pause before reconnecting
// Currently unused
const ErrDeviceDelay5 = "{device-delay-5}"
const ErrDeviceDelay10 = "{device-delay-10}"
const ErrDeviceDelay15 = "{device-delay-15}"
const ErrDeviceDelay20 = "{device-delay-20}"
const ErrDeviceDelay30 = "{device-delay-30}"
const ErrDeviceDelay60 = "{device-delay-60}"

// ErrorContains tests to see if an error contains an error keyword that we might expect
func ErrorContains(err error, errKeyword string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(fmt.Sprintf("%s", err), errKeyword)
}

var errToHttpStatusMap map[string]int

func defineError(errKeyword string, httpStatus int) string {
	if errToHttpStatusMap == nil {
		errToHttpStatusMap = make(map[string]int)
	}
	errToHttpStatusMap[errKeyword] = httpStatus
	return errKeyword
}

// This scans a response.Err string for known error keywords and returns the appropriate HTTP status code
// If there are multiple error keywords, the first one found is used as the source for the code.
// We choose the first one because that should be the most relevant to the specific failure.
// If no known error keywords are found, we return HTTP 500 Internal Server Error.
func ErrorHttpStatus(errstr string) int {
	if errstr == "" {
		return http.StatusOK
	}
	start := strings.Index(errstr, "{")
	end := strings.Index(errstr, "}")
	if start == -1 || end < start {
		// Error message without a keyword. Assume it's an internal server error
		return http.StatusInternalServerError
	}
	errKeyword := errstr[start : end+1]
	if status, present := errToHttpStatusMap[errKeyword]; present {
		return status
	}
	return http.StatusInternalServerError
}

// ErrorClean removes all error keywords from an error string
func ErrorClean(err error) error {
	errstr := fmt.Sprintf("%s", err)
	for {
		left := strings.SplitN(errstr, "{", 2)
		if len(left) == 1 {
			break
		}
		errstr = left[0]
		b := strings.SplitN(left[1], "}", 2)
		if len(b) > 1 {
			errstr += strings.TrimPrefix(b[1], " ")
		}
	}
	return fmt.Errorf("%s", errstr)
}

// ErrorString safely returns a string from any error, returning "" for nil
func ErrorString(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%s", err)
}

// ErrorJSON returns a JSON object with nothing but an error code, and with an optional message
func ErrorJSON(message string, err error) (rspJSON []byte) {
	if message == "" {
		rspJSON = []byte(fmt.Sprintf("{\"err\":\"%q\"}", err))
	} else if err == nil {
		rspJSON = []byte(fmt.Sprintf("{\"err\":\"%q\"}", message))
	} else {
		rspJSON = []byte(fmt.Sprintf("{\"err\":\"%q: %q\"}", message, err))
	}
	return
}
