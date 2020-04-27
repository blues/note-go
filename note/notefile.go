// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

// SyncPriorityLowest (golint)
const SyncPriorityLowest = -3

// SyncPriorityLower (golint)
const SyncPriorityLower = -2

// SyncPriorityLow (golint)
const SyncPriorityLow = -1

// SyncPriorityNormal (golint)
const SyncPriorityNormal = 0

// SyncPriorityHigh (golint)
const SyncPriorityHigh = 1

// SyncPriorityHigher (golint)
const SyncPriorityHigher = 2

// SyncPriorityHighest (golint)
const SyncPriorityHighest = 3

// NotefileInfo has parameters about the Notefile
type NotefileInfo struct {
	// The count of modified notes in this notefile. This is used in the Req API, but not in the Notebox info
	Changes int `json:"changes,omitempty"`
	// This is a unidirectional "to-hub" or "from-hub" endpoint
	SyncHubEndpointID string `json:"sync_hub_endpoint,omitempty"`
	// Relative positive/negative priority of data, with 0 being normal
	SyncPriority int `json:"sync_priority,omitempty"`
	// Timed: Target for sync period, if modified and if the value hasn't been synced sooner
	SyncPeriodSecs int `json:"sync_secs,omitempty"`
	// ReqTime is specified if notes stored in this notefile must have a valid time associated with them
	ReqTime bool `json:"req_time,omitempty"`
	// ReqLoc is specified if notes stored in this notefile must have a valid location associated with them
	ReqLoc bool `json:"req_loc,omitempty"`
	// AnonAddAllowed is specified if anyone is allowed to drop into this notefile without authentication
	AnonAddAllowed bool `json:"anon_add,omitempty"`
}
