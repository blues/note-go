package api

import "github.com/blues/note-go/note"

// GetDeviceSessionsResponse is the structure returned from a GetDeviceSessions call
type GetDeviceSessionsResponse struct {
	// Sessions
	//
	// The requested page of session logs for the device
	//
	// required: true
	Sessions []note.DeviceSession `json:"sessions"`

	// HasMore
	//
	// A boolean indicating whether there is at least one more
	// page of data available after this page
	//
	// required: true
	HasMore bool `json:"has_more"`
}
