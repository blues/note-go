// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"github.com/blues/note-go/note"
)

// Request Types (L suffix means Legacy as of 2019-11-18, and can be removed after we ship)

// ReqFileAdd (golint)
const ReqFileAdd = "file.add"

// ReqFileSet (golint)
const ReqFileSet = "file.set"

// ReqFileDelete (golint)
const ReqFileDelete = "file.delete"

// ReqFileChanges (golint)
const ReqFileChanges = "file.changes"

// ReqFileChangesPending (golint)
const ReqFileChangesPending = "file.changes.pending"

// ReqFileGetL (golint)
const ReqFileGetL = "file.get"

// ReqFileSync (golint)
const ReqFileSync = "file.sync"

// ReqFileStats (golint)
const ReqFileStats = "file.stats"

// ReqServiceSync (golint)
const ReqServiceSync = "service.sync"

// ReqNotesGetL (golint)
const ReqNotesGetL = "notes.get"

// ReqNoteChanges (golint)
const ReqNoteChanges = "note.changes"

// ReqNoteAdd (golint)
const ReqNoteAdd = "note.add"

// ReqNoteEvent (golint)
const ReqNoteEvent = "note.event"

// ReqNoteTemplate (golint)
const ReqNoteTemplate = "note.template"

// ReqNoteGet (golint)
const ReqNoteGet = "note.get"

// ReqNoteUpdate (golint)
const ReqNoteUpdate = "note.update"

// ReqNoteDelete (golint)
const ReqNoteDelete = "note.delete"

// ReqCardTime (golint)
const ReqCardTime = "card.time"

// ReqCardContact (golint)
const ReqCardContact = "card.contact"

// ReqCardAttn (golint)
const ReqCardAttn = "card.attn"

// ReqCardVersion (golint)
const ReqCardVersion = "card.version"

// ReqCardStatus (golint)
const ReqCardStatus = "card.status"

// ReqCardRestart (golint)
const ReqCardRestart = "card.restart"

// ReqCardRestore (golint)
const ReqCardRestore = "card.restore"

// ReqCardLocation (golint)
const ReqCardLocation = "card.location"

// ReqCardLocationMode (golint)
const ReqCardLocationMode = "card.location.mode"

// ReqCardLocationTrack (golint)
const ReqCardLocationTrack = "card.location.track"

// ReqCardTemp (golint)
const ReqCardTemp = "card.temp"

// ReqCardVoltage (golint)
const ReqCardVoltage = "card.voltage"

// ReqCardMotion (golint)
const ReqCardMotion = "card.motion"

// ReqCardMotionMode (golint)
const ReqCardMotionMode = "card.motion.mode"

// ReqCardMotionSync (golint)
const ReqCardMotionSync = "card.motion.sync"

// ReqCardMotionTrack (golint)
const ReqCardMotionTrack = "card.motion.track"

// ReqCardIO (golint)
const ReqCardIO = "card.io"

// ReqCardTrace (golint)
const ReqCardTrace = "card.trace"

// ReqCardWireless (golint)
const ReqCardWireless = "card.wireless"

// ReqCardAUX (golint)
const ReqCardAUX = "card.aux"

// ReqCardUsageGet (golint)
const ReqCardUsageGet = "card.usage.get"

// ReqCardUsageTest (golint)
const ReqCardUsageTest = "card.usage.test"

// ReqCardUsageRate (golint)
const ReqCardUsageRate = "card.usage.rate"

// ReqServiceEnvL (golint)
const ReqServiceEnvL = "service.env"

// ReqServiceSet (golint)
const ReqServiceSet = "service.set"

// ReqServiceGet (golint)
const ReqServiceGet = "service.get"

// ReqServiceStatus (golint)
const ReqServiceStatus = "service.status"

// ReqServiceSignal (golint)
const ReqServiceSignal = "service.signal"

// ReqServiceSyncStatus (golint)
const ReqServiceSyncStatus = "service.sync.status"

// ReqEnvGet (golint)
const ReqEnvGet = "env.get"

// ReqEnvTime (golint)
const ReqEnvTime = "env.time"

// ReqEnvLocation (golint)
const ReqEnvLocation = "env.location"

// ReqWebGet (golint)
const ReqWebGet = "web.get"

// ReqWebPut (golint)
const ReqWebPut = "web.put"

// ReqWebPost (golint)
const ReqWebPost = "web.post"

// ReqDFUStatus (golint)
const ReqDFUStatus = "dfu.status"

// ReqDFUGet (golint)
const ReqDFUGet = "dfu.get"

// ReqDFUServiceGet (golint)
const ReqDFUServiceGet = "dfu.service.get"

// Request is the core API request/response data structure
type Request struct {
	Req              string                        `json:"req,omitempty"`
	Err              string                        `json:"err,omitempty"`
	RequestID        uint32                        `json:"id,omitempty"`
	NotefileID       string                        `json:"file,omitempty"`
	TrackerID        string                        `json:"tracker,omitempty"`
	NoteID           string                        `json:"note,omitempty"`
	Body             *map[string]interface{}       `json:"body,omitempty"`
	Payload          *[]byte                       `json:"payload,omitempty"`
	Deleted          bool                          `json:"deleted,omitempty"`
	Start            bool                          `json:"start,omitempty"`
	Stop             bool                          `json:"stop,omitempty"`
	Delete           bool                          `json:"delete,omitempty"`
	USB              bool                          `json:"usb,omitempty"`
	Connected        bool                          `json:"connected,omitempty"`
	Secure           bool                          `json:"secure,omitempty"`
	Unsecure         bool                          `json:"unsecure,omitempty"`
	Alert            bool                          `json:"alert,omitempty"`
	Retry            bool                          `json:"retry,omitempty"`
	Signals          int32                         `json:"signals,omitempty"`
	Max              int32                         `json:"max,omitempty"`
	Changes          int32                         `json:"changes,omitempty"`
	Seconds          int32                         `json:"seconds,omitempty"`
	SecondsV         string                        `json:"vseconds,omitempty"`
	Minutes          int32                         `json:"minutes,omitempty"`
	MinutesV         string                        `json:"vminutes,omitempty"`
	Hours            int32                         `json:"hours,omitempty"`
	HoursV           string                        `json:"vhours,omitempty"`
	Days             int32                         `json:"days,omitempty"`
	Result           int32                         `json:"result,omitempty"`
	I2C              int32                         `json:"i2c,omitempty"`
	Status           string                        `json:"status,omitempty"`
	Version          string                        `json:"version,omitempty"`
	Name             string                        `json:"name,omitempty"`
	Org              string                        `json:"org,omitempty"`
	Role             string                        `json:"role,omitempty"`
	Email            string                        `json:"email,omitempty"`
	Area             string                        `json:"area,omitempty"`
	Country          string                        `json:"country,omitempty"`
	Zone             string                        `json:"zone,omitempty"`
	Mode             string                        `json:"mode,omitempty"`
	Host             string                        `json:"host,omitempty"`
	Movements        string                        `json:"movements,omitempty"`
	ProductUID       string                        `json:"product,omitempty"`
	DeviceUID        string                        `json:"device,omitempty"`
	RouteUID         string                        `json:"route,omitempty"`
	Files            *[]string                     `json:"files,omitempty"`
	FileInfo         *map[string]note.NotefileInfo `json:"info,omitempty"`
	Notes            *map[string]note.Info         `json:"notes,omitempty"`
	Pad              int32                         `json:"pad,omitempty"`
	Storage          int32                         `json:"storage,omitempty"`
	LocationOLC      string                        `json:"olc,omitempty"`
	Latitude         float64                       `json:"lat,omitempty"`
	Longitude        float64                       `json:"lon,omitempty"`
	Value            float64                       `json:"value,omitempty"`
	ValueV           string                        `json:"vvalue,omitempty"`
	SN               string                        `json:"sn,omitempty"`
	Text             string                        `json:"text,omitempty"`
	Offset           int32                         `json:"offset,omitempty"`
	Length           int32                         `json:"length,omitempty"`
	Total            int32                         `json:"total,omitempty"`
	BytesSent        uint32                        `json:"bytes_sent,omitempty"`
	BytesReceived    uint32                        `json:"bytes_received,omitempty"`
	NotesSent        uint32                        `json:"notes_sent,omitempty"`
	NotesReceived    uint32                        `json:"notes_received,omitempty"`
	SessionsStandard uint32                        `json:"sessions_standard,omitempty"`
	SessionsSecure   uint32                        `json:"sessions_secure,omitempty"`
	Megabytes        int32                         `json:"megabytes,omitempty"`
	BytesPerDay      int32                         `json:"bytes_per_day,omitempty"`
	DataRate         float64                       `json:"rate,omitempty"`
	NumBytes         int32                         `json:"bytes,omitempty"`
	Template         bool                          `json:"template,omitempty"`
	BodyTemplate     string                        `json:"body_template,omitempty"`
	PayloadTemplate  int32                         `json:"payload_template,omitempty"`
	Allow            bool                          `json:"allow,omitempty"`
	Align            bool                          `json:"align,omitempty"`
	Limit            bool                          `json:"limit,omitempty"`
	Pending          bool                          `json:"pending,omitempty"`
	ReqTime          bool                          `json:"reqtime,omitempty"`
	ReqLoc           bool                          `json:"reqloc,omitempty"`
	Trace            string                        `json:"trace,omitempty"`
	Usage            *[]string                     `json:"usage,omitempty"`
	State            *[]PinState                   `json:"state,omitempty"`
	Time             int64                         `json:"time,omitempty"`
	VMin             float64                       `json:"vmin,omitempty"`
	VMax             float64                       `json:"vmax,omitempty"`
	VAvg             float64                       `json:"vavg,omitempty"`
	Daily            float64                       `json:"daily,omitempty"`
	Weekly           float64                       `json:"weekly,omitempty"`
	Montly           float64                       `json:"monthly,omitempty"`
	Verify           bool                          `json:"verify,omitempty"`
	Set              bool                          `json:"set,omitempty"`
	Reset            bool                          `json:"reset,omitempty"`
	Calibration      float64                       `json:"calibration,omitempty"`
	Heartbeat        bool                          `json:"heartbeat,omitempty"`
	Threshold        int32                         `json:"threshold,omitempty"`
	Count            uint32                        `json:"count,omitempty"`
	Sync             bool                          `json:"sync,omitempty"`
	Live             bool                          `json:"live,omitempty"`
	Type             int32                         `json:"type,omitempty"`
	Number           int64                         `json:"number,omitempty"`
	SKU              string                        `json:"sku,omitempty"`
	Net              *NetInfo                      `json:"net,omitempty"`
	Sensitivity      int32                         `json:"sensitivity,omitempty"`
	Requested        int32                         `json:"requested,omitempty"`
	Completed        int32                         `json:"completed,omitempty"`
}

// PinState describes the state of an AUX pin for hardware-related Notecard requests
type PinState struct {
	High  bool     `json:"high,omitempty"`
	Low   bool     `json:"low,omitempty"`
	Count []uint32 `json:"count,omitempty"`
}

// SyncLogLevelMajor is just major events
const SyncLogLevelMajor = 0

// SyncLogLevelMinor is just major and minor events
const SyncLogLevelMinor = 1

// SyncLogLevelDetail is major, minor, and detailed events
const SyncLogLevelDetail = 2

// SyncLogLevelProg is everything plus programmatically-targeted
const SyncLogLevelProg = 3

// SyncLogLevelAll is all events
const SyncLogLevelAll = SyncLogLevelProg

// SyncLogLevelNone is no events
const SyncLogLevelNone = -1

// SyncLogNotefile is the special notefile containing sync log info
const SyncLogNotefile = "_synclog.qi"

// SyncLogBody is the data structure used in the SyncLogNotefile
type SyncLogBody struct {
	TimeSecs    int64  `json:"time,omitempty"`
	BootMs      int64  `json:"sequence,omitempty"`
	DetailLevel uint32 `json:"level,omitempty"`
	Subsystem   string `json:"subsystem,omitempty"`
	Text        string `json:"text,omitempty"`
}
