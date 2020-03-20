// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notehub

import (
	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
)

// Supported requests

// HubDeviceContact (golint)
const HubDeviceContact = "hub.device.contact"

// HubQuery (golint)
const HubQuery = "hub.app.data.query"

// HubAppUpload (golint)
const HubAppUpload = "hub.app.upload.add"

// HubAppUploads (golint)
const HubAppUploads = "hub.app.upload.query"

// HubAppUploadSet (golint)
const HubAppUploadSet = "hub.app.upload.set"

// HubAppUploadDelete (golint)
const HubAppUploadDelete = "hub.app.upload.delete"

// HubAppUploadRead (golint)
const HubAppUploadRead = "hub.app.upload.get"

// HubRequest is is the core data structure for notehub-specific requests
type HubRequest struct {
	notecard.Request `json:",omitempty"`
	Contact			 *note.EventContact `json:"contact,omitempty"`
	AppUID           string            `json:"app,omitempty"`
	FleetUID         string            `json:"fleet,omitempty"`
	DbQuery          *DbQuery          `json:"query,omitempty"`
	Uploads          *[]HubRequestFile `json:"uploads,omitempty"`
	Contains         string            `json:"contains,omitempty"`
	Handlers         *[]string         `json:"handlers,omitempty"`
	FileType         string            `json:"type,omitempty"`
	FileTags         string            `json:"tags,omitempty"`
	FileNotes        string            `json:"notes,omitempty"`
}

// File Types

// HubFileTypeUnknown (golint)
const HubFileTypeUnknown = ""

// HubFileTypeUserFirmware (golint)
const HubFileTypeUserFirmware = "firmware"

// HubFileTypeCardFirmware (golint)
const HubFileTypeCardFirmware = "notecard"

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
	// The composite version number of the firmware, generally major.minor as a string
	Version string `json:"version,omitempty"`
	// The build number of the firmware, for numeric comparison
	Major uint16 `json:"ver_major,omitempty"`
	Minor uint16 `json:"ver_minor,omitempty"`
	Patch uint16 `json:"ver_patch,omitempty"`
	Build uint16 `json:"ver_build,omitempty"`
	// The build number of the firmware, generally just a date and time
	Built string `json:"built,omitempty"`
	// The entity who built or is responsible for the firmware - a display string
	Builder string `json:"builder,omitempty"`
}

// HubRequestFile is is the body of the object uploaded for each file
type HubRequestFile struct {
	Name     string                  `json:"name,omitempty"`
	Length   int                     `json:"length,omitempty"`
	MD5      string                  `json:"md5,omitempty"`
	CRC32    uint32                  `json:"crc32,omitempty"`
	Created  int64                   `json:"created,omitempty"`
	Modified int64                   `json:"modified,omitempty"`
	Source   string                  `json:"source,omitempty"`
	Contains string                  `json:"contains,omitempty"`
	Found    string                  `json:"found,omitempty"`
	FileType string                  `json:"type,omitempty"`
	Tags     string                  `json:"tags,omitempty"`  // comma-separated, no spaces, case-insensitive
	Notes    string                  `json:"notes,omitempty"` // markdown
	Firmware *HubRequestFileFirmware `json:"firmware,omitempty"`
	// Arbitrary metadata that the user may define - we don't interpret the schema at all
	Info *map[string]interface{} `json:"info,omitempty"`
}

// HubRequestFileTagPublish indicates that this should be published in the UI
const HubRequestFileTagPublish = "publish"
