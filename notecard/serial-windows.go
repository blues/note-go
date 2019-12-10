// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

// +build windows

// If you have odd serial port behavior (where responses are apparently lost or delayed), try this:
// 1) open Control Panel -> Device Manager -> Ports (COM & LPT)
// 2) right-click for USB Serial Device Properties on the appropriate port
// 3) Port Settings tab
// 4) Click Advanced... button
// 5) UN-CHECK "Use FIFO buffers"

package notecard

import ()

// Get the default serial device
func serialDefault() (device string, speed int) {
	return defaultSerialDefault()
}

// Set or display the serial port
func serialPortEnum(knownNotecardsOnly bool) (names []string, err error) {
	return defaultSerialPortEnum(knownNotecardsOnly bool)
}
