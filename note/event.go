// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

// EventAdd (golint)
const EventAdd = "note.add"

// EventUpdate (golint)
const EventUpdate = "note.update"

// EventDelete (golint)
const EventDelete = "note.delete"

// EventPost (golint)
const EventPost = "post"

// EventPut (golint)
const EventPut = "put"

// EventGet (golint)
const EventGet = "get"

// EventNoAction (golint)
const EventNoAction = ""

// Event is the request structure passed to the Notification proc
type Event struct {
	Body          *interface{}             `json:"body,omitempty"`
	Req           string                   `json:"req,omitempty"`
	Rsp           string                   `json:"rsp,omitempty"`
	Error         string                   `json:"err,omitempty"`
	When          int64                    `json:"when,omitempty"`
	Where         string                   `json:"where,omitempty"`
	WhereLat      float64                  `json:"where_lat,omitempty"`
	WhereLon      float64                  `json:"where_lon,omitempty"`
	WhereLocation string                   `json:"where_location,omitempty"`
	WhereCountry  string                   `json:"where_country,omitempty"`
	WhereTimeZone string                   `json:"where_timezone,omitempty"`
	Routed        int64                    `json:"routed,omitempty"`
	NoteID        string                   `json:"note,omitempty"`
	NotefileID    string                   `json:"file,omitempty"`
	Updates       int32                    `json:"updates,omitempty"`
	Deleted       bool                     `json:"deleted,omitempty"`
	Sent          bool                     `json:"queued,omitempty"`
	Bulk          bool                     `json:"bulk,omitempty"`
	TowerCountry  string                   `json:"tower_country,omitempty"`
	TowerLocation string                   `json:"tower_location,omitempty"`
	TowerTimeZone string                   `json:"tower_timezone,omitempty"`
	TowerLat      float64                  `json:"tower_lat,omitempty"`
	TowerLon      float64                  `json:"tower_lon,omitempty"`
	LogAttn       bool                     `json:"logattn,omitempty"`
	Log           map[string]EventLogEntry `json:"log,omitempty"`
	App           *EventApp                `json:"project,omitempty"`
	DeviceContact *EventContact            `json:"device_contact,omitempty"`
	EndpointID    string                   `json:"endpoint,omitempty"`
	DeviceSN      string                   `json:"sn,omitempty"`
	DeviceUID     string                   `json:"device,omitempty"`
	ProductUID    string                   `json:"product,omitempty"`
	SessionUID    string                   `json:"session,omitempty"`
	EventUID      string                   `json:"event,omitempty"`
	Payload       []byte                   `json:"payload,omitempty"`
}

// EventLogEntry is the log entry used by notification processing
type EventLogEntry struct {
	Attn   bool   `json:"attn,omitempty"`
	Status string `json:"status,omitempty"`
	Text   string `json:"text,omitempty"`
}

// EventContact has the basic contact info structure
type EventContact struct {
	Name        string `json:"name,omitempty"`
	Affiliation string `json:"org,omitempty"`
	Role        string `json:"role,omitempty"`
	Email       string `json:"email,omitempty"`
}

// EventContacts has contact info for this app
type EventContacts struct {
	Admin *EventContact `json:"admin,omitempty"`
	Tech  *EventContact `json:"tech,omitempty"`
}

// EventApp has information about the app
type EventApp struct {
	AppUID   string        `json:"id,omitempty"`
	AppLabel string        `json:"name,omitempty"`
	Contacts EventContacts `json:"contacts,omitempty"`
}
