// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package api

// GetAppResponse v1
//
// The response object for getting an app.
type GetAppResponse struct {
	UID   string `json:"uid"`
	Label string `json:"label"`
	// RFC3339 timestamp, in UTC.
	Created string `json:"created"`

	AdministrativeContact *ContactResponse `json:"administrative_contact"`
	TechnicalContact      *ContactResponse `json:"technical_contact"`

	// "owner", "developer", or "viewer"
	Role *string `json:"role"`
}

// ContactResponse v1
//
// The response object for an app contact.
type ContactResponse struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	Organization string `json:"organization"`
}

// GenerateClientAppResponse v1
//
// The response object for generating a new client app for
// a specific app
type GenerateClientAppResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}
