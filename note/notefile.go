// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

// Notefile is The outermost data structure of a Notefile JSON
// object, containing a set of notes that may be synchronized.
type Notefile struct {
    modCount int                // Incremented every modification, for checkpointing purposes
    eventFn EventFunc			// The function to call when eventing of a change
    eventCtx interface{}		// An argument to be passed to the event call
	eventDeviceUID string		// The deviceUID being dealt with at the time of the event setup
	eventDeviceSN string		// The deviceSN being dealt with at the time of the event setup
	eventProductUID string		// The productUID being dealt with at the time of the event setup
	eventAppUID string			// The appUID being dealt with at the time of the event setup
    notefileID string           // The NotefileID for the open notefile
    notefileInfo NotefileInfo   // The NotefileInfo for the open notefile
    Queue bool                  `json:"Q,omitempty"`
    Notes map[string]Note       `json:"N,omitempty"`
    Trackers map[string]Tracker `json:"T,omitempty"`
    Change int64                `json:"C,omitempty"`
}

// Tracker is the structure maintained on a per-endpoint basis
type Tracker struct {
    Change int64                `json:"c,omitempty"`
    SessionID int64             `json:"i,omitempty"`
}

// SyncPriorityLowest (golint)
const SyncPriorityLowest =          -3
// SyncPriorityLower (golint)
const SyncPriorityLower =           -2
// SyncPriorityLow (golint)
const SyncPriorityLow =             -1
// SyncPriorityNormal (golint)
const SyncPriorityNormal =          0
// SyncPriorityHigh (golint)
const SyncPriorityHigh =            1
// SyncPriorityHigher (golint)
const SyncPriorityHigher =          2
// SyncPriorityHighest (golint)
const SyncPriorityHighest =         3

// NotefileInfo has parameters about the Notefile
type NotefileInfo struct {
	// The count of modified notes in this notefile. This is used in the Req API, but not in the Notebox info
	Changes int					`json:"changes,omitempty"`
    // This is a unidirectional "to-hub" or "from-hub" endpoint
    SyncHubEndpointID string    `json:"sync_hub_endpoint,omitempty"`
    // Relative positive/negative priority of data, with 0 being normal
    SyncPriority int            `json:"sync_priority,omitempty"`
    // True if this notefile should immediately sync upon any change
    SyncOnChange bool			`json:"sync_on_change,omitempty"`
    // Timed: Target for sync period, if modified and if the value hasn't been synced sooner
    SyncPeriodSecs int          `json:"sync_secs,omitempty"`
	// ReqTime is specified if notes stored in this notefile must have a valid time associated with them
    ReqTime bool				`json:"req_time,omitempty"`
	// ReqLoc is specified if notes stored in this notefile must have a valid location associated with them
    ReqLoc bool					`json:"req_loc,omitempty"`
}
