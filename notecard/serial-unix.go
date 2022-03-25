// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

//go:build !windows
// +build !windows

package notecard

// Get the default serial device
func serialDefault() (device string, speed int) {
	return defaultSerialDefault()
}

// Set or display the serial port
func serialPortEnum() (allports []string, usbports []string, notecardports []string, err error) {
	return defaultSerialPortEnum()
}
