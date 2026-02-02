package api

import (
	"strings"

	"github.com/blues/note-go/note"
)

// GetDevicesResponse v1
//
// The response object for getting devices.
type GetDevicesResponse struct {
	Devices []GetDeviceResponse `json:"devices"`
	HasMore bool                `json:"has_more"`
}

// Part of the response object for a device.
type DeviceHealthLogEntry struct {
	When  string `json:"when"`
	Text  string `json:"text"`
	Alert bool   `json:"alert"`
}

// DeviceResponse v1
//
// The response object for a device.
type GetDeviceResponse struct {
	UID          string `json:"uid"`
	SerialNumber string `json:"serial_number,omitempty"`
	SKU          string `json:"sku,omitempty"`

	// RFC3339 timestamps, in UTC.
	Provisioned  string  `json:"provisioned"`
	LastActivity *string `json:"last_activity"`

	FirmwareHost     string `json:"firmware_host,omitempty"`
	FirmwareNotecard string `json:"firmware_notecard,omitempty"`

	Contact *ContactResponse `json:"contact,omitempty"`

	ProductUID string   `json:"product_uid"`
	FleetUIDs  []string `json:"fleet_uids"`

	TowerInfo            *TowerInformation `json:"tower_info,omitempty"`
	TowerLocation        *Location         `json:"tower_location,omitempty"`
	GPSLocation          *Location         `json:"gps_location,omitempty"`
	TriangulatedLocation *Location         `json:"triangulated_location,omitempty"`
	BestLocation         *Location         `json:"best_location,omitempty"`

	Voltage     float64      `json:"voltage"`
	Temperature float64      `json:"temperature"`
	DFUEnv      *note.DFUEnv `json:"dfu,omitempty"`
	Disabled    bool         `json:"disabled,omitempty"`
	Tags        string       `json:"tags,omitempty"`

	// Activity
	RecentActivityBase   string `json:"recent_event_base,omitempty"`
	RecentEventCount     []int  `json:"recent_event_count,omitempty"`
	RecentSessionCount   []int  `json:"recent_session_count,omitempty"`
	RecentSessionSeconds []int  `json:"recent_session_seconds,omitempty"`

	// Health
	HealthLog []DeviceHealthLogEntry `json:"health_log,omitempty"`
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
	ProductUID string    `json:"product_uid"`
	DeviceSN   string    `json:"device_sn"`
	FleetUIDs  *[]string `json:"fleet_uids,omitempty"`
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

var allDfuPhases = []note.DfuPhase{
	note.DfuPhaseUnknown,
	note.DfuPhaseIdle,
	note.DfuPhaseError,
	note.DfuPhaseDownloading,
	note.DfuPhaseSideloading,
	note.DfuPhaseReady,
	note.DfuPhaseReadyRetry,
	note.DfuPhaseUpdating,
	note.DfuPhaseCompleted,
}

func ParseDfuPhase(phase string) note.DfuPhase {
	phase = strings.ToLower(phase)
	for _, validPhase := range allDfuPhases {
		if phase == string(validPhase) {
			return validPhase
		}
	}
	return note.DfuPhaseUnknown
}

func IsDfuTerminal(phase note.DfuPhase) bool {
	return phase == note.DfuPhaseError ||
		phase == note.DfuPhaseCompleted ||
		phase == note.DfuPhaseIdle
}
