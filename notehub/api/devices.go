package api

// GetDevicesResponse v1
//
// The response object for getting devices.
type GetDevicesResponse struct {
	Devices []DeviceResponse `json:"devices"`
	HasMore bool             `json:"has_more"`
}

// DeviceResponse v1
//
// The response object for a device.
type DeviceResponse struct {
	UID          string `json:"uid"`
	SerialNumber string `json:"serial_number"`

	// RFC3339 timestamps, in UTC.
	Provisioned  string  `json:"provisioned"`
	LastActivity *string `json:"last_activity"`

	Contact *ContactResponse `json:"contact"`

	ProductUID string   `json:"product_uid"`
	FleetUIDs  []string `json:"fleet_uids"`

	TowerInfo            *TowerInformation `json:"tower_info"`
	TowerLocation        *Location         `json:"tower_location"`
	GPSLocation          *Location         `json:"gps_location"`
	TriangulatedLocation *Location         `json:"triangulated_location"`

	Voltage     float64 `json:"voltage"`
	Temperature float64 `json:"temperature"`
}

// GetDevicesPublicKeysResponse v1
//
// The response object for retrieving a collection of devices' public keys
type GetDevicesPublicKeysResponse struct {
	DevicePublicKeys []DevicePublicKey `json:"device_public_keys"`
	HasMore          bool              `json:"has_more"`
}

// DevicePublicKey v1
//
// A structure representing the public key for a specific device
type DevicePublicKey struct {
	UID       string `json:"uid"`
	PublicKey string `json:"key"`
}

// ProvisionDeviceRequest v1
//
// The request object for provisioning a device
type ProvisionDeviceRequest struct {
	ProductUID string `json:"product_uid"`
}

// GetDeviceLatestResponse v1
//
// The response object for retrieving the latest notefile values for a device
type GetDeviceLatestResponse struct {
	LatestEvents []LatestEvent `json:"latest_events"`
}

// LatestEvent v1
//
// The response object of the returnable information from a "latest" event for
// a device
type LatestEvent struct {
	File     string                  `json:"file"`
	Captured string                  `json:"captured"`
	Received string                  `json:"received"`
	EventUID string                  `json:"event_uid"`
	Body     *map[string]interface{} `json:"body"`
}

// Location v1
//
// The response object for a location.
type Location struct {
	When      string  `json:"when"`
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Timezone  string  `json:"timezone"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// TowerInformation v1
//
// The response object for tower information.
type TowerInformation struct {
	// Mobile Country Code
	MCC int `json:"mcc"`
	// Mobile Network Code
	MNC int `json:"mnc"`
	// Location Area Code
	LAC    int `json:"lac"`
	CellID int `json:"cell_id"`
}

// GetDeviceHealthLogResponse v1
//
// The response object for getting a device's health log.
type GetDeviceHealthLogResponse struct {
	HealthLog []HealthLogEntry `json:"health_log"`
}

// HealthLogEntry v1
//
// The response object for a health log entry.
type HealthLogEntry struct {
	When  string `json:"when"`
	Alert bool   `json:"alert"`
	Text  string `json:"text"`
}
