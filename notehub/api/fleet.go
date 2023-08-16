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

	Is     string   `json:"is,omitempty"`
	IsLike []string `json:"is_like,omitempty"`
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

	Is     string   `json:"is,omitempty"`
	IsLike []string `json:"is_like,omitempty"`
}

// PutFleetRequest v1
//
// The request object for updating a fleet within a project
type PutFleetRequest struct {
	Label         string   `json:"label"`
	AddDevices    []string `json:"addDevices,omitempty"`
	RemoveDevices []string `json:"removeDevices,omitempty"`
	Is            string   `json:"is,omitempty"`
	IsLike        []string `json:"is_like,omitempty"`
}
