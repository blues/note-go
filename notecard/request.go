// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"github.com/blues/note-go/note"
)

// Request Types

// ReqFilesAdd (golint)
const ReqFilesAdd =     "files.add"
// ReqFilesSet (golint)
const ReqFilesSet =     "files.set"
// ReqFilesDelete (golint)
const ReqFilesDelete =  "files.delete"
// ReqFilesGet (golint)
const ReqFilesGet =     "files.get"
// ReqFilesSync (golint)
const ReqFilesSync =    "files.sync"
// ReqServiceSync (golint)
const ReqServiceSync =  "service.sync"
// ReqNotesGet (golint)
const ReqNotesGet =     "notes.get"
// ReqNoteAdd (golint)
const ReqNoteAdd =      "note.add"
// ReqNoteEvent (golint)
const ReqNoteEvent =    "note.event"
// ReqNoteTemplate (golint)
const ReqNoteTemplate = "note.template"
// ReqNoteGet (golint)
const ReqNoteGet =      "note.get"
// ReqNoteUpdate (golint)
const ReqNoteUpdate =   "note.update"
// ReqNoteDelete (golint)
const ReqNoteDelete =   "note.delete"
// ReqCardTime (golint)
const ReqCardTime =		"card.time"
// ReqCardContact (golint)
const ReqCardContact =	"card.contact"
// ReqCardAttn (golint)
const ReqCardAttn =		"card.attn"
// ReqCardVersion (golint)
const ReqCardVersion =	"card.version"
// ReqCardStatus (golint)
const ReqCardStatus =	"card.status"
// ReqCardRestart (golint)
const ReqCardRestart =	"card.restart"
// ReqCardRestore (golint)
const ReqCardRestore =	"card.restore"
// ReqCardLocation (golint)
const ReqCardLocation =	"card.location"
// ReqCardLocationMode (golint)
const ReqCardLocationMode =	"card.location.mode"
// ReqCardTemp (golint)
const ReqCardTemp =		"card.temp"
// ReqCardVoltage (golint)
const ReqCardVoltage =	"card.voltage"
// ReqCardIO (golint)
const ReqCardIO =		"card.io"
// ReqCardAUX (golint)
const ReqCardAUX =		"card.aux"
// ReqCardUsageGet (golint)
const ReqCardUsageGet =	"card.usage.get"
// ReqCardUsageTest (golint)
const ReqCardUsageTest = "card.usage.test"
// ReqCardUsageRate (golint)
const ReqCardUsageRate = "card.usage.rate"
// ReqServiceEnv (golint)
const ReqServiceEnv =	"service.env"
// ReqServiceSet (golint)
const ReqServiceSet =	"service.set"
// ReqServiceGet (golint)
const ReqServiceGet =	"service.get"
// ReqServiceStatus (golint)
const ReqServiceStatus = "service.status"
// ReqServiceSignal (golint) 
const ReqServiceSignal = "service.signal"
// ReqServiceSyncStatus (golint)
const ReqServiceSyncStatus = "service.sync.status"
// ReqWebGet (golint)
const ReqWebGet =		"web.get"
// ReqWebPut (golint)
const ReqWebPut =		"web.put"
// ReqWebPost (golint)
const ReqWebPost =		"web.post"
// ReqDFUStatus (golint)
const ReqDFUStatus =	"dfu.status"
// ReqDFUGet (golint)
const ReqDFUGet =		"dfu.get"
// ReqDFUServiceGet (golint)
const ReqDFUServiceGet = "dfu.service.get"

// Request is the core API request/response data structure
type Request struct {
    Req string					`json:"req,omitempty"`
    Err string					`json:"err,omitempty"`
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
    USB bool					`json:"usb,omitempty"`
    Connected bool				`json:"connected,omitempty"`
    Secure bool					`json:"secure,omitempty"`
    Signals int32				`json:"signals,omitempty"`
    Max int32					`json:"max,omitempty"`
    Changes int32				`json:"changes,omitempty"`
    Seconds int32				`json:"seconds,omitempty"`
    Minutes int32				`json:"minutes,omitempty"`
    Hours int32					`json:"hours,omitempty"`
    Days int32					`json:"days,omitempty"`
    Result int32				`json:"result,omitempty"`
    Port int32					`json:"port,omitempty"`
    Status string				`json:"status,omitempty"`
    Version string				`json:"version,omitempty"`
    Name string 				`json:"name,omitempty"`
	Org string					`json:"org,omitempty"`
	Role string					`json:"role,omitempty"`
	Email string				`json:"email,omitempty"`
    Area string 				`json:"area,omitempty"`
    Country string 				`json:"country,omitempty"`
    Zone string 				`json:"zone,omitempty"`
    Mode string					`json:"mode,omitempty"`
    Host string					`json:"host,omitempty"`
    Target string				`json:"target,omitempty"`
	ProductUID string			`json:"product,omitempty"`
    DeviceUID string			`json:"device,omitempty"`
    Route string				`json:"route,omitempty"`
    Files *[]string				`json:"files,omitempty"`
    FileInfo *map[string]note.NotefileInfo `json:"info,omitempty"`
    Notes *map[string]note.Info `json:"notes,omitempty"`
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
    Align bool					`json:"align,omitempty"`
    Limit bool					`json:"limit,omitempty"`
    ReqTime bool				`json:"reqtime,omitempty"`
    ReqLoc bool					`json:"reqloc,omitempty"`
    Trace string				`json:"trace,omitempty"`
    Usage *[]string				`json:"usage,omitempty"`
    State *[]PinState			`json:"state,omitempty"`
	Serial string				`json:"serial,omitempty"`
    Time uint32					`json:"time,omitempty"`
    VMin float64				`json:"vmin,omitempty"`
    VMax float64				`json:"vmax,omitempty"`
    VAvg float64				`json:"vavg,omitempty"`
    Daily float64				`json:"daily,omitempty"`
    Weekly float64				`json:"weekly,omitempty"`
    Montly float64				`json:"monthly,omitempty"`
    Verify bool					`json:"verify,omitempty"`
}

// PinState describes the state of an AUX pin for hardware-related Notecard requests
type PinState struct {
	High bool					`json:"high,omitempty"`
	Count []uint32				`json:"count,omitempty"`
}
