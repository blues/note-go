// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

// EventAdd (golint)
const EventAdd =				"note.add"
// EventUpdate (golint)
const EventUpdate =				"note.update"
// EventDelete (golint)
const EventDelete =				"note.delete"
// EventPost (golint)
const EventPost =				"post"
// EventPut (golint)
const EventPut =				"put"
// EventGet (golint)
const EventGet =				"get"
// EventNoAction (golint)
const EventNoAction =			""

// Event is the request structure passed to the Notification proc
type Event struct {
    Req string                  `json:"req,omitempty"`
    Rsp string					`json:"rsp,omitempty"`
    Error string                `json:"err,omitempty"`
	NoteID string				`json:"note,omitempty"`
    Deleted bool                `json:"deleted,omitempty"`
    Sent bool					`json:"queued,omitempty"`
    Bulk bool                   `json:"bulk,omitempty"`
    NotefileID string           `json:"file,omitempty"`
    DeviceUID string            `json:"device,omitempty"`
	DeviceSN string				`json:"sn,omitempty"`
	AppUID string				`json:"app,omitempty"`
	ProductUID string			`json:"product,omitempty"`
	EndpointID string			`json:"endpoint,omitempty"`
	TowerCountry string			`json:"tower_country,omitempty"`
	TowerLocation string		`json:"tower_location,omitempty"`
	TowerTimeZone string		`json:"tower_timezone,omitempty"`
	TowerLat float64			`json:"tower_lat,omitempty"`
	TowerLon float64			`json:"tower_lon,omitempty"`
	When int64					`json:"when,omitempty"`
	Where string				`json:"where,omitempty"`
	WhereLat float64			`json:"where_lat,omitempty"`
	WhereLon float64			`json:"where_lon,omitempty"`
	WhereLocation string		`json:"where_location,omitempty"`
	WhereCountry string			`json:"where_country,omitempty"`
	WhereTimeZone string		`json:"where_timezone,omitempty"`
	Routed int64				`json:"routed,omitempty"`
    Updates int32               `json:"updates,omitempty"`
    Body *interface{}			`json:"body,omitempty"`
    Payload []byte              `json:"payload,omitempty"`
	SessionUID string			`json:"session,omitempty"`
	// Used only for event logs
	LogAttn bool				`json:"logattn,omitempty"`
	Log map[string]EventLogEntry `json:"log,omitempty"`
}

// EventLogEntry is the log entry used by notification processing
type EventLogEntry struct {
	Attn bool					`json:"attn,omitempty"`
    Status string				`json:"status,omitempty"`
    Text string					`json:"text,omitempty"`
}
