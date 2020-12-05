// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

// CardTest is a structure that is returned by the notecard after completing its self-test
type CardTest struct {
	DeviceUID  string `json:"device,omitempty"`
	Error      string `json:"err,omitempty"`
	Status     string `json:"status,omitempty"`
	Tests      string `json:"tests,omitempty"`
	FailTest   string `json:"fail_test,omitempty"`
	FailReason string `json:"fail_reason,omitempty"`
	Info       string `json:"info,omitempty"`
	Board      uint32 `json:"board,omitempty"`
	Modem      string `json:"modem,omitempty"`
	ICCID      string `json:"iccid,omitempty"`
	IMSI       string `json:"imsi,omitempty"`
	IMEI       string `json:"imei,omitempty"`
	When       int64  `json:"when,omitempty"`
	SKU        string `json:"sku,omitempty"`
	Station    string `json:"station,omitempty"`
	Operator   string `json:"operator,omitempty"`
	Check      uint32 `json:"check,omitempty"`
	Key        string `json:"key,omitempty"`
	// Firmware info
	FirmwareOrg     string `json:"org,omitempty"`
	FirmwareProduct string `json:"product,omitempty"`
	FirmwareVersion string `json:"version,omitempty"`
	FirmwareMajor   uint32 `json:"ver_major,omitempty"`
	FirmwareMinor   uint32 `json:"ver_minor,omitempty"`
	FirmwarePatch   uint32 `json:"ver_patch,omitempty"`
	FirmwareBuild   uint32 `json:"ver_build,omitempty"`
	FirmwareBuilt   string `json:"built,omitempty"`
	// Certificate info
	CertSN string `json:"certsn,omitempty"`
	Cert   string `json:"cert,omitempty"`
	// Setup requests
	SetupRequests string `json:"setup,omitempty"`
}
