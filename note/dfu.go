// Copyright 2025 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

// Package note dfu.go contains DFU-related structures generated/parsed by the notecard
package note

// DFUState is the state of the DFU in progress
type DFUState struct {
	Type               string `json:"type,omitempty"`
	File               string `json:"file,omitempty"`
	Length             uint32 `json:"length,omitempty"`
	CRC32              uint32 `json:"crc32,omitempty"`
	MD5                string `json:"md5,omitempty"`
	Phase              string `json:"mode,omitempty"`
	Status             string `json:"status,omitempty"`
	BeganSecs          uint32 `json:"began,omitempty"`
	RetryCount         uint32 `json:"retry,omitempty"`
	ConsecutiveErrors  uint32 `json:"errors,omitempty"`
	BinaryRetries      uint32 `json:"binretry,omitempty"`
	DFUStartCount      uint32 `json:"dfu_started,omitempty"`
	DFUCompletedCount  uint32 `json:"dfu_completed,omitempty"`
	ODFUStartedCount   uint32 `json:"odfu_started,omitempty"`
	ODFUTarget         string `json:"odfu_target,omitempty"`
	ReadFromService    uint32 `json:"read,omitempty"`
	UpdatedSecs        uint32 `json:"updated,omitempty"`
	DownloadComplete   bool   `json:"dl_complete,omitempty"`
	DisabledReason     string `json:"disabled,omitempty"`
	MinNotecardVersion string `json:"min_card_version,omitempty"`

	// This will always point to the current running version
	Version string `json:"version,omitempty"`
}

// DFUEnv is the data structure passed to Notehub when DFU info changes
type DFUEnv struct {
	Card  *DFUState `json:"card,omitempty"`
	User  *DFUState `json:"user,omitempty"`
	Modem *DFUState `json:"modem,omitempty"`
	Star  *DFUState `json:"star,omitempty"`
}

type DfuPhase string

const (
	DfuPhaseUnknown     DfuPhase = ""
	DfuPhaseIdle        DfuPhase = "idle"
	DfuPhaseError       DfuPhase = "error"
	DfuPhaseDownloading DfuPhase = "downloading"
	DfuPhaseSideloading DfuPhase = "sideloading"
	DfuPhaseReady       DfuPhase = "ready"
	DfuPhaseReadyRetry  DfuPhase = "ready-retry"
	DfuPhaseUpdating    DfuPhase = "updating"
	DfuPhaseCompleted   DfuPhase = "completed"
)
