package api

// GetEventsResponse v1
//
// The response object for getting events.
type GetEventsResponse struct {
	Events  []EventResponse `json:"events"`
	Through string          `json:"through"`
	HasMore bool            `json:"has_more"`
}

// GetEventsByCursorResponse v1
//
// The response object for getting events by cursor.
type GetEventsByCursorResponse struct {
	Events     []EventResponse `json:"events"`
	NextCursor string          `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
}

// EventResponse v1
//
// The response object for a event.
type EventResponse struct {
	UID string `json:"uid"`

	DeviceUID string `json:"device_uid"`

	// Notefile name
	File string `json:"file"`

	// RFC3339 timestamps, in UTC.
	// When the device created the event
	Captured string `json:"captured"`
	// When the server received the event
	Received string `json:"received"`

	Body *map[string]interface{} `json:"body"`

	TowerLocation *Location `json:"tower_location"`
	GPSLocation   *Location `json:"gps_location"`
}
