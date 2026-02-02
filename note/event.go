// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

import (
	"time"
)

// EventAdd (golint)
const EventAdd = "note.add"

// EventUpdate (golint)
const EventUpdate = "note.update"

// EventDelete (golint)
const EventDelete = "note.delete"

// EventTest (golint)
const EventTest = "test"

// EventPost (golint)
const EventPost = "post"

// EventPut (golint)
const EventPut = "put"

// EventGet (golint)
const EventGet = "get"

// EventNoAction (golint)
const EventNoAction = ""

// EventSessionBegin (golint)
const EventSessionBegin = "session.begin"

// EventSessionEndNotehub (golint)
const EventSessionEnd = "session.end"

// EventGeolocation (golint)
const EventGeolocation = "device.geolocation"

// EventTower (golint)
const EventTower = "device.tower"

// EventSocket (golint)
const EventSocket = "web.socket"

// EventWebhook (golint)
const EventWebhook = "webhook"

// Event is the request structure passed to the Notification proc
//
// NOTE: This structure's underlying storage has been decoupled from the use of
// the structure in business logic.  As such, please share any changes to these
// structures with cloud services to ensure that storage and testing frameworks
// are kept in sync with these structures used for business logic
type Event struct {
	EventUID string `json:"event,omitempty"`
	// Indicates whether or not this event is a "platform event" - that is, an event generated automatically
	// somewhere in the notecard or notehub largely for administrative purposes that doesn't pertain to either
	// implicit or explicit user data.
	Platform bool `json:"platform,omitempty"`
	// These fields, and only these fields, are regarded as "user data".  All
	// the rest of the fields are regarded as "metadata".
	When       int64                   `json:"when,omitempty"`
	NotefileID string                  `json:"file,omitempty"`
	NoteID     string                  `json:"note,omitempty"`
	Body       *map[string]interface{} `json:"body,omitempty"`
	Payload    []byte                  `json:"payload,omitempty"`
	Details    *map[string]interface{} `json:"details,omitempty"`
	// Metadata
	SessionUID       string  `json:"session,omitempty"`
	SessionBegan     int64   `json:"session_began,omitempty"`
	TLS              bool    `json:"tls,omitempty"`
	Transport        string  `json:"transport,omitempty"`
	Continuous       bool    `json:"continuous,omitempty"`
	BestID           string  `json:"best_id,omitempty"`
	DeviceUID        string  `json:"device,omitempty"`
	DeviceSN         string  `json:"sn,omitempty"`
	ProductUID       string  `json:"product,omitempty"`
	AppUID           string  `json:"app,omitempty"`
	Received         float64 `json:"received,omitempty"`
	Req              string  `json:"req,omitempty"`
	Error            string  `json:"err,omitempty"`
	Updates          int32   `json:"updates,omitempty"`
	Deleted          bool    `json:"deleted,omitempty"`
	Sent             bool    `json:"queued,omitempty"`
	Bulk             bool    `json:"bulk,omitempty"`
	BulkReceived     float64 `json:"batch_received,omitempty"`
	BulkNumber       uint32  `json:"batch_number,omitempty"`
	BulkTotal        uint32  `json:"batch_total,omitempty"`
	FirmwareHost     string  `json:"firmware_host,omitempty"`
	FirmwareNotecard string  `json:"firmware_notecard,omitempty"`
	// This field is ONLY used when we remove the payload for storage reasons, to show the app how large it was
	MissingPayloadLength int64 `json:"payload_length,omitempty"`
	// Location
	BestLocationType string  `json:"best_location_type,omitempty"`
	BestLocationWhen int64   `json:"best_location_when,omitempty"`
	BestLat          float64 `json:"best_lat,omitempty"`
	BestLon          float64 `json:"best_lon,omitempty"`
	BestLocation     string  `json:"best_location,omitempty"`
	BestCountry      string  `json:"best_country,omitempty"`
	BestTimeZone     string  `json:"best_timezone,omitempty"`
	Where            string  `json:"where_olc,omitempty"`
	WhereWhen        int64   `json:"where_when,omitempty"`
	WhereLat         float64 `json:"where_lat,omitempty"`
	WhereLon         float64 `json:"where_lon,omitempty"`
	WhereLocation    string  `json:"where_location,omitempty"`
	WhereCountry     string  `json:"where_country,omitempty"`
	WhereTimeZone    string  `json:"where_timezone,omitempty"`
	TowerWhen        int64   `json:"tower_when,omitempty"`
	TowerLat         float64 `json:"tower_lat,omitempty"`
	TowerLon         float64 `json:"tower_lon,omitempty"`
	TowerCountry     string  `json:"tower_country,omitempty"`
	TowerLocation    string  `json:"tower_location,omitempty"`
	TowerTimeZone    string  `json:"tower_timezone,omitempty"`
	TowerID          string  `json:"tower_id,omitempty"`
	TriWhen          int64   `json:"tri_when,omitempty"`
	TriLat           float64 `json:"tri_lat,omitempty"`
	TriLon           float64 `json:"tri_lon,omitempty"`
	TriLocation      string  `json:"tri_location,omitempty"`
	TriCountry       string  `json:"tri_country,omitempty"`
	TriTimeZone      string  `json:"tri_timezone,omitempty"`
	TriPoints        int32   `json:"tri_points,omitempty"`

	// Triangulation
	Triangulate *map[string]interface{} `json:"triangulate,omitempty"`
	// "Routed" environment variables beginning with a "$" prefix
	Env       *map[string]string `json:"environment,omitempty"`
	Status    EventRoutingStatus `json:"status,omitempty"`
	FleetUIDs *[]string          `json:"fleets,omitempty"`

	// ONLY POPULATED FOR EventSessionBegin with info both from notecard and notehub
	DeviceSKU          string  `json:"sku,omitempty"`
	DeviceOrderingCode string  `json:"ordering_code,omitempty"`
	DeviceFirmware     int64   `json:"firmware,omitempty"`
	Bearer             string  `json:"bearer,omitempty"`
	CellID             string  `json:"cellid,omitempty"`
	Bssid              string  `json:"bssid,omitempty"`
	Ssid               string  `json:"ssid,omitempty"`
	Iccid              string  `json:"iccid,omitempty"`
	Apn                string  `json:"apn,omitempty"`
	Rssi               int     `json:"rssi,omitempty"`
	Sinr               int     `json:"sinr,omitempty"`
	Rsrp               int     `json:"rsrp,omitempty"`
	Rsrq               int     `json:"rsrq,omitempty"`
	Rat                string  `json:"rat,omitempty"`
	Bars               uint32  `json:"bars,omitempty"`
	Voltage            float64 `json:"voltage,omitempty"`
	Temp               float64 `json:"temp,omitempty"`
	Moved              int64   `json:"moved,omitempty"`
	Orientation        string  `json:"orientation,omitempty"`
	PowerCharging      bool    `json:"power_charging,omitempty"`
	PowerUsb           bool    `json:"power_usb,omitempty"`
	PowerPrimary       bool    `json:"power_primary,omitempty"`
	PowerMahUsed       float64 `json:"power_mah,omitempty"`

	// ONLY POPULATED FOR EventSessionEnd because it comes from the notehub
	NotehubLastWorkDone int64  `json:"hub_last_work_done,omitempty"`
	NotehubDurationSecs int64  `json:"hub_duration_secs,omitempty"`
	NotehubEventCount   int64  `json:"hub_events_routed,omitempty"`
	NotehubRcvdBytes    uint32 `json:"hub_rcvd_bytes,omitempty"`
	NotehubSentBytes    uint32 `json:"hub_sent_bytes,omitempty"`
	NotehubTCPSessions  uint32 `json:"hub_tcp_sessions,omitempty"`
	NotehubTLSSessions  uint32 `json:"hub_tls_sessions,omitempty"`
	NotehubRcvdNotes    uint32 `json:"hub_rcvd_notes,omitempty"`
	NotehubSentNotes    uint32 `json:"hub_sent_notes,omitempty"`

	// ONLY POPULATED for EventSessionEndNotecard because it comes from the notecard
	NotecardRcvdBytes          uint32 `json:"card_rcvd_bytes,omitempty"`
	NotecardSentBytes          uint32 `json:"card_sent_bytes,omitempty"`
	NotecardRcvdBytesSecondary uint32 `json:"card_rcvd_bytes_secondary,omitempty"`
	NotecardSentBytesSecondary uint32 `json:"card_sent_bytes_secondary,omitempty"`
	NotecardTCPSessions        uint32 `json:"card_tcp_sessions,omitempty"`
	NotecardTLSSessions        uint32 `json:"card_tls_sessions,omitempty"`
	NotecardRcvdNotes          uint32 `json:"card_rcvd_notes,omitempty"`
	NotecardSentNotes          uint32 `json:"card_sent_notes,omitempty"`
}

type EventRoutingStatus string

const (
	EventStatusEmpty      EventRoutingStatus = ""
	EventStatusSuccess    EventRoutingStatus = "success"
	EventStatusFailure    EventRoutingStatus = "failure"
	EventStatusInProgress EventRoutingStatus = "in_progress"
)

// RouteLogEntry is the log entry used by notification processing
type RouteLogEntry struct {
	EventSerial int64         `json:"event,omitempty"`
	RouteSerial int64         `json:"route,omitempty"`
	Date        time.Time     `json:"date,omitempty"`
	Attn        bool          `json:"attn,omitempty"`
	Status      string        `json:"status,omitempty"`
	Text        string        `json:"text,omitempty"`
	URL         string        `json:"url,omitempty"`
	Source      RoutingSource `json:"source,omitempty"`

	// Time in milliseconds that the route took to process
	// We're making a simplifying assumption that the route will always
	// take at least 1ms. So 0 means we didn't record the duration.
	Duration int64 `json:"duration,omitempty"`
}

type RoutingSource uint8

const (
	RoutingSourceUnknown RoutingSource = iota
	RoutingSourceNormal
	RoutingSourceProxy
	RoutingSourceRetry
	RoutingSourceManual
	RoutingSourceDirect
	RoutingSourceTest
)

// String returns a string representation of the routing source
func (s RoutingSource) String() string {
	switch s {
	case RoutingSourceUnknown:
		return "" // display nothing if no entry/default
	case RoutingSourceNormal:
		return "Normal Routing"
	case RoutingSourceProxy:
		return "Web Proxy Request"
	case RoutingSourceRetry:
		return "Auto-Retry"
	case RoutingSourceManual:
		return "Manual Reroute"
	case RoutingSourceDirect:
		return "Direct Routing" //only used for test events, should never show in route logs
	case RoutingSourceTest:
		return "Test" // only used for tests
	default:
		return "invalid"
	}
}

// GetAggregateEventStatus returns the status of the event given all
// of the route logs for the event.
//
// The aggregate status is determined by taking the most recent status
// for each route.  If any of these are failures then the overall status
// is EventStatusFailure, otherwise it's EventStatusSuccess
func GetAggregateEventStatus(logs []RouteLogEntry) EventRoutingStatus {
	if len(logs) == 0 {
		return EventStatusEmpty
	}

	latest := make(map[int64]RouteLogEntry)
	for _, log := range logs {
		if val, ok := latest[log.RouteSerial]; !ok || log.Date.After(val.Date) {
			latest[log.RouteSerial] = log
		}
	}

	for _, latestLogEntry := range latest {
		if latestLogEntry.Attn {
			return EventStatusFailure
		}
	}

	return EventStatusSuccess
}
