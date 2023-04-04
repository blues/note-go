// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"strings"

	"go.bug.st/serial/enumerator"
)

// Notecard's USB VID/PID
const (
	bluesincVID = "30A4"
	notecardPID = "0001"
)

// Get the default serial device
func defaultSerialDefault() (device string, speed int) {
	// Enum all ports
	speed = 115200
	ports, err2 := enumerator.GetDetailedPortsList()
	if err2 != nil {
		return
	}
	if len(ports) == 0 {
		return
	}

	// First, look for the notecard
	for _, port := range ports {
		if port.IsUSB {
			if strings.EqualFold(port.VID, bluesincVID) && strings.EqualFold(port.PID, notecardPID) {
				device = port.Name
				return
			}
		}
	}

	// Otherwise, look for anything from Blues
	for _, port := range ports {
		if port.IsUSB && strings.EqualFold(port.VID, bluesincVID) {
			device = port.Name
			return
		}
	}

	// Not found
	return
}

// Set or display the serial port
func defaultSerialPortEnum() (allports []string, usbports []string, notecardports []string, err error) {
	// Enum all ports
	ports, err2 := enumerator.GetDetailedPortsList()
	if err2 != nil {
		err = err2
		return
	}
	if len(ports) == 0 {
		return
	}

	// First, look for the notecard
	for _, port := range ports {
		allports = append(allports, port.Name)
		if port.IsUSB {
			usbports = append(usbports, port.Name)
			if strings.EqualFold(port.VID, bluesincVID) && strings.EqualFold(port.PID, notecardPID) {
				notecardports = append(notecardports, port.Name)
			}
		}
	}

	// Otherwise, look for anything from Blues
	if len(notecardports) == 0 {
		for _, port := range ports {
			if port.IsUSB && strings.EqualFold(port.VID, bluesincVID) {
				notecardports = append(notecardports, port.Name)
			}
		}
	}

	// Done
	return
}
