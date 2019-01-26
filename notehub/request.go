// Copyright 2019 Blues Inc.  All rights reserved.  
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notehub

import (
	"github.com/rayozzie/note-go/note"
)

// Supported requests

// HubRequest is is the core data structure for notehub-specific requests
type Request struct {
	note.Request				`json:",omitempty"`
	*DbQuery					`json:",omitempty"`
	Uploads *[]RequestFile		`json:"uploads,omitempty"`
}

// HubRequestFile is is the body of the object uploaded for each file
type RequestFile struct {
	Name string					`json:"name,omitempty"`
	Length int					`json:"length,omitempty"`
	MD5 string					`json:"md5,omitempty"`
	CRC32 uint32				`json:"crc32,omitempty"`
	Created int64				`json:"created,omitempty"`
	Source string				`json:"source,omitempty"`
	Contains string				`json:"contains,omitempty"`
	Found string				`json:"found,omitempty"`
    Info *interface{}           `json:"info,omitempty"`
}
