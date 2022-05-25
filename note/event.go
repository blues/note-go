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

// EventSession (golint)
const EventSession = "session.begin"

// EventGeolocation (golint)
const EventGeolocation = "device.geolocation"

// Event is the request structure passed to the Notification proc
//
// NOTE: This structure's underlying storage has been decoupled from the use of
// the structure in business logic.  As such, please share any changes to these
// structures with cloud services to ensure that storage and testing frameworks
// are kept in sync with these structures used for business logic
type Event struct {
	EventUID   string                  `json:"event,omitempty"`
	SessionUID string                  `json:"session,omitempty"`
	TLS        bool                    `json:"tls,omitempty"`
	BestID     string                  `json:"best_id,omitempty"`
	DeviceUID  string                  `json:"device,omitempty"`
	DeviceSN   string                  `json:"sn,omitempty"`
	ProductUID string                  `json:"product,omitempty"`
	EndpointID string                  `json:"endpoint,omitempty"`
	Received   float64                 `json:"received,omitempty"`
	Routed     int64                   `json:"routed,omitempty"`
	Req        string                  `json:"req,omitempty"`
	Rsp        string                  `json:"rsp,omitempty"`
	Error      string                  `json:"err,omitempty"`
	When       int64                   `json:"when,omitempty"`
	NotefileID string                  `json:"file,omitempty"`
	NoteID     string                  `json:"note,omitempty"`
	Updates    int32                   `json:"updates,omitempty"`
	Deleted    bool                    `json:"deleted,omitempty"`
	Sent       bool                    `json:"queued,omitempty"`
	Bulk       bool                    `json:"bulk,omitempty"`
	Body       *map[string]interface{} `json:"body,omitempty"`
	Payload    []byte                  `json:"payload,omitempty"`
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
	// Motion
	Moved       int64  `json:"moved,omitempty"`
	Orientation string `json:"orientation,omitempty"`
	// Signal strength/quality of the cell
	Rssi int    `json:"rssi,omitempty"`
	Sinr int    `json:"sinr,omitempty"`
	Rsrp int    `json:"rsrp,omitempty"`
	Rsrq int    `json:"rsrq,omitempty"`
	Rat  string `json:"rat,omitempty"`
	Bars uint32 `json:"bars,omitempty"`
	// Physical device info
	Voltage float64 `json:"voltage,omitempty"`
	Temp    float64 `json:"temp,omitempty"`
	// Triangulation
	Triangulate *map[string]interface{} `json:"triangulate,omitempty"`
	// "Routed" environment variables beginning with a "$" prefix
	Env *map[string]string `json:"environment,omitempty"`
	// App and device meta info
	App           *EventApp     `json:"project,omitempty"`
	DeviceContact *EventContact `json:"device_contact,omitempty"`
	// Where to reply
	ReplyURL string `json:"reply,omitempty"`
	// Event log
	LogAttn bool                     `json:"logattn,omitempty"`
	Log     map[string]EventLogEntry `json:"log,omitempty"`
	// True if the event was not routed due to exceeding the project's event limit.
	LogEventLimit bool `json:"log_event_limit,omitempty"`
}

// EventLogEntry is the log entry used by notification processing
//
// NOTE: This structure's underlying storage has been decoupled from the use of
// the structure in business logic.  As such, please share any changes to these
// structures with cloud services to ensure that storage and testing frameworks
// are kept in sync with these structures used for business logic
type EventLogEntry struct {
	Attn   bool   `json:"attn,omitempty"`
	Status string `json:"status,omitempty"`
	Text   string `json:"text,omitempty"`
	URL    string `json:"url,omitempty"`
}

// EventContact has the basic contact info structure
//
// NOTE: This structure's underlying storage has been decoupled from the use of
// the structure in business logic.  As such, please share any changes to these
// structures with cloud services to ensure that storage and testing frameworks
// are kept in sync with these structures used for business logic
type EventContact struct {
	Name        string `json:"name,omitempty"`
	Affiliation string `json:"org,omitempty"`
	Role        string `json:"role,omitempty"`
	Email       string `json:"email,omitempty"`
}

// EventContacts has contact info for this app
//
// NOTE: This structure's underlying storage has been decoupled from the use of
// the structure in business logic.  As such, please share any changes to these
// structures with cloud services to ensure that storage and testing frameworks
// are kept in sync with these structures used for business logic
type EventContacts struct {
	Admin *EventContact `json:"admin,omitempty"`
	Tech  *EventContact `json:"tech,omitempty"`
}

// EventApp has information about the app
//
// NOTE: This structure's underlying storage has been decoupled from the use of
// the structure in business logic.  As such, please share any changes to these
// structures with cloud services to ensure that storage and testing frameworks
// are kept in sync with these structures used for business logic
type EventApp struct {
	AppUID   string        `json:"id,omitempty"`
	AppLabel string        `json:"name,omitempty"`
	Contacts EventContacts `json:"contacts,omitempty"`
}
