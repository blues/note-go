// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notehub

import (
	"fmt"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
)

// Supported requests

// HubDeviceContact (golint)
const HubDeviceContact = "hub.device.contact"

// HubAppGetSchemas (golint)
const HubAppGetSchemas = "hub.app.schemas.get"

// HubQuery (golint)
const HubQuery = "hub.app.data.query"

// HubEventQuery (golint)
const HubEventQuery = "hub.app.event.query"

// HubSessionQuery (golint)
const HubSessionQuery = "hub.app.session.query"

// HubAppUpload (golint)
const HubAppUpload = "hub.app.upload.add"

// HubUpload (golint)
const HubUpload = "hub.upload.add"

// HubAppUploads (golint)
const HubAppUploads = "hub.app.upload.query"

// HubUploads (golint)
const HubUploads = "hub.upload.query"

// HubAppUploadSet (golint)
const HubAppUploadSet = "hub.app.upload.set"

// HubUploadSet (golint)
const HubUploadSet = "hub.upload.set"

// HubAppUploadDelete (golint)
const HubAppUploadDelete = "hub.app.upload.delete"

// HubUploadDelete (golint)
const HubUploadDelete = "hub.upload.delete"

// HubAppUploadRead (golint)
const HubAppUploadRead = "hub.app.upload.get"

// HubUploadRead (golint)
const HubUploadRead = "hub.upload.get"

// HubAppSetTransform (golint)
const HubAppSetTransform = "hub.app.transform.set"

// HubAppGetTransform (golint)
const HubAppGetTransform = "hub.app.transform.get"

// HubEnvSet (golint)
const HubEnvSet = "hub.env.set"

// HubEnvGet (golint)
const HubEnvGet = "hub.env.get"

// HubEnvScopeApp (golint)
const HubEnvScopeApp = "app"

// HubEnvScopeProject (golint)
const HubEnvScopeProject = "project"

// HubEnvScopeFleet (golint)
const HubEnvScopeFleet = "fleet"

// HubEnvScopeFleets (golint)
const HubEnvScopeFleets = "fleets"

// HubEnvScopeDevice (golint)
const HubEnvScopeDevice = "device"

// HubCompressModeSnappy (golint)
const HubCompressModeSnappy = "snappy"

// HubCompressModeCobs (golint)
const HubCompressModeCobs = "cobs"

// HubRequest is is the core data structure for notehub-specific requests
type HubRequest struct {
	notecard.Request `json:",omitempty"`
	Contact          *note.Contact                 `json:"contact,omitempty"`
	AppUID           string                        `json:"app,omitempty"`
	FleetUID         string                        `json:"fleet,omitempty"`
	EventSerials     []string                      `json:"events,omitempty"`
	DbQuery          *DbQuery                      `json:"query,omitempty"`
	Uploads          []UploadMetadata              `json:"uploads,omitempty"`
	Contains         string                        `json:"contains,omitempty"`
	Handlers         *[]string                     `json:"handlers,omitempty"`
	FileType         UploadType                    `json:"type,omitempty"`
	FileTags         string                        `json:"tags,omitempty"`
	FileNotes        string                        `json:"filenotes,omitempty"`
	Provision        bool                          `json:"provision,omitempty"`
	Scope            string                        `json:"scope,omitempty"`
	Env              *map[string]string            `json:"env,omitempty"`
	FleetEnv         *map[string]map[string]string `json:"fleet_env,omitempty"`
	PIN              string                        `json:"pin,omitempty"`
	Compress         string                        `json:"compress,omitempty"`
	MD5              string                        `json:"md5,omitempty"`
	DeviceEndpoint   bool                          `json:"device_endpoint,omitempty"`
}

type UploadType string

const (
	UploadTypeUnknown          UploadType = ""
	UploadTypeHostFirmware     UploadType = "firmware"
	UploadTypeNotecardFirmware UploadType = "notecard"
	UploadTypeModemFirmware    UploadType = "modem"
	UploadTypeUserData         UploadType = "data"
)

var allFileTypes = []UploadType{
	UploadTypeUnknown,
	UploadTypeHostFirmware,
	UploadTypeNotecardFirmware,
	UploadTypeModemFirmware,
	UploadTypeUserData,
}

func ParseUploadType(fileType string) UploadType {
	for _, validType := range allFileTypes {
		if string(validType) == fileType {
			return validType
		}
	}
	return UploadTypeUnknown
}

// HubRequestFileFirmware is firmware-specific metadata
type HubRequestFileFirmware struct {
	// The organization accountable for the firmware - a display string
	Organization string `json:"org,omitempty"`
	// A description of the firmware - a display string
	Description string `json:"description,omitempty"`
	// The name and model number of the product containing the firmware - a display string
	Product string `json:"product,omitempty"`
	// The identifier of the only firmware that will be acceptable and downloaded to this device
	Firmware string `json:"firmware,omitempty"`
	// The composite version number of the firmware, generally major.minor.patch as a string
	Version string `json:"version,omitempty"`
	// The target CPU of the firmware (see notecard/src/board.h)
	Target string `json:"target,omitempty"`
	// The build number of the firmware, for numeric comparison
	Major uint32 `json:"ver_major,omitempty"`
	Minor uint32 `json:"ver_minor,omitempty"`
	Patch uint32 `json:"ver_patch,omitempty"`
	Build uint32 `json:"ver_build,omitempty"`
	// The build number of the firmware, generally just a date and time
	Built string `json:"built,omitempty"`
	// The entity who built or is responsible for the firmware - a display string
	Builder string `json:"builder,omitempty"`
}

func (metadata HubRequestFileFirmware) VersionString() string {
	return fmt.Sprintf("%d.%d.%d.%d", metadata.Major, metadata.Minor, metadata.Patch, metadata.Build)
}

// UploadMetadata is the body of the object uploaded for each file
type UploadMetadata struct {
	Name     string                  `json:"name,omitempty"`
	Length   int                     `json:"length,omitempty"`
	MD5      string                  `json:"md5,omitempty"`
	CRC32    uint32                  `json:"crc32,omitempty"`
	Created  int64                   `json:"created,omitempty"`
	Modified int64                   `json:"modified,omitempty"`
	Source   string                  `json:"source,omitempty"`
	Contains string                  `json:"contains,omitempty"`
	Found    string                  `json:"found,omitempty"`
	FileType UploadType              `json:"type,omitempty"`
	Tags     string                  `json:"tags,omitempty"`  // comma-separated, no spaces, case-insensitive
	Notes    string                  `json:"notes,omitempty"` // Should be simple text
	Firmware *HubRequestFileFirmware `json:"firmware,omitempty"`
	// Arbitrary metadata that the user may define - we don't interpret the schema at all
	Info map[string]interface{} `json:"info,omitempty"`
}

// HubRequestFileTagPublish indicates that this should be published in the UI
const HubRequestFileTagPublish = "publish"
