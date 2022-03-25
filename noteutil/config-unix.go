// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

//go:build !windows
// +build !windows

// Before usage you must load the i2c-dev kernel module.
// Each i2c bus can address 127 independent i2c devices, and most
// linux systems contain several buses.

package noteutil

import (
	"os"
	"os/user"
)

// ConfigDir returns the config directory
func ConfigDir() string {
	usr, err := user.Current()
	if err != nil {
		return "."
	}
	path := usr.HomeDir + "/note"
	os.MkdirAll(path, 0777)
	return path
}

// Get the pathname of config settings
func configSettingsPath() string {
	return ConfigDir() + "/config.json"
}
