package api

// GetDeviceSessionsResponse is the structure returned from a GetDeviceSessions call
type GetDeviceSessionsResponse struct {
	// Sessions
	//
	// The requested page of session logs for the device
	//
	// required: true
	Sessions []DeviceSession `json:"sessions"`

	// HasMore
	//
	// A boolean indicating whether there is at least one more
	// page of data available after this page
	//
	// required: true
	HasMore bool `json:"has_more"`
}

// DeviceSession is the structure describing a device session
type DeviceSession struct {
	// Timestamp
	Started       string    `json:"started"`
	Duration      uint      `json:"duration"`
	Notes         uint      `json:"notes"`
	Bytes         uint      `json:"bytes"`
	TowerLocation *Location `json:"tower_location"`
}
