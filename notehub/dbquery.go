// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notehub

// DbQuery is the structure for a database query
type DbQuery struct {
	Columns    string `json:"columns,omitempty"`
	Format     string `json:"format,omitempty"`
	Count      bool   `json:"count,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	NoHeader   bool   `json:"noheader,omitempty"`
	Where      string `json:"where,omitempty"`
	Last       string `json:"last,omitempty"`
	Order      string `json:"order,omitempty"`
	Descending bool   `json:"descending,omitempty"`
}
