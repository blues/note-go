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
}
