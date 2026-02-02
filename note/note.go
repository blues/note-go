// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"time"
)

// DefaultDeviceEndpointID is the default endpoint name of the edge, chosen for its length in protocol messages
const DefaultDeviceEndpointID = ""

// DefaultHubEndpointID is the default endpoint name of the hub, chosen for its length in protocol messages
const DefaultHubEndpointID = "1"

// HubDefaultInboundNotefile is the hard-wired default notefile for user data
const HubDefaultInboundNotefile = "data.qi"

// HubDefaultOutboundNotefile is the hard-wired default notefile for user data
const HubDefaultOutboundNotefile = "data.qo"

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
	Tower     *TowerLocation         `json:"T,omitempty"`
}

// History records the update history, optimized so that if the most recent entry
// is by the same endpoint as an update/delete, that entry is re-used.  The primary use
// of History is for conflict detection, and you don't need to detect conflicts
// against yourself.
type History struct {
	When       int64  `json:"w,omitempty"`
	Where      string `json:"l,omitempty"`
	WhereWhen  int64  `json:"m,omitempty"`
	EndpointID string `json:"e,omitempty"`
	Sequence   int32  `json:"s,omitempty"`
}

// Info is a general "content" structure
type Info struct {
	NoteID    string                  `json:"id,omitempty"`
	When      int64                   `json:"time,omitempty"`
	WhereLat  float64                 `json:"lat,omitempty"`
	WhereLon  float64                 `json:"lon,omitempty"`
	WhereWhen int64                   `json:"ltime,omitempty"`
	Body      *map[string]interface{} `json:"body,omitempty"`
	Payload   *[]byte                 `json:"payload,omitempty"`
	Deleted   bool                    `json:"deleted,omitempty"`
	Edge      bool                    `json:"edge,omitempty"`
	Pending   bool                    `json:"pending,omitempty"`
}

// CreateNote creates the core data structure for an object, given a JSON body
func CreateNote(body []byte, payload []byte) (newNote Note, err error) {
	newNote.Payload = payload
	err = newNote.SetBody(body)
	return
}

// SetBody sets the application-supplied Body field of a given Note given some JSON
func (note *Note) SetBody(body []byte) (err error) {
	if len(body) == 0 {
		note.Body = nil
	} else {
		note.Body = map[string]interface{}{}
		err = JSONUnmarshal(body, &note.Body)
		if err != nil {
			return
		}
	}
	return
}

// JSONToBody unmarshals the specified object and returns it as a map[string]interface{}
func JSONToBody(bodyJSON []byte) (body map[string]interface{}, err error) {
	err = JSONUnmarshal(bodyJSON, &body)
	return
}

// ObjectToJSON Marshals the specified object and returns it as a []byte
func ObjectToJSON(object interface{}) (bodyJSON []byte, err error) {
	bodyJSON, err = JSONMarshal(object)
	return
}

// ObjectToBody Marshals the specified object and returns it as map
func ObjectToBody(object interface{}) (body map[string]interface{}, err error) {
	var bodyJSON []byte
	bodyJSON, err = JSONMarshal(object)
	if err == nil {
		err = JSONUnmarshal(bodyJSON, &body)
	}
	return
}

// BodyToObject Unmarshals the specified map into an object
func BodyToObject(body *map[string]interface{}, object interface{}) (err error) {
	if body == nil {
		return
	}
	var bodyJSON []byte
	bodyJSON, err = JSONMarshal(body)
	if err == nil {
		err = JSONUnmarshal(bodyJSON, object)
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
	if note.Body == nil {
		return []byte("{}")
	}
	data, err := JSONMarshal(note.Body)
	if err != nil {
		return []byte("{}")
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

// GetWhen retrieves the epoch modification time
func (note *Note) When() (when int64) {
	if note.Histories == nil || len(*note.Histories) == 0 {
		return 0
	}
	h := (*note.Histories)[0]
	if h.When < 1483228800 || h.When > math.MaxUint32 {
		// Before 1/1/2017 or can't fit into a uint32
		h.When = 0
	}
	return h.When
}

// GetEndpointID retrieves the endpoint that last modified the note
func (note *Note) GetEndpointID() (endpointID string) {
	if note.Histories == nil || len(*note.Histories) == 0 {
		return ""
	}
	h := (*note.Histories)[0]
	return h.EndpointID
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

// JSONUnmarshal uses JSON Numbers, rather than assuming Floats.  This fixes an issue
// in which, when decoding to an arbitrary interface, the JSON package decodes
// large numbers (like Unix epoch) into floats.
func JSONUnmarshal(data []byte, v interface{}) (err error) {
	d := json.NewDecoder(strings.NewReader(string(data)))
	d.UseNumber()
	return d.Decode(v)
}

// JSONMarshal is the equivalent to the json package's Marshal, however it does not escape HTML
// sitting inside JSON strings.
func JSONMarshal(v interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	clean := bytes.TrimSuffix(buffer.Bytes(), []byte("\n"))
	return clean, err
}

// JSONMarshalIndent is like Marshal but applies Indent to format the output.
// Each JSON element in the output will begin on a new line beginning with prefix
// followed by one or more copies of indent according to the indentation nesting.
func JSONMarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	b, err := JSONMarshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, b, prefix, indent)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
