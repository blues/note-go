// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

//go:build windows

package notecard

import (
	"fmt"
)

// Get the default i2c device
func i2cDefault() (port string, portConfig int) {
	port = "???"
	portConfig = 0x17
	return
}

// Set the port config of the open port
func i2cSetConfig(portConfig int) (err error) {
	return fmt.Errorf("i2c not yet implemented")
}

// Open the i2c port
func i2cOpen(port string, portConfig int) (err error) {
	return fmt.Errorf("i2c not yet implemented")
}

// WriteBytes writes a buffer to I2C
func i2cWriteBytes(buf []byte, i2cAddr int) (err error) {
	return fmt.Errorf("i2c not yet implemented")
}

// ReadBytes reads a buffer from I2C and returns how many are still pending
func i2cReadBytes(datalen int, i2cAddr int) (outbuf []byte, available int, err error) {
	err = fmt.Errorf("i2c not yet implemented")
	return
}

// Close I2C
func i2cClose() error {
	return fmt.Errorf("i2c not yet implemented")
}

// Enum I2C ports
func i2cPortEnum() (allports []string, usbports []string, notecardports []string, err error) {
	return
}
