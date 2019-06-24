// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

// DeviceSession is the basic unit of recorded device usage history
type DeviceSession struct {
	// Session ID that can be mapped to the events created during that session
	SessionUID string					`json:"session,omitempty"`
	// Info from the device structure
    DeviceUID string                    `json:"device,omitempty"`
    DeviceSN string						`json:"sn,omitempty"`
    ProductUID string					`json:"product,omitempty"`
    FleetUID string						`json:"fleet,omitempty"`
	// IP address of the session
	Addr string							`json:"addr,omitempty"`
	// Cell ID where the session originated ("mcc,mnc,lac,cellid" all in base 10)
	CellID string						`json:"cell,omitempty"`
	// Last known tower location where device pinged
	Where TowerLocation					`json:"tower,omitempty"`
	// Total device usage at the beginning of the period
	This DeviceUsage					`json:"this,omitempty"`
	// Total device usage at the beginning of the next period, whenever it happens to occur
	Next DeviceUsage					`json:"next,omitempty"`
	// Usage during the period - initially estimated, but then corrected when we get to the next period
	Period DeviceUsage					`json:"period,omitempty"`
	// Physical device info
	Voltage DeviceUsage					`json:"voltage,omitempty"`
	Temp DeviceUsage					`json:"temp,omitempty"`
	// For keeping track of when the last work was done for a session
	LastWorkDone int64
}

// TowerLocation is the cell tower location structure generated by the tower utility
type TowerLocation struct {
    Name            string      `json:"n,omitempty"`		// name of the location
    CountryCode     string      `json:"c,omitempty"`		// country code
    TimeZoneID      int			`json:"z,omitempty"`		// timezone id (see tz.go)
    OLC				string		`json:"l,omitempty"`		// location
	Lat				float64		`json:"lat,omitempty"`		// latitude
	Lon				float64		`json:"lon,omitempty"`		// longitude
    TimeZone	    string		`json:"zone,omitempty"`		// timezone name
    MCC				int			`json:"mcc,omitempty"`
    MNC				int			`json:"mnc,omitempty"`
    LAC				int			`json:"lac,omitempty"`
    CID				int			`json:"cid,omitempty"`
}
