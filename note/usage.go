// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

// DeviceUsage is the device usage metric representing values from the beginning of time, since Provisioned
type DeviceUsage struct {
	Since        int64  `json:"since,omitempty"`
	DurationSecs uint32 `json:"duration,omitempty"`
	RcvdBytes    uint32 `json:"bytes_rcvd,omitempty"`
	SentBytes    uint32 `json:"bytes_sent,omitempty"`
	TCPSessions  uint32 `json:"sessions_tcp,omitempty"`
	TLSSessions  uint32 `json:"sessions_tls,omitempty"`
	RcvdNotes    uint32 `json:"notes_rcvd,omitempty"`
	SentNotes    uint32 `json:"notes_sent,omitempty"`
}
