package api

import "github.com/blues/note-go/note"

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
	SerialNumber string `json:"serial_number,omitempty"`
	SKU          string `json:"sku,omitempty"`

	// RFC3339 timestamps, in UTC.
	Provisioned  string  `json:"provisioned"`
	LastActivity *string `json:"last_activity"`

	Contact *ContactResponse `json:"contact,omitempty"`

	ProductUID string   `json:"product_uid"`
	FleetUIDs  []string `json:"fleet_uids"`

	TowerInfo            *TowerInformation `json:"tower_info,omitempty"`
	TowerLocation        *Location         `json:"tower_location,omitempty"`
	GPSLocation          *Location         `json:"gps_location,omitempty"`
	TriangulatedLocation *Location         `json:"triangulated_location,omitempty"`

	Voltage     float64 `json:"voltage"`
	Temperature float64 `json:"temperature"`
	DFUEnv      *DFUEnv `json:"dfu,omitempty"`
	Disabled    bool    `json:"disabled,omitempty"`
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
	DeviceSN   string `json:"device_sn"`
}

// GetDeviceLatestResponse v1
//
// The response object for retrieving the latest notefile values for a device
type GetDeviceLatestResponse struct {
	LatestEvents []note.Event `json:"latest_events"`
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

// DFUState is the state of the DFU in progress
type DFUState struct {
	Type              string `json:"type,omitempty"`
	File              string `json:"file,omitempty"`
	Length            uint32 `json:"length,omitempty"`
	CRC32             uint32 `json:"crc32,omitempty"`
	MD5               string `json:"md5,omitempty"`
	Phase             string `json:"mode,omitempty"`
	Status            string `json:"status,omitempty"`
	BeganSecs         uint32 `json:"began,omitempty"`
	RetryCount        uint32 `json:"retry,omitempty"`
	ConsecutiveErrors uint32 `json:"errors,omitempty"`
	ReadFromService   uint32 `json:"read,omitempty"`
	UpdatedSecs       uint32 `json:"updated,omitempty"`
	Version           string `json:"version,omitempty"`
}

// DFUEnv is the data structure passed to Notehub when DFU info changes
type DFUEnv struct {
	Card *DFUState `json:"card,omitempty"`
	User *DFUState `json:"user,omitempty"`
}
