// Copyright 2019 Blues Inc.  All rights reserved.  
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notehub

import (
	"github.com/blues/note-go/notecard"
)

// Supported requests

// HubDeviceMonitor (golint)
const HubDeviceMonitor  = "hub.device.monitor"
// HubDeviceSignal (golint)
const HubDeviceSignal  = "hub.device.signal"
// HubQuery (golint)
const HubQuery          = "hub.app.data.query"
// HubAppUpload (golint)
const HubAppUpload      = "hub.app.upload.add"
// HubAppUploads (golint)
const HubAppUploads     = "hub.app.upload.query"
// HubAppUploadSet (golint)
const HubAppUploadSet	= "hub.app.upload.set"
// HubAppUploadDelete (golint)
const HubAppUploadDelete = "hub.app.upload.delete"
// HubAppUploadRead (golint)
const HubAppUploadRead  = "hub.app.upload.get"
// HubAppMonitor (golint)
const HubAppMonitor		= "hub.app.monitor"
// HubAppHandlers (golint)
const HubAppHandlers	= "hub.app.handlers"

// HubRequest is is the core data structure for notehub-specific requests
type HubRequest struct {
	notecard.Request			`json:",omitempty"`
	AppUID string				`json:"app,omitempty"`
	FleetUID string				`json:"fleet,omitempty"`
	*DbQuery					`json:",omitempty"`
	Uploads *[]HubRequestFile	`json:"uploads,omitempty"`
	Contains string				`json:"contains,omitempty"`
	Handlers *[]string			`json:"handlers,omitempty"`
    Card bool					`json:"card,omitempty"`
}

// HubRequestFile is is the body of the object uploaded for each file
type HubRequestFile struct {
	Name string					`json:"name,omitempty"`
	Length int					`json:"length,omitempty"`
	MD5 string					`json:"md5,omitempty"`
	CRC32 uint32				`json:"crc32,omitempty"`
	Created int64				`json:"created,omitempty"`
	Source string				`json:"source,omitempty"`
	Contains string				`json:"contains,omitempty"`
	Found string				`json:"found,omitempty"`
	Card bool					`json:"card,omitempty"`
    Info *interface{}           `json:"info,omitempty"`
}
