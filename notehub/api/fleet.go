package api

// GetFleetsResponse v1
//
// The response object for getting fleets.
type GetFleetsResponse struct {
	Fleets []FleetResponse `json:"fleets"`
}

// FleetResponse v1
//
// The response object for a fleet.
type FleetResponse struct {
	UID   string `json:"uid"`
	Label string `json:"label"`
	// RFC3339 timestamp, in UTC.
	Created string `json:"created"`

	EnvironmentVariables map[string]string `json:"environment_variables"`

	SmartRule             string                     `json:"smart_rule,omitempty"`
	WatchdogMins          int64                      `json:"watchdog_mins,omitempty"`
	ConnectivityAssurance FleetConnectivityAssurance `json:"connectivity_assurance,omitempty"`
}

// PutDeviceFleetsRequest v1
//
// The request object for adding a device to fleets
type PutDeviceFleetsRequest struct {
	// FleetUIDs
	//
	// The fleets the device belong to
	//
	// required: true
	FleetUIDs []string `json:"fleet_uids"`
}

// DeleteDeviceFleetsRequest v1
//
// The request object for removing a device from fleets
type DeleteDeviceFleetsRequest struct {
	// FleetUIDs
	//
	// The fleets the device should be disassociated from
	//
	// required: true
	FleetUIDs []string `json:"fleet_uids"`
}

// PostFleetRequest v1
//
// The request object for adding a fleet for a project
type PostFleetRequest struct {
	Label string `json:"label"`

	SmartRule    string `json:"smart_rule,omitempty"`
	WatchdogMins int64  `json:"watchdog_mins,omitempty"`
}

// PutFleetRequest v1
//
// The request object for updating a fleet within a project
type PutFleetRequest struct {
	Label         string   `json:"label"`
	AddDevices    []string `json:"addDevices,omitempty"`
	RemoveDevices []string `json:"removeDevices,omitempty"`

	SmartRule    string `json:"smart_rule,omitempty"`
	WatchdogMins int64  `json:"watchdog_mins,omitempty"`
}

// FleetConnectivityAssurance v1
//
// Includes, Enabled = Whether Connectivity Assurance is enabled for this fleet
// With flexibility to add more information in the future
type FleetConnectivityAssurance struct {
	Enabled bool `json:"enabled"`
}
