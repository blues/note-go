// Copyright 2017 Inca Roads LLC.  All rights reserved. 
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

// +build !windows

package notecard

import (
	"fmt"
	"strings"
    "io/ioutil"
)

// Get the default serial device.  This is a symlink that maps to /dev/ttyS0 (the serial port on RPi)
// and also to /dev/ttyAMA0 (the serial port on the RPi CM3)
func serialDefault() (port string, portConfig int) {
	port = "/dev/serial0"
	portConfig = 115200
	return
}

// Set or display the serial port
func serialPortEnum() (names []string) {
    files, err := ioutil.ReadDir("/dev/")
    if err != nil {
        fmt.Printf("%s\n", err)
        return
    }
    for _, f := range files {
        name := f.Name()
        if strings.HasPrefix(name, "tty.usb") {
            names = append(names, "/dev/"+name)
        }
    }
    return
}
