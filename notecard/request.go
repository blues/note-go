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

// ReqFileGetL (golint)
const ReqFileGetL = "file.get"

// ReqFileChanges (golint)
const ReqFileChanges = "file.changes"

// ReqFileChangesPending (golint)
const ReqFileChangesPending = "file.changes.pending"

// ReqFileSync (golint)
const ReqFileSync = "file.sync"

// ReqFileStats (golint)
const ReqFileStats = "file.stats"

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

// ReqNoteEncrypt (golint)
const ReqNoteEncrypt = "note.encrypt"

// ReqNoteDecrypt (golint)
const ReqNoteDecrypt = "note.decrypt"

// ReqCardTime (golint)
const ReqCardTime = "card.time"

// ReqCardRandom (golint)
const ReqCardRandom = "card.random"

// ReqCardContact (golint)
const ReqCardContact = "card.contact"

// ReqCardAttn (golint)
const ReqCardAttn = "card.attn"

// ReqCardStatus (golint)
const ReqCardStatus = "card.status"

// ReqCardRestart (golint)
const ReqCardRestart = "card.restart"

// ReqCardCheckpoint (golint)
const ReqCardCheckpoint = "card.checkpoint"

// ReqCardRestore (golint)
const ReqCardRestore = "card.restore"

// ReqCardLocation (golint)
const ReqCardLocation = "card.location"

// ReqCardLocationMode (golint)
const ReqCardLocationMode = "card.location.mode"

// ReqCardLocationTrack (golint)
const ReqCardLocationTrack = "card.location.track"

// ReqCardTriangulate (golint)
const ReqCardTriangulate = "card.triangulate"

// ReqCardTemp (golint)
const ReqCardTemp = "card.temp"

// ReqCardIllumination (golint)
const ReqCardIllumination = "card.illumination"

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

// ReqCardAUX (golint)
const ReqCardAUX = "card.aux"

// ReqCardMonitor (golint)
const ReqCardMonitor = "card.monitor"

// ReqCardCarrier (golint)
const ReqCardCarrier = "card.carrier"

// ReqCardReserve (golint)
const ReqCardReserve = "card.reserve"

// ReqCardTrace (golint)
const ReqCardTrace = "card.trace"

// ReqCardUsageGet (golint)
const ReqCardUsageGet = "card.usage.get"

// ReqCardUsageTest (golint)
const ReqCardUsageTest = "card.usage.test"

// ReqCardUsageRate (golint)
const ReqCardUsageRate = "card.usage.rate"

// ReqEnvModified (golint)
const ReqEnvModified = "env.modified"

// ReqEnvGet (golint)
const ReqEnvGet = "env.get"

// ReqEnvSet (golint)
const ReqEnvSet = "env.set"

// ReqEnvDefault (golint)
const ReqEnvDefault = "env.default"

// ReqEnvTime (golint)
const ReqEnvTime = "env.time"

// ReqEnvLocation (golint)
const ReqEnvLocation = "env.location"

// ReqEnvSync (golint)
const ReqEnvSync = "env.sync"

// ReqWeb (golint)
const ReqWeb = "web"

// ReqWebGet (golint)
const ReqWebGet = "web.get"

// ReqWebPut (golint)
const ReqWebPut = "web.put"

// ReqWebPost (golint)
const ReqWebPost = "web.post"

// ReqWebDelete (golint)
const ReqWebDelete = "web.delete"

// ReqDFUStatus (golint)
const ReqDFUStatus = "dfu.status"

// ReqDFUGet (golint)
const ReqDFUGet = "dfu.get"

// ReqDFUPut (golint)
const ReqDFUPut = "dfu.put"

// ReqCardDFU (golint)
const ReqCardDFU = "card.dfu"

// ReqEnvVersion (golint)
const ReqEnvVersion = "env.version"

// ReqCardVersion (golint)
const ReqCardVersion = "card.version"

// ReqCardBootloader (golint)
const ReqCardBootloader = "card.bootloader"

// ReqCardTest (golint)
const ReqCardTest = "card.test"

// ReqCardSetup (golint)
const ReqCardSetup = "card.setup"

// ReqCardWireless (golint)
const ReqCardWireless = "card.wireless"

// ReqCardWirelessPenalty (golint)
const ReqCardWirelessPenalty = "card.wireless.penalty"

// ReqCardWiFi (golint)
const ReqCardWiFi = "card.wifi"

// ReqCardLog (golint)
const ReqCardLog = "card.log"

// ReqHubSync (golint)
const ReqHubSync = "hub.sync"

// ReqHubSyncL (golint)
const ReqHubSyncL = "service.sync"

// ReqHubLog (golint)
const ReqHubLog = "hub.log"

// ReqHubLogL (golint)
const ReqHubLogL = "service.log"

// ReqHubEnvL (golint)
const ReqHubEnvL = "hub.env"

// ReqHubEnvLL (golint)
const ReqHubEnvLL = "service.env"

// ReqHubSet (golint)
const ReqHubSet = "hub.set"

// ReqHubSetL (golint)
const ReqHubSetL = "service.set"

// ReqHubGet (golint)
const ReqHubGet = "hub.get"

// ReqHubGetL (golint)
const ReqHubGetL = "service.get"

// ReqHubStatus (golint)
const ReqHubStatus = "hub.status"

// ReqHubStatusL (golint)
const ReqHubStatusL = "service.status"

// ReqHubSignal (golint)
const ReqHubSignal = "hub.signal"

// ReqHubSignalL (golint)
const ReqHubSignalL = "service.signal"

// ReqHubSyncStatus (golint)
const ReqHubSyncStatus = "hub.sync.status"

// ReqHubSyncStatusL (golint)
const ReqHubSyncStatusL = "service.sync.status"

// ReqDFUHubGet (golint)
const ReqDFUHubGet = "hub.dfu.get"

// ReqDFUHubGetL (golint)
const ReqDFUHubGetL = "dfu.service.get"

// Request is the core API request/response data structure
type Request struct {
	Req              string                        `json:"req,omitempty"`
	Cmd              string                        `json:"cmd,omitempty"`
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
	APN              string                        `json:"apn,omitempty"`
	Text             string                        `json:"text,omitempty"`
	Base             int32                         `json:"base,omitempty"`
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
	Allow            bool                          `json:"allow,omitempty"`
	Align            bool                          `json:"align,omitempty"`
	Limit            bool                          `json:"limit,omitempty"`
	Pending          bool                          `json:"pending,omitempty"`
	Charging         bool                          `json:"charging,omitempty"`
	On               bool                          `json:"on,omitempty"`
	Off              bool                          `json:"off,omitempty"`
	ReqTime          bool                          `json:"reqtime,omitempty"`
	ReqLoc           bool                          `json:"reqloc,omitempty"`
	Trace            string                        `json:"trace,omitempty"`
	Usage            *[]string                     `json:"usage,omitempty"`
	State            *[]PinState                   `json:"state,omitempty"`
	Time             int64                         `json:"time,omitempty"`
	Motion           int64                         `json:"motion,omitempty"`
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
	Now              bool                          `json:"now,omitempty"`
	Type             int32                         `json:"type,omitempty"`
	Number           int64                         `json:"number,omitempty"`
	SKU              string                        `json:"sku,omitempty"`
	Board            string                        `json:"board,omitempty"`
	Net              *NetInfo                      `json:"net,omitempty"`
	Sensitivity      int32                         `json:"sensitivity,omitempty"`
	Requested        int32                         `json:"requested,omitempty"`
	Completed        int32                         `json:"completed,omitempty"`
	WiFi             bool                          `json:"wifi,omitempty"`
	Cell             bool                          `json:"cell,omitempty"`
	GPS              bool                          `json:"gps,omitempty"`
	Inbound          int32                         `json:"inbound,omitempty"`
	InboundV         string                        `json:"vinbound,omitempty"`
	Outbound         int32                         `json:"outbound,omitempty"`
	OutboundV        string                        `json:"voutbound,omitempty"`
	Duration         int32                         `json:"duration,omitempty"`
	Temperature      float64                       `json:"temperature,omitempty"`
	Pressure         float64                       `json:"pressure,omitempty"`
	Humidity         float64                       `json:"humidity,omitempty"`
	API              uint32                        `json:"api,omitempty"`
	SSID             string                        `json:"ssid,omitempty"`
	Password         string                        `json:"password,omitempty"`
	Security         string                        `json:"security,omitempty"`
	Key              string                        `json:"key,omitempty"`
	Method           string                        `json:"method,omitempty"`
	Content          string                        `json:"content,omitempty"`
	Min              int32                         `json:"min,omitempty"`
	Add              int32                         `json:"add,omitempty"`
	Encrypt          bool                          `json:"encrypt,omitempty"`
	Decrypt          bool                          `json:"decrypt,omitempty"`
	Alt              bool                          `json:"alt,omitempty"`
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
