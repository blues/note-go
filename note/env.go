// Copyright 2026 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package note

import (
	"strconv"
	"strings"
)

// EnvVarGPSFuzzDegrees is a user-defined environment variable used to obscure
// reported location. Its value is a floating point number in decimal degrees.
const EnvVarGPSFuzzDegrees = "_gps_fuzz_degrees"

// EnvVarSuppressGeolocation is a user-defined environment variable used to
// prevent the notehub from performing geolookups of any kind, when a customer
// determines that performing such a lookup might impact privacy within
// the context of their application.
const EnvVarSuppressGeolocation = "_suppress_geolocation"

// EnvVarSuppressTowerGeolocation does the same but just for cell tower
const EnvVarSuppressTowerGeolocation = "_suppress_tower_geolocation"

// EnvVarSuppressTriGeolocation does the same but just for triangulation
const EnvVarSuppressTriGeolocation = "_suppress_tri_geolocation"

// Interpret an env var as a bool using exactly the same alg as Notecard envGetBool()
func EnvGetBool(env map[string]string, key string) bool {
	if env == nil {
		return false
	}
	value, present := env[key]
	if !present {
		return false
	}
	lcValue := strings.ToLower(value)
	if strings.HasPrefix(lcValue, "t") {
		return true
	}
	if strings.HasPrefix(lcValue, "f") {
		return false
	}
	if n, err := strconv.Atoi(value); err == nil && n != 0 {
		return true
	}
	return false
}
