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
var hostInitialized bool
var openI2CPort *I2C

// Get the default i2c device
func i2cDefault() (port string, portConfig int) {
	port = "" // Null string opens first available bus
	portConfig = 0x17
	return
}

// Open the i2c port
func i2cOpen(port string, portConfig int) (err error) {

	// Open the periph.io host
	if !hostInitialized {
		openI2CPort = &I2C{}
		openI2CPort.host, err = host.Init()
		if err != nil {
			return
		}
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

// Set the port config of the open port
func i2cSetConfig(portConfig int) (err error) {
	openI2CPort.device = &i2c.Dev{Bus: openI2CPort.bus, Addr: uint16(portConfig)}
	return
}

// WriteBytes writes a buffer to I2C
func i2cWriteBytes(buf []byte) (err error) {
	time.Sleep(1 * time.Millisecond) // By design, must not send more than once every 1Ms
	reg := make([]byte, 1)
	reg[0] = byte(len(buf))
	reg = append(reg, buf...)
	err = openI2CPort.device.Tx(reg, nil)
	if err != nil {
		err = fmt.Errorf("wb: %s", err)
	}
	return
}

// ReadBytes reads a buffer from I2C and returns how many are still pending
func i2cReadBytes(datalen int) (outbuf []byte, available int, err error) {
	time.Sleep(1 * time.Millisecond) // By design, must not send more than once every 1Ms
	readbuf := make([]byte, datalen+2)
	for i := 0; ; i++ { // Retry just for robustness
		reg := make([]byte, 2)
		reg[0] = byte(0)
		reg[1] = byte(datalen)
		err = openI2CPort.device.Tx(reg, readbuf)
		if err == nil {
			break
		}
		if i >= 10 {
			err = fmt.Errorf("rb: %s", err)
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	if len(readbuf) < 2 {
		err = fmt.Errorf("rb: not enough data (%d < 2)", len(readbuf))
		return
	}
	available = int(readbuf[0])
	if available > 253 {
		err = fmt.Errorf("rb: available too large (%d >253)", available)
		return
	}
	good := readbuf[1]
	if len(readbuf) < int(2+good) {
		err = fmt.Errorf("rb: insufficient data (%d < %d)", len(readbuf), 2+good)
		return
	}
	if 2 > 2+good {
		if false {
			fmt.Printf("i2cReadBytes(%d): %v\n", datalen, readbuf)
		}
		err = fmt.Errorf("rb: %d bytes returned while expecting %d", good, datalen)
		return
	}
	outbuf = readbuf[2 : 2+good]
	return
}

// Close I2C
func i2cClose() error {
	return openI2CPort.bus.Close()
}

// Enum I2C ports
func i2cPortEnum() (allports []string, usbports []string, notecardports []string, err error) {

	// Open the periph.io host
	if !hostInitialized {
		openI2CPort = &I2C{}
		openI2CPort.host, err = host.Init()
		if err != nil {
			return
		}
	}

	// Enum
	for _, ref := range i2creg.All() {
		port := ref.Name
		if ref.Number != -1 {
			allports = append(allports, port)
			notecardports = append(notecardports, port)
		}
	}
	return
}
