// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

const (
	NetworkBearerUnknown      = -1
	NetworkBearerGsm          = 0
	NetworkBearerTdScdma      = 1
	NetworkBearerWcdma        = 2
	NetworkBearerCdma2000     = 3
	NetworkBearerWiMax        = 4
	NetworkBearerLteTdd       = 5
	NetworkBearerLteFdd       = 6
	NetworkBearerNBIot        = 7
	NetworkBearerWLan         = 21
	NetworkBearerBluetooth    = 22
	NetworkBearerIeee802p15p4 = 23
	NetworkBearerEthernet     = 41
	NetworkBearerDsl          = 42
	NetworkBearerPlc          = 43
)

// NetInfo is the composite structure with all networking connection info
type NetInfo struct {
	Iccid                    string `json:"iccid,omitempty"`
	Iccid2                   string `json:"iccid2,omitempty"`
	IccidExternal            string `json:"iccid_external,omitempty"`
	Imsi                     string `json:"imsi,omitempty"`
	Imsi2                    string `json:"imsi2,omitempty"`
	ImsiExternal             string `json:"imsi_external,omitempty"`
	Imei                     string `json:"imei,omitempty"`
	ModemFirmware            string `json:"modem,omitempty"`
	Band                     string `json:"band,omitempty"`
	AccessTechnology         string `json:"rat,omitempty"`
	AccessTechnologyFilter   string `json:"ratf,omitempty"`
	ReportedAccessTechnology string `json:"ratr,omitempty"`
	ReportedCarrier          string `json:"carrier,omitempty"`
	Bssid                    string `json:"bssid,omitempty"`
	Ssid                     string `json:"ssid,omitempty"`
	// Internal vs external SIM used at any given moment
	InternalSIMSelected bool `json:"internal,omitempty"`
	// Radio signal strength in dBm, or ModemValueUnknown if it is not
	// available from the modem.
	RssiRange int32 `json:"rssir,omitempty"`
	// GSM RxQual, or ModemValueUnknown if it is not available from the modem.
	Rxqual int32 `json:"rxqual,omitempty"`
	// General received signal strength, in dBm
	Rssi int32 `json:"rssi,omitempty"`
	// An integer indicating the reference signal received power (RSRP)
	Rsrp int32 `json:"rsrp,omitempty"`
	// An integer indicating the signal to interference plus noise ratio (SINR).
	// Logarithmic value of SINR. Values are in 1/5th of a dB. The range is 0-250
	// which translates to -20dB - +30dB
	Sinr int32 `json:"sinr,omitempty"`
	// An integer indicating the reference signal received quality (RSRQ)
	Rsrq int32 `json:"rsrq,omitempty"`
	// An integer indicating relative signal strength in a human-readable way
	Bars uint32 `json:"bars,omitempty"`
	// IP address assigned to the device
	IP string `json:"ip,omitempty"`
	// IP address that the device is talking to (if known)
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
	// Modem debug
	ModemDebugEvents int32 `json:"modem_test_events,omitempty"`
	// Overcurrent events
	OvercurrentEvents    int32 `json:"oc_events,omitempty"`
	OvercurrentEventSecs int32 `json:"oc_event_time,omitempty"`
	// When the signal strength fields were last updated
	Modified int64 `json:"updated,omitempty"`
}
