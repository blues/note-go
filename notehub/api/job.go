// Copyright 2025 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package api

// This header is present for every type of job
type HubJob struct {
	Type      HubJobType `json:"type,omitempty"`
	Version   string     `json:"version,omitempty"`
	Name      string     `json:"name,omitempty"`
	Comment   string     `json:"comment,omitempty"`
	Created   int64      `json:"created,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
}

// This header is present for every type of report
type HubJobReport struct {
	Type        HubJobType `json:"type,omitempty"`
	Version     string     `json:"version,omitempty"`
	Comment     string     `json:"comment,omitempty"`
	JobId       string     `json:"job_id"`
	JobName     string     `json:"job_name"`
	Status      string     `json:"status,omitempty"`
	DryRun      bool       `json:"dry_run,omitempty"`
	Cancel      bool       `json:"cancel,omitempty"`
	SubmittedBy string     `json:"who_submitted,omitempty"`
	Submitted   int64      `json:"when_submitted,omitempty"`
	Started     int64      `json:"when_started,omitempty"`
	Updated     int64      `json:"when_updated,omitempty"`
	Completed   int64      `json:"when_completed,omitempty"`
}

// Types of jobs
type HubJobType string

const (
	HubJobTypeUnspecified    HubJobType = ""
	HubJobTypeReconciliation HubJobType = "reconciliation"
)

const (
	HubJobStatusCancelled = "cancelled"
	HubJobStatusSubmitted = "submitted"
)
