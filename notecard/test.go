// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

// CardTest is a structure that is returned by the notecard after completing its self-test
type CardTest struct {
	DeviceUID           string `json:"device,omitempty"`
	Error               string `json:"err,omitempty"`
	Status              string `json:"status,omitempty"`
	Tests               string `json:"tests,omitempty"`
	FailTest            string `json:"fail_test,omitempty"`
	FailReason          string `json:"fail_reason,omitempty"`
	Info                string `json:"info,omitempty"`
	BoardVersion        uint32 `json:"board,omitempty"`
	BoardType           uint32 `json:"board_type,omitempty"`
	Modem               string `json:"modem,omitempty"`
	ICCID               string `json:"iccid,omitempty"`
	ICCID2              string `json:"iccid2,omitempty"`
	IMSI                string `json:"imsi,omitempty"`
	IMSI2               string `json:"imsi2,omitempty"`
	IMEI                string `json:"imei,omitempty"`
	Apn                 string `json:"apn,omitempty"`
	Band                string `json:"band,omitempty"`
	Channel             string `json:"channel,omitempty"`
	When                uint32 `json:"when,omitempty"`
	SKU                 string `json:"sku,omitempty"`
	OrderingCode        string `json:"ordering_code,omitempty"`
	DefaultProductUID   string `json:"default_product,omitempty"`
	SIMActivationKey    string `json:"key,omitempty"`
	SIMless             bool   `json:"simless,omitempty"`
	Station             string `json:"station,omitempty"`
	Operator            string `json:"operator,omitempty"`
	Check               uint32 `json:"check,omitempty"`
	CellUsageBytes      uint32 `json:"cell_used,omitempty"`
	CellProvisionedTime uint32 `json:"cell_provisioned,omitempty"`
	// Firmware info
	FirmwareOrg     string `json:"org,omitempty"`
	FirmwareProduct string `json:"product,omitempty"`
	FirmwareVersion string `json:"version,omitempty"`
	FirmwareMajor   uint32 `json:"ver_major,omitempty"`
	FirmwareMinor   uint32 `json:"ver_minor,omitempty"`
	FirmwarePatch   uint32 `json:"ver_patch,omitempty"`
	FirmwareBuild   uint32 `json:"ver_build,omitempty"`
	FirmwareBuilt   string `json:"built,omitempty"`
	// Certificate and cert info
	CertSN string `json:"certsn,omitempty"`
	Cert   string `json:"cert,omitempty"`
	// Card initialization requests
	SetupRequests string `json:"setup,omitempty"`
	// Detailed information about LSE stability
	LSEStability string `json:"lse,omitempty"`
	// LoRa notecard provisioning info
	DevEui    string `json:"deveui,omitempty"`
	AppEui    string `json:"appeui,omitempty"`
	AppKey    string `json:"appkey,omitempty"`
	FreqPlan  string `json:"freqplan,omitempty"`
	LWVersion string `json:"lorawan,omitempty"`
	PHVersion string `json:"regional,omitempty"`
	// For manufacturing
	CPN string `json:"cpn,omitempty"`
	// For Iridium
	IriSku   string `json:"iri_sku,omitempty"`
	IriSn    string `json:"iri_sn,omitempty"`
	IriImei  string `json:"iri_imei,omitempty"`
	IriIccid string `json:"iri_iccid,omitempty"`
	// For Starnote
	Hardware string `json:"hardware,omitempty"`
	Mtu      uint16 `json:"mtu,omitempty"`
	DownMtu  uint16 `json:"down_mtu,omitempty"`
	UpMtu    uint16 `json:"up_mtu,omitempty"`
	Policy   string `json:"policy,omitempty"`
	Cid      uint32 `json:"cid,omitempty"`
}

// Remove fields that are not useful or are sensitive when externalizing for public consumption
func CardTestExternalized(ct CardTest) CardTest {
	ct.BoardVersion = 0      // distracting
	ct.BoardType = 0         // distracting
	ct.SIMActivationKey = "" // security
	ct.Station = ""          // privacy
	ct.Operator = ""         // privacy
	ct.Check = 0             // invalid after externalizing
	ct.FirmwareOrg = ""      // distracting
	ct.FirmwareProduct = ""  // distracting
	ct.FirmwareMajor = 0     // distracting
	ct.FirmwareMinor = 0     // distracting
	ct.FirmwarePatch = 0     // distracting
	ct.FirmwareBuild = 0     // distracting
	ct.FirmwareBuilt = ""    // distracting
	ct.CertSN = ""           // security
	ct.Cert = ""             // security
	ct.LSEStability = ""     // distracting
	ct.SetupRequests = ""    // security
	ct.LSEStability = ""     // distracting
	return ct
}
