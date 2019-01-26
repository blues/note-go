// Copyright 2017 Inca Roads LLC.  All rights reserved. 
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.
// Forked from github.com/d2r2/go-i2c
// Forked from github.com/davecheney/i2c

// +build !windows

// Before usage you must load the i2c-dev kernel module.
// Each i2c bus can address 127 independent i2c devices, and most
// linux systems contain several buses.

package notecard

import (
	"os"
	"fmt"
	"syscall"
)

const (
	// I2CSlave is the slave device address
	I2CSlave = 0x0703
)

// I2C is the handle to the I2C subsystem
type I2C struct {
	rc *os.File
}

// The open I2C port
var openI2CPort *I2C

// Get the default i2c device
func i2cDefault() (port string, portConfig int) {
	// The I2C port on Raspberry Pi is i2c-1 because it is "bus 1" as in I2C1
	port = "/dev/i2c-1"
	portConfig = 0x17
	return
}

// Open the i2c port
func i2cOpen(addr uint8, port string) (error) {
	f, err := os.OpenFile(port, os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("error on os.OpenFile: %s", err)
	}
	if err = ioctl(f.Fd(), I2CSlave, uintptr(addr)); err != nil {
		return fmt.Errorf("error on ioctl in i2cOpen: %s", err)
	}
	openI2CPort = &I2C{rc: f}
	return nil
}

// WriteBytes writes a buffer to I2C and return how many written
func i2cWriteBytes(buf []byte) (n int, err error) {
	return openI2CPort.write(buf)
}

// WriteByte writes a single byte to I2C
func i2cWriteByte(b byte) (n int, err error) {
 	return openI2CPort.write([]byte{b})
}

// ReadBytes reads a buffer from I2C and return how many written
func i2cReadBytes(buf []byte) (n int, err error) {
	return openI2CPort.read(buf)
}

// Close I2C
func i2cClose() error {
	return openI2CPort.rc.Close()
}

// Write a buffer to I2C
func (v *I2C) write(buf []byte) (n int, err error) {
	n, err = v.rc.Write(buf)
	if err != nil {
		err = fmt.Errorf("i2c write error: %s", err)
	}
	return
}

// Read a buffer from I2C
func (v *I2C) read(buf []byte) (n int, err error) {
	n, err = v.rc.Read(buf)
	if err != nil {
		err = fmt.Errorf("i2c read error: %s", err)
	}
	return
}

// Lowest level IO
func ioctl(fd, cmd, arg uintptr) (err error) {
	_, _, errnum := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
	if errnum != 0 {
		err = fmt.Errorf("i2c syscall error: %d", errnum)
	}
	return nil
}

// Enum I2C ports
func i2cPortEnum() (names []string) {
	return
}
