// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"github.com/rayozzie/note-go/note"
)

// Request Types

// ReqFilesSet (golint)
const ReqFilesSet =     "files.set"
// ReqFilesAdd (golint)
const ReqFilesAdd =     "files.add"
// ReqFilesDelete (golint)
const ReqFilesDelete =  "files.delete"
// ReqFilesGet (golint)
const ReqFilesGet =     "files.get"
// ReqNotesGet (golint)
const ReqNotesGet =     "notes.get"
// ReqNoteAdd (golint)
const ReqNoteAdd =      "note.add"
// ReqNoteGet (golint)
const ReqNoteGet =      "note.get"
// ReqNoteUpdate (golint)
const ReqNoteUpdate =   "note.update"
// ReqNoteDelete (golint)
const ReqNoteDelete =   "note.delete"
// ReqCardIO (golint)
const ReqCardIO =		"card.io"

// CardRequest is the core API request/response data structure
type CardRequest struct {
    Request string              `json:"req,omitempty"`
    Error string				`json:"err,omitempty"`
    RequestID uint32            `json:"id,omitempty"`
    NotefileID string           `json:"file,omitempty"`
    TrackerID string            `json:"tracker,omitempty"`
    NoteID string               `json:"note,omitempty"`
    Body *interface{}           `json:"body,omitempty"`
    Payload *[]byte             `json:"payload,omitempty"`
    Deleted bool                `json:"deleted,omitempty"`
    Start bool                  `json:"start,omitempty"`
    Stop bool                   `json:"stop,omitempty"`
    Delete bool                 `json:"delete,omitempty"`
    Max int32					`json:"max,omitempty"`
    Changes int32				`json:"changes,omitempty"`
    Seconds int32				`json:"seconds,omitempty"`
    Minutes int32				`json:"minutes,omitempty"`
    Hours int32					`json:"hours,omitempty"`
    Days int32					`json:"days,omitempty"`
    Result int32				`json:"result,omitempty"`
    Port int32					`json:"port,omitempty"`
    Status string				`json:"status,omitempty"`
    Name string 				`json:"name,omitempty"`
    Mode string					`json:"mode,omitempty"`
    Host string					`json:"host,omitempty"`
    Target string				`json:"target,omitempty"`
	ProductUID string			`json:"product,omitempty"`
    DeviceUID string			`json:"device,omitempty"`
    Route string				`json:"route,omitempty"`
    Files *[]string				`json:"files,omitempty"`
    FileInfo *map[string]note.NotefileInfo `json:"info,omitempty"`
    Notes *map[string]note.NoteInfo  `json:"notes,omitempty"`
    Pad int32					`json:"pad,omitempty"`
    Storage int32				`json:"storage,omitempty"`
    LocationOLC string			`json:"olc,omitempty"`
    Latitude float64			`json:"lat,omitempty"`
    Longitude float64			`json:"lon,omitempty"`
    Value float64               `json:"value,omitempty"`
	Wireless string				`json:"wireless,omitempty"`
	SN string					`json:"sn,omitempty"`
	Text string					`json:"text,omitempty"`
    Offset int32				`json:"offset,omitempty"`
    Length int32				`json:"length,omitempty"`
    Total int32					`json:"total,omitempty"`
    BytesSent uint32			`json:"bytes_sent,omitempty"`
    BytesReceived uint32		`json:"bytes_received,omitempty"`
    NotesSent uint32			`json:"notes_sent,omitempty"`
    NotesReceived uint32		`json:"notes_received,omitempty"`
    SessionsStandard uint32		`json:"sessions_standard,omitempty"`
    SessionsSecure uint32		`json:"sessions_secure,omitempty"`
    Megabytes int32				`json:"megabytes,omitempty"`
    BytesPerDay int32			`json:"bytes_per_day,omitempty"`
    DataRate float64			`json:"rate,omitempty"`
    NumBytes int32				`json:"bytes,omitempty"`
    Template bool               `json:"template,omitempty"`
	BodyTemplate string	        `json:"body_template,omitempty"`
    PayloadTemplate int32		`json:"payload_template,omitempty"`
    Allow bool					`json:"allow,omitempty"`
    Trace string				`json:"trace,omitempty"`
    Usage *[]string				`json:"usage,omitempty"`
    State *[]PinState			`json:"state,omitempty"`
	Serial string				`json:"serial,omitempty"`
    Time uint32					`json:"time,omitempty"`
}

// PinState describes the state of an AUX pin for hardware-related Notecard requests
type PinState struct {
	High bool					`json:"high,omitempty"`
	Count []uint32				`json:"count,omitempty"`
}
