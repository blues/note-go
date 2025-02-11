package api

import "github.com/blues/note-go/note"

// GetEventsResponse v1
//
// The response object for getting events.
type GetEventsResponse struct {
	Events  []note.Event `json:"events"`
	Through string       `json:"through,omitempty"`
	HasMore bool         `json:"has_more"`
}

// GetEventsResponseSelectedFields v1
//
// The response object for getting events with selected fields.
type GetEventsResponseSelectedFields struct {
	Events  []note.Event `json:"events"`
	Through string       `json:"through,omitempty"`
	HasMore bool         `json:"has_more"`
}

// GetEventsByCursorResponse v1
//
// The response object for getting events by cursor.
type GetEventsByCursorResponse struct {
	Events     []note.Event `json:"events"`
	NextCursor string       `json:"next_cursor"`
	HasMore    bool         `json:"has_more"`
}
