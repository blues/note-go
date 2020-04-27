// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

// Package note errors.go contains programmatically-testable error strings
package note

import (
	"fmt"
	"strings"
)

// ErrTimeout (golint)
const ErrTimeout = "{timeout}"

// ErrClosed (golint)
const ErrClosed = "{closed}"

// ErrFileNoExist (golint)
const ErrFileNoExist = "{file-noexist}"

// ErrNotefileName (golint)
const ErrNotefileName = "{notefile-bad-name}"

// ErrNotefileInUse (golint)
const ErrNotefileInUse = "{notefile-in-use}"

// ErrNotefileExists (golint)
const ErrNotefileExists = "{notefile-exists}"

// ErrNotefileNoExist (golint)
const ErrNotefileNoExist = "{notefile-noexist}"

// ErrNotefileQueueDisallowed (golint)
const ErrNotefileQueueDisallowed = "{notefile-queue-disallowed}"

// ErrNoteNoExist (golint)
const ErrNoteNoExist = "{note-noexist}"

// ErrNoteExists (golint)
const ErrNoteExists = "{note-exists}"

// ErrTrackerNoExist (golint)
const ErrTrackerNoExist = "{tracker-noexist}"

// ErrTrackerExists (golint)
const ErrTrackerExists = "{tracker-exists}"

// ErrTransportConnected (golint)
const ErrTransportConnected = "{connected}"

// ErrTransportDisconnected (golint)
const ErrTransportDisconnected = "{disconnected}"

// ErrTransportConnecting (golint)
const ErrTransportConnecting = "{connecting}"

// ErrTransportConnectFailure (golint)
const ErrTransportConnectFailure = "{connect-failure}"

// ErrTransportConnectedClosed (golint)
const ErrTransportConnectedClosed = "{connected-closed}"

// ErrTransportWaitService (golint)
const ErrTransportWaitService = "{wait-service}"

// ErrTransportWaitData (golint)
const ErrTransportWaitData = "{wait-data}"

// ErrTransportWaitGateway (golint)
const ErrTransportWaitGateway = "{wait-gateway}"

// ErrTransportWaitModule (golint)
const ErrTransportWaitModule = "{wait-module}"

// ErrAuth (golint)
const ErrAuth = "{auth}"

// ErrTicket (golint)
const ErrTicket = "{ticket}"

// ErrHubNoHandler (golint)
const ErrHubNoHandler = "{no-handler}"

// ErrIdle (golint)
const ErrIdle = "{idle}"

// ErrDeviceNotFound (golint)
const ErrDeviceNotFound = "{device-noexist}"

// ErrProductNotFound (golint)
const ErrProductNotFound = "{product-noexist}"

// ErrCardIo (golint)
const ErrCardIo = "{io}"

// ErrorContains tests to see if an error contains an error keyword that we might expect
func ErrorContains(err error, errKeyword string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(fmt.Sprintf("%s", err), errKeyword)
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
	return fmt.Errorf(errstr)
}

// ErrorString safely returns a string from any error, returning "" for nil
func ErrorString(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%s", err)
}
