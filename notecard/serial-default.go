// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"go.bug.st/serial/enumerator"
	"strings"
)

// Notecard's USB VID/PID
const notecardVID = "30A4"
const notecardPID = "0001"

// Get the default serial device
func defaultSerialDefault() (device string, speed int) {
	speed = 115200
	ports, err2 := enumerator.GetDetailedPortsList()
	if err2 != nil {
		return
	}
	if len(ports) == 0 {
		return
	}
	for _, port := range ports {
		if port.IsUSB {
			if strings.EqualFold(port.VID, notecardVID) && strings.EqualFold(port.PID, notecardPID) {
				device = port.Name
				return
			}
		}
	}
	return
}

// Set or display the serial port
func defaultSerialPortEnum() (allports []string, usbports []string, notecardports []string, err error) {
	ports, err2 := enumerator.GetDetailedPortsList()
	if err2 != nil {
		err = err2
		return
	}
	if len(ports) == 0 {
		return
	}
	for _, port := range ports {
		allports = append(allports, port.Name)
		if port.IsUSB {
			usbports = append(usbports, port.Name)
			if strings.EqualFold(port.VID, notecardVID) && strings.EqualFold(port.PID, notecardPID) {
				notecardports = append(notecardports, port.Name)
			}
		}
	}
	return
}
