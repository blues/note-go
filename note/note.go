// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

// Note is the most fundamental data structure, containing
// user data referred to as its "body" and its "payload".  All
// access to these fields, and changes to these fields, must
// be done indirectly through the note API.
type Note struct {
    Body interface{}            `json:"b,omitempty"`
    Payload []byte              `json:"p,omitempty"`
    Change int64                `json:"c,omitempty"`
    Histories *[]NoteHistory    `json:"h,omitempty"`
    Conflicts *[]Note           `json:"x,omitempty"`
    Updates int32               `json:"u,omitempty"`
    Deleted bool                `json:"d,omitempty"`
    Sent bool                   `json:"s,omitempty"`
    Bulk bool                   `json:"k,omitempty"`
}

// NoteHistory records the update history, optimized so that if the most recent entry
// is by the same endpoint as an update/delete, that entry is re-used.  The primary use
// of NoteHistory is for conflict detection, and you don't need to detect conflicts
// against yourself.
type NoteHistory struct {
    When int64                  `json:"w,omitempty"`
    Where string				`json:"l,omitempty"`
    EndpointID string           `json:"e,omitempty"`
    Sequence int32              `json:"s,omitempty"`
}

// NoteInfo is the info returned on a per-note basis on requests
type NoteInfo struct {
    Body *interface{}           `json:"body,omitempty"`
    Payload *[]byte             `json:"payload,omitempty"`
    Deleted bool                `json:"deleted,omitempty"`
}
