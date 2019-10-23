// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

import (
	"encoding/json"
	"time"
)

// DefaultDeviceEndpointID is the default endpoint name of the edge, chosen for its length in protocol messages
const DefaultDeviceEndpointID = ""

// DefaultHubEndpointID is the default endpoint name of the hub, chosen for its length in protocol messages
const DefaultHubEndpointID = "1"

// Note is the most fundamental data structure, containing
// user data referred to as its "body" and its "payload".  All
// access to these fields, and changes to these fields, must
// be done indirectly through the note API.
type Note struct {
	Body      map[string]interface{} `json:"b,omitempty"`
	Payload   []byte                 `json:"p,omitempty"`
	Change    int64                  `json:"c,omitempty"`
	Histories *[]History             `json:"h,omitempty"`
	Conflicts *[]Note                `json:"x,omitempty"`
	Updates   int32                  `json:"u,omitempty"`
	Deleted   bool                   `json:"d,omitempty"`
	Sent      bool                   `json:"s,omitempty"`
	Bulk      bool                   `json:"k,omitempty"`
	XPOff     uint32                 `json:"O,omitempty"`
	XPLen     uint32                 `json:"L,omitempty"`
}

// History records the update history, optimized so that if the most recent entry
// is by the same endpoint as an update/delete, that entry is re-used.  The primary use
// of History is for conflict detection, and you don't need to detect conflicts
// against yourself.
type History struct {
	When       int64  `json:"w,omitempty"`
	Where      string `json:"l,omitempty"`
	EndpointID string `json:"e,omitempty"`
	Sequence   int32  `json:"s,omitempty"`
}

// Info is a general "content" structure
type Info struct {
	Body    *map[string]interface{} `json:"body,omitempty"`
	Payload *[]byte                 `json:"payload,omitempty"`
	Deleted bool                    `json:"deleted,omitempty"`
}

// CreateNote creates the core data structure for an object, given a JSON body
func CreateNote(body []byte, payload []byte) (newNote Note, err error) {
	newNote.Payload = payload
	err = newNote.SetBody(body)
	return
}

// SetBody sets the application-supplied Body field of a given Note given some JSON
func (note *Note) SetBody(body []byte) (err error) {
	if body == nil {
		note.Body = nil
	} else {
		note.Body = map[string]interface{}{}
		err = json.Unmarshal(body, &note.Body)
		if err != nil {
			return
		}
	}
	return
}

// JSONToBody unmarshals the specify object and returns it as a map[string]interface{}
func JSONToBody(bodyJSON []byte) (body map[string]interface{}, err error) {
	err = json.Unmarshal(bodyJSON, &body)
	return
}

// ObjectToJSON Marshals the specify object and returns it as a []byte
func ObjectToJSON(object interface{}) (bodyJSON []byte, err error) {
	bodyJSON, err = json.Marshal(object)
	return
}

// ObjectToBody Marshals the specify object and returns it as map
func ObjectToBody(object interface{}) (body map[string]interface{}, err error) {
	var bodyJSON []byte
	bodyJSON, err = json.Marshal(object)
	if err == nil {
		err = json.Unmarshal(bodyJSON, &body)
	}
	return
}

// SetPayload sets the application-supplied Payload field of a given Note,
// which must be binary bytes that will ultimately be rendered as base64 in JSON
func (note *Note) SetPayload(payload []byte) {
	note.Payload = payload
}

// Close closes and frees the object on a note {
func (note *Note) Close() {
}

// Dup duplicates the note
func (note *Note) Dup() Note {
	newNote := *note
	return newNote
}

// GetBody retrieves the application-specific Body of a given Note
func (note *Note) GetBody() []byte {
	data, err := json.Marshal(note.Body)
	if err != nil {
		data = []byte{}
	}
	return data
}

// GetPayload retrieves the Payload from a given Note
func (note *Note) GetPayload() []byte {
	return note.Payload
}

// EndpointID determines the endpoint that last modified the note
func (note *Note) EndpointID() string {
	if note.Histories == nil {
		return ""
	}
	histories := *note.Histories
	if len(histories) == 0 {
		return ""
	}
	return histories[0].EndpointID
}

// HasConflicts determines whether or not a given Note has conflicts
func (note *Note) HasConflicts() bool {
	if note.Conflicts == nil {
		return false
	}
	return len(*note.Conflicts) != 0
}

// GetConflicts fetches the conflicts, so that they may be displayed
func (note *Note) GetConflicts() []Note {
	if note.Conflicts == nil {
		return []Note{}
	}
	return *note.Conflicts
}

// GetModified retrieves information about the note's modification
func (note *Note) GetModified() (isAvailable bool, endpointID string, when string, where string, updates int32) {
	if note.Histories == nil || len(*note.Histories) == 0 {
		return
	}
	histories := *note.Histories
	endpointID = histories[0].EndpointID
	when = time.Unix(0, histories[0].When*1000000000).UTC().Format("2006-01-02T15:04:05Z")
	where = histories[0].Where
	updates = histories[0].Sequence
	isAvailable = true
	return
}
