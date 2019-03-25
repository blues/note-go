// Copyright 2019 Blues Inc.  All rights reserved. 
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

// +build windows

package noteutil

import (
	"os/user"
	"github.com/blues/note-go/notecard"
)

// Get the default directory
func ConfigDir() string {
    usr, err := user.Current()
    if err != nil {
		return "."
    }
	path := usr.HomeDir + "\note"
    os.MkdirAll(path, 0777)
	return path
}

// Get the pathname of config settings
func configSettingsPath() string {
	return ConfigDir() + "\config.json"
}
