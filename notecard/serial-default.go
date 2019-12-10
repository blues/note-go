// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.


package notecard

import (
	"go.bug.st/serial/enumerator"
)

// Notecard's USB VID/PID
const notecardVID	= "30A4"
const notecardPID	= "0001"

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
			if port.VID == notecardVID && port.PID == notecardPID {
				device = port.Name
				return
			}
		}
	}
	return
}

// Set or display the serial port
func defaultSerialPortEnum(knownNotecardsOnly bool) (names []string, err error) {
	ports, err2 := enumerator.GetDetailedPortsList()
	if err2 != nil {
		err = err2
		return
	}
	if len(ports) == 0 {
		return
	}
	for _, port := range ports {
		if !knownNotecardsOnly {
			names = append(names, port.Name)
		} else {
			if port.IsUSB && port.VID == notecardVID && port.PID == notecardPID {
				names = append(names, port.Name)
			}
		}
	}
	return
}
