// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

// NetInfo is the composite structure with all networking connection info
type NetInfo struct {
	Iccid            string `json:"iccid,omitempty"`
	IccidExternal    string `json:"iccid_external,omitempty"`
	Imsi             string `json:"imsi,omitempty"`
	ImsiExternal     string `json:"imsi_external,omitempty"`
	Imei             string `json:"imei,omitempty"`
	ModemFirmware    string `json:"modem,omitempty"`
	Band             string `json:"band,omitempty"`
	AccessTechnology string `json:"rat,omitempty"`
	// Radio signal strength in dBm, or ModemValueUnknown if it is not
	// available from the modem.
	RssiRange int32 `json:"rssir,omitempty"`
	Rssi      int32 `json:"rssi,omitempty"`
	// An integer indicating the reference signal received power (RSRP)
	Rsrp int32 `json:"rsrp,omitempty"`
	// An integer indicating the reference signal received quality (RSRQ)
	Rsrq int32 `json:"rsrq,omitempty"`
	// An integer indicating relative signal strength in a human-readable way
	Bars uint32 `json:"bars,omitempty"`
	// An integer indicating the signal to interference plus noise ratio (SINR).
	// Logarithmic value of SINR. Values are in 1/5th of a dB. The range is 0-250
	// which translates to -20dB - +30dB
	Sinr int32 `json:"sinr,omitempty"`
	// GSM RxQual, or ModemValueUnknown if it is not available from the modem.
	Rxqual int32 `json:"rxqual,omitempty"`
	// Device IP address
	IP string `json:"ip,omitempty"`
	// Device GW address
	Gw string `json:"gateway,omitempty"`
	// Device APN name
	Apn string `json:"apn,omitempty"`
	// Location area code (16 bits) or ModemValueUnknown if it is not avail from modem
	Lac uint32 `json:"lac,omitempty"`
	// Cell ID (28 bits) or ModemValueUnknown if it is not available from the modem.
	Cellid uint32 `json:"cid,omitempty"`
	// Network info
	NetworkBearer int32  `json:"bearer,omitempty"`
	Mcc           uint32 `json:"mcc,omitempty"`
	Mnc           uint32 `json:"mnc,omitempty"`
	// Overcurrent events
	OvercurrentEvents    int32 `json:"oc_events,omitempty"`
	OvercurrentEventSecs int32 `json:"oc_event_time,omitempty"`
	// Modem debug
	ModemDebugEvents int32 `json:"modem_test_events,omitempty"`
	// When the signal strength fields were last updated
	Modified int64 `json:"updated,omitempty"`
}
