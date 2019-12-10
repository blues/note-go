// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.
// Forked from github.com/d2r2/go-i2c
// Forked from github.com/davecheney/i2c

// +build !windows

// Before usage you must load the i2c-dev kernel module.
// Each i2c bus can address 127 independent i2c devices, and most
// linux systems contain several buses.

// Note: I2C Device Interface is accessed through periph.io library
// Example: https://github.com/google/periph/blob/master/devices/bmxx80/bmx280.go

package notecard

import (
	"fmt"
	"periph.io/x/periph"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
	"time"
)

const (
	// I2CSlave is the slave device address
	I2CSlave = 0x0703
)

// I2C is the handle to the I2C subsystem
type I2C struct {
	host   *periph.State
	bus    i2c.BusCloser
	device *i2c.Dev
}

// The open I2C port
var openI2CPort *I2C

// Get the default i2c device
func i2cDefault() (port string, portConfig int) {
	port = "" // Null string opens first available bus
	portConfig = 0x17
	return
}

// Open the i2c port
func i2cOpen(addr uint8, port string, portConfig int) (err error) {

	// Open the periph.io host
	openI2CPort = &I2C{}
	openI2CPort.host, err = host.Init()
	if err != nil {
		return
	}

	// Open the I2C instance
	openI2CPort.bus, err = i2creg.Open(port)
	if err != nil {
		return
	}

	// Instantiate the device
	openI2CPort.device = &i2c.Dev{Bus: openI2CPort.bus, Addr: uint16(portConfig)}

	return nil
}

// WriteBytes writes a buffer to I2C
func i2cWriteBytes(buf []byte) (err error) {
	time.Sleep(1 * time.Millisecond) // By design, must not send more than once every 1Ms
	reg := make([]byte, 1)
	reg[0] = byte(len(buf))
	reg = append(reg, buf...)
	return openI2CPort.device.Tx(reg, nil)
}

// ReadBytes reads a buffer from I2C and returns how many are still pending
func i2cReadBytes(datalen int) (outbuf []byte, available int, err error) {
	time.Sleep(1 * time.Millisecond) // By design, must not send more than once every 1Ms
	readbuf := make([]byte, datalen+2)
	reg := make([]byte, 2)
	reg[0] = byte(0)
	reg[1] = byte(datalen)
	err = openI2CPort.device.Tx(reg, readbuf)
	if err != nil {
		return
	}
	available = int(readbuf[0])
	good := readbuf[1]
	_ = good
	outbuf = readbuf[2 : 2+good]
	return
}

// Close I2C
func i2cClose() error {
	return openI2CPort.bus.Close()
}

// Enum I2C ports
func i2cPortEnum(knownNotecardsOnly bool) (names []string, err error) {
	for _, ref := range i2creg.All() {
		port := ref.Name
		if ref.Number != -1 {
			port = fmt.Sprintf("%s", ref.Name)
		}
		names = append(names, port)
	}
	return
}
