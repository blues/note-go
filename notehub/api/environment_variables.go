// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package api

// GetAppEnvironmentVariablesResponse v1
//
// The response object for getting app environment variables.
type GetAppEnvironmentVariablesResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this app.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// PutAppEnvironmentVariablesRequest v1
//
// The request object for setting app environment variables.
type PutAppEnvironmentVariablesRequest struct {
	// EnvironmentVariables
	//
	// The environment variables scoped at the app level
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// PutAppEnvironmentVariablesResponse v1
//
// The response object for setting app environment variables.
type PutAppEnvironmentVariablesResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this app.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// DeleteAppEnvironmentVariableResponse v1
//
// The response object for deleting an app environment variable.
type DeleteAppEnvironmentVariableResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this app.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// GetFleetEnvironmentVariablesResponse v1
//
// The response object for getting fleet environment variables.
type GetFleetEnvironmentVariablesResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this fleet.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// PutFleetEnvironmentVariablesRequest v1
//
// The request object for setting fleet environment variables.
type PutFleetEnvironmentVariablesRequest struct {
	// EnvironmentVariables
	//
	// The environment variables scoped at the fleet level
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// PutFleetEnvironmentVariablesResponse v1
//
// The response object for setting fleet environment variables.
type PutFleetEnvironmentVariablesResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this fleet.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// DeleteFleetEnvironmentVariableResponse v1
//
// The response object for deleting an fleet environment variable.
type DeleteFleetEnvironmentVariableResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this fleet.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// GetDeviceEnvironmentVariablesResponse v1
//
// The response object for getting device environment variables.
type GetDeviceEnvironmentVariablesResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this device that have been set using host firmware or the Notehub API or UI.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`

	// EnvironmentVariablesEnvDefault
	//
	// The environment variables that have been set using the env.default request through the Notecard API.
	//
	// required: true
	EnvironmentVariablesEnvDefault map[string]string `json:"environment_variables_env_default"`
}

// PutDeviceEnvironmentVariablesRequest v1
//
// The request object for setting device environment variables.
type PutDeviceEnvironmentVariablesRequest struct {
	// EnvironmentVariables
	//
	// The environment variables scoped at the device level
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// PutDeviceEnvironmentVariablesResponse v1
//
// The response object for setting device environment variables.
type PutDeviceEnvironmentVariablesResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this device that have been set using host firmware or the Notehub API or UI.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// DeleteDeviceEnvironmentVariableResponse v1
//
// The response object for deleting a device environment variable.
type DeleteDeviceEnvironmentVariableResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this device that have been set using host firmware or the Notehub API or UI.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// GetDeviceEnvironmentVariablesWithPINResponse v1
//
// The response object for getting device environment variables with a PIN.
type GetDeviceEnvironmentVariablesWithPINResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this device that have been set using host firmware or the Notehub API or UI.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`

	// EnvironmentVariablesEnvDefault
	//
	// The environment variables that have been set using the env.default request through the Notecard API.
	//
	// required: true
	EnvironmentVariablesEnvDefault map[string]string `json:"environment_variables_env_default"`
}

// PutDeviceEnvironmentVariablesWithPINRequest v1
//
// The request object for setting device environment variables with a PIN. (The PIN comes in via a header)
type PutDeviceEnvironmentVariablesWithPINRequest struct {
	// EnvironmentVariables
	//
	// The environment variables scoped at the device level
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

// PutDeviceEnvironmentVariablesWithPINResponse v1
//
// The response object for setting device environment variables with a PIN.
type PutDeviceEnvironmentVariablesWithPINResponse struct {
	// EnvironmentVariables
	//
	// The environment variables for this device that have been set using host firmware or the Notehub API or UI.
	//
	// required: true
	EnvironmentVariables map[string]string `json:"environment_variables"`
}
