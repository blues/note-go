// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

// TrackNotefile is the hard-wired notefile that the notecard can use for tracking the device
const TrackNotefile = "_track.qo"

// NotecardRequestNotefile is a special notefile for sending notecard requests
const NotecardRequestNotefile = "_req.qis"

// NotecardResponseNotefile is a special notefile for sending notecard responses
const NotecardResponseNotefile = "_rsp.qos"

// LogNotefile is the hard-wired notefile that the notecard uses for debug logging
const LogNotefile = "_log.qo"

// EnvNotefile is the hard-wired notefile that the notecard uses for env vars
const EnvNotefile = "_env.dbs"

// SessionNotefile is the hard-wired notefile that the notehub uses when starting a session
const SessionNotefile = "_session.qo"

// HealthNotefile is the hard-wired notefile that the notecard uses for health-related info
const HealthNotefile = "_health.qo"

// HealthHostNotefile is the hard-wired notefile that the host uses for health-related info
const HealthHostNotefile = "_health_host.qo"

// GeolocationNotefile is the hard-wired notefile that the notehub uses when performing a geolocation
const GeolocationNotefile = "_geolocate.qo"

// TowerNotefile is the hard-wired notefile that the notehub uses when performing tower updates
const TowerNotefile = "_tower.qo"

// SocketNotefile is the hard-wired notefile that the notehub uses when doing websocket I/O
const SocketNotefile = "_socket.qo"

// WebNotefile is the hard-wired notefile that the notehub uses when performing web requests
const WebNotefile = "_web.qo"

// WatchdogNotefile is the hard-wired notefile that the notehub uses when adding watchdog messages
const WatchdogNotefile = "_watchdog.qo"

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
	// The count of total notes in this notefile. This is used in the Req API, but not in the Notebox info
	Total int `json:"total,omitempty"`
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
	// ImportTime is the epoch time of when an external data source (such as a feed) last sync'ed data inbound
	ImportTime int64 `json:"import_time,omitempty"`
	// ExportTime is the epoch time of when an external data source (such as a feed) last sync'ed data outbound
	ExportTime int64 `json:"export_time,omitempty"`
}

// Information about notefiles and their templates
type NotefileDesc struct {
	NotefileID      string       `json:"file,omitempty"`
	Info            NotefileInfo `json:"info,omitempty"`
	BodyTemplate    string       `json:"body_template,omitempty"`
	PayloadTemplate uint32       `json:"payload_template,omitempty"`
	TemplateFormat  uint32       `json:"template_format,omitempty"`
	TemplatePort    uint16       `json:"template_port,omitempty"`
}
