// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/blues/note-go/note"
	"go.bug.st/serial"
)

// Protect against multiple concurrent callers
var transLock sync.RWMutex

// Module communication interfaces
const (
	NotecardInterfaceSerial = "serial"
	NotecardInterfaceI2C    = "i2c"
	NotecardInterfaceRemote = "remote"
)

// CardI2CMax controls chunk size that's socially appropriate on the I2C bus.
// It must be 1-253 bytes as per spec (which allows space for the 2-byte header in a 255-byte read)
const CardI2CMax = 253

// The notecard is a real-time device that has a fixed size interrupt buffer.  We can push data
// at it far, far faster than it can process it, therefore we push it in segments with a pause
// between each segment.

// CardRequestSegmentMaxLen (golint)
const CardRequestSegmentMaxLen = 1000

// CardRequestSegmentDelayMs (golint)
const CardRequestSegmentDelayMs = 250

// IoErrorIsRecoverable is a configuration parameter describing library capabilities.
// Set this to true if the error recovery of the implementation supports re-open.  On all implementations
// tested to date, I can't yet get the close/reopen working the way it does on microcontrollers.  For
// example, on the go serial, I get a nil pointer dereference within the go library.  This MAY have
// soemthing to do with the fact that we don't cleanly implement the shutdown/restart of the inputHandler
// in trace, in which case that should be fixed.  In the meantime, this is disabled.
const IoErrorIsRecoverable = true

// Context for the port that is open
type Context struct {

	// True to emit trace output
	Debug bool

	// Pretty-print trace output JSON
	Pretty bool

	// Reset should be done on next transaction
	resetRequired  bool
	reopenRequired bool

	// Class functions
	PortEnumFn     func() (allports []string, usbports []string, notecardports []string, err error)
	PortDefaultsFn func() (port string, portConfig int)
	CloseFn        func(context *Context)
	ReopenFn       func(context *Context) (err error)
	ResetFn        func(context *Context) (err error)
	TransactionFn  func(context *Context, reqJSON []byte) (rspJSON []byte, err error)

	// Serial instance state
	isSerial       bool
	openSerialPort serial.Port
	serialName     string
	serialConfig   serial.Mode
	i2cName        string
	i2cAddress     int

	// Remote instance state
	isRemote            bool
	farmURL             string
	farmCheckoutMins    int
	farmCheckoutExpires int64
	farmCard            RemoteCard
}

// Report a critical card error
func cardReportError(context *Context, err error) {
	if context != nil && context.Debug {
		fmt.Printf("*** %s\n", err)
	}
	if IoErrorIsRecoverable {
		time.Sleep(500 * time.Millisecond)
		context.reopenRequired = true
	}
}

// DebugOutput enables/disables debug output
func (context *Context) DebugOutput(enabled bool, pretty bool) {
	context.Debug = enabled
	context.Pretty = pretty
}

// EnumPorts returns the list of all available ports on the specified interface
func (context *Context) EnumPorts() (allports []string, usbports []string, notecardports []string, err error) {
	return context.PortEnumFn()
}

// PortDefaults gets the defaults for the specified port
func (context *Context) PortDefaults() (port string, portConfig int) {
	return context.PortDefaultsFn()
}

// Identify this Notecard connection
func (context *Context) Identify() (protocol string, port string, portConfig int) {
	if context.isSerial {
		return "serial", context.serialName, context.serialConfig.BaudRate
	}
	return "I2C", context.i2cName, context.i2cAddress
}

// Defaults gets the default interface, port, and config
func Defaults() (moduleInterface string, port string, portConfig int) {
	moduleInterface = NotecardInterfaceSerial
	port, portConfig = serialDefault()
	return
}

// SetConfig sets port config on the open port
func (context *Context) SetConfig(portConfig int) (err error) {
	if !context.isSerial {
		err = i2cSetConfig(portConfig)
	}
	return
}

// Open the card to establish communications
func Open(moduleInterface string, port string, portConfig int) (context Context, err error) {

	if moduleInterface == "" {
		moduleInterface, _, _ = Defaults()
	}

	switch moduleInterface {
	case NotecardInterfaceSerial:
		context, err = OpenSerial(port, portConfig)
		break
	case NotecardInterfaceI2C:
		context, err = OpenI2C(port, portConfig)
		break
	case NotecardInterfaceRemote:
		context, err = OpenRemote(port, portConfig)
		break
	default:
		err = fmt.Errorf("unknown interface: %s", moduleInterface)
		break
	}
	if err != nil {
		cardReportError(nil, err)
		err = fmt.Errorf("error opening port: %s %s", err, note.ErrCardIo)
		return
	}

	return

}

// Reset serial to a known state
func cardResetSerial(context *Context) (err error) {

	// Exit if not open
	if context.openSerialPort == nil {
		err = fmt.Errorf("port not open " + note.ErrCardIo)
		cardReportError(context, err)
		return
	}

	// In order to ensure that we're not getting the reply to a failed
	// transaction from a prior session, drain any pending input prior
	// to transmitting a command.  Note that we use this technique of
	// looking for a known reply to \n, rather than just "draining
	// anything pending on serial", because the nature of read() is
	// that it blocks (until timeout) if there's nothing available.
	var length int
	buf := make([]byte, 2048)
	for {
		_, err = context.openSerialPort.Write([]byte("\n"))
		if err != nil {
			err = fmt.Errorf("error transmitting to module: %s %s", err, note.ErrCardIo)
			cardReportError(context, err)
			return
		}
		time.Sleep(750 * time.Millisecond)
		readBeganMs := int(time.Now().UnixNano() / 1000000)
		length, err = context.openSerialPort.Read(buf)
		readElapsedMs := int(time.Now().UnixNano()/1000000) - readBeganMs
		if readElapsedMs == 0 && length == 0 && err == io.EOF {
			// On Linux, hardware port failures come back simply as immediate EOF
			err = fmt.Errorf("hardware failure")
		}
		if err != nil {
			err = fmt.Errorf("error reading from module: %s %s", err, note.ErrCardIo)
			cardReportError(context, err)
			return
		}
		somethingFound := false
		nonCRLFFound := false
		for i := 0; i < length && !nonCRLFFound; i++ {
			if false {
				fmt.Printf("chr: 0x%02x '%c'\n", buf[i], buf[i])
			}
			somethingFound = true
			if buf[i] != '\r' && buf[i] != '\n' {
				nonCRLFFound = true
			}
		}
		if somethingFound && !nonCRLFFound {
			break
		}
	}

	// Done
	return

}

// OpenSerial opens the card on serial
func OpenSerial(port string, portConfig int) (context Context, err error) {

	// Use default if not specified
	if port == "" {
		port, portConfig = serialDefault()
	}

	// Set up class functions
	context.PortEnumFn = serialPortEnum
	context.PortDefaultsFn = serialDefault
	context.CloseFn = cardCloseSerial
	context.ReopenFn = cardReopenSerial
	context.ResetFn = cardResetSerial
	context.TransactionFn = cardTransactionSerial

	// Record serial configuration
	context.isSerial = true
	context.serialName = port
	context.serialConfig.BaudRate = portConfig

	// Open the serial port
	if true {
		context.reopenRequired = true
	} else {
		err = cardReopenSerial(&context)
		if err != nil {
			err = fmt.Errorf("error opening serial port %s at %d: %s %s", port, portConfig, err, note.ErrCardIo)
			return
		}
	}

	// All set
	return

}

// Reset I2C to a known good state
func cardResetI2C(context *Context) (err error) {

	// Synchronize by guaranteeing not only that I2C works, but that we drain the remainder of any
	// pending partial reply from a previously-aborted session.
	chunklen := 0
	for {

		// Read the next chunk of available data
		_, available, err2 := i2cReadBytes(chunklen)
		if err2 != nil {
			err = fmt.Errorf("error reading chunk: %s %s", err2, note.ErrCardIo)
			return
		}

		// If nothing left, we're ready to transmit a command to receive the data
		if available == 0 {
			break
		}

		// For the next iteration, reaad the min of what's available and what we're permitted to read
		chunklen = available
		if chunklen > CardI2CMax {
			chunklen = CardI2CMax
		}

	}

	// Done
	return

}

// OpenI2C opens the card on I2C
func OpenI2C(port string, portConfig int) (context Context, err error) {

	// Use default if not specified
	if port == "" {
		port, portConfig = i2cDefault()
	}

	// Set up class functions
	context.PortEnumFn = i2cPortEnum
	context.PortDefaultsFn = i2cDefault
	context.CloseFn = cardCloseI2C
	context.ReopenFn = cardReopenI2C
	context.ResetFn = cardResetI2C
	context.TransactionFn = cardTransactionI2C

	// Record I2C configuration
	context.isSerial = false
	context.i2cName = port
	context.i2cAddress = portConfig

	// Open the I2C port
	err = i2cOpen(port, portConfig)
	if err != nil {
		if false {
			ports, _, _, _ := I2CPorts()
			fmt.Printf("Available ports: %v\n", ports)
		}
		err = fmt.Errorf("i2c init error: %s", err)
		return
	}

	// Done
	return

}

// Reset the port
func (context *Context) Reset() (err error) {
	context.resetRequired = false
	return context.ResetFn(context)
}

// Close the port
func (context *Context) Close() {
	context.CloseFn(context)
}

// Close serial
func cardCloseSerial(context *Context) {
	if context.openSerialPort != nil {
		context.openSerialPort.Close()
		context.openSerialPort = nil
	}
}

// Close I2C
func cardCloseI2C(context *Context) {
	i2cClose()
}

// ReopenIfRequired reopens the port but only if required
func (context *Context) ReopenIfRequired() (err error) {
	if context.reopenRequired {
		err = context.ReopenFn(context)
	}
	return
}

// Reopen the port
func (context *Context) Reopen() (err error) {
	context.reopenRequired = false
	err = context.ReopenFn(context)
	return
}

// Reopen serial
func cardReopenSerial(context *Context) (err error) {

	// Handle deferred insertion
	if context.serialName == "" {
		context.serialName, context.serialConfig.BaudRate = serialDefault()
	}

	// Close if open
	cardCloseSerial(context)

	// Open the serial port
	context.openSerialPort, err = serial.Open(context.serialName, &context.serialConfig)
	if err != nil {
		context.openSerialPort = nil
		return fmt.Errorf("error opening serial port %s at %d: %s %s", context.serialName, context.serialConfig.BaudRate, err, note.ErrCardIo)
	}

	// Reset serial to a known good state
	return cardResetSerial(context)
}

// Reopen I2C
func cardReopenI2C(context *Context) (err error) {
	fmt.Printf("error i2c reopen not yet supported since I can't test it yet")
	return
}

// SerialDefaults returns the default serial parameters
func SerialDefaults() (port string, portConfig int) {
	return serialDefault()
}

// I2CDefaults returns the default serial parameters
func I2CDefaults() (port string, portConfig int) {
	return i2cDefault()
}

// SerialPorts returns the list of available serial ports
func SerialPorts() (allports []string, usbports []string, notecardports []string, err error) {
	return serialPortEnum()
}

// I2CPorts returns the list of available I2C ports
func I2CPorts() (allports []string, usbports []string, notecardports []string, err error) {
	return i2cPortEnum()
}

// TransactionRequest performs a card transaction with a Req structure
func (context *Context) TransactionRequest(req Request) (rsp Request, err error) {

	// Marshal the request to JSON
	reqJSON, err2 := note.JSONMarshal(req)
	if err2 != nil {
		err = fmt.Errorf("error marshaling request for module: %s", err2)
		return
	}

	// Perform the transaction in a way that ALSO assumes that the JSON coming back from
	// the device is unmarshalled even on error.
	var rspJSON []byte
	rspJSON, err = context.TransactionJSON(reqJSON)
	note.JSONUnmarshal(rspJSON, &rsp)

	// Done
	return

}

// NewRequest creates a new request.  Note that this method is provided
// merely as syntactic sugar, as of the form
// req := note.NewRequest("note.add")
func NewRequest(reqType string) (req map[string]interface{}) {
	req["req"] = reqType
	return
}

// NewBody creates a new body.  Note that this method is provided
// merely as syntactic sugar, as of the form
// body := note.NewBody()
func NewBody() (body map[string]interface{}) {
	return
}

// Request performs a card transaction with a JSON structure and doesn't return a response
// (This is for semantic compatibility with other languages.)
func (context *Context) Request(req map[string]interface{}) (err error) {
	_, err = context.Transaction(req)
	return
}

// RequestResponse performs a card transaction with a JSON structure and doesn't return a response
// (This is for semantic compatibility with other languages.)
func (context *Context) RequestResponse(req map[string]interface{}) (rsp map[string]interface{}, err error) {
	return context.Transaction(req)
}

// Response is used in rare cases where there is a transaction that returns multiple responses
func (context *Context) Response() (rsp map[string]interface{}, err error) {
	return context.Transaction(nil)
}

// Transaction performs a card transaction with a JSON structure
func (context *Context) Transaction(req map[string]interface{}) (rsp map[string]interface{}, err error) {

	// Handle the special case where we are just processing a response
	var reqJSON []byte
	if req == nil {

		reqJSON = []byte("")

	} else {

		// Marshal the request to JSON
		reqJSON, err = note.JSONMarshal(req)
		if err != nil {
			err = fmt.Errorf("error marshaling request for module: %s", err)
			return
		}

	}

	// Perform the transaction
	rspJSON, err2 := context.TransactionJSON(reqJSON)
	if err2 != nil {
		err = fmt.Errorf("error from TransactionJSON: %s", err2)
		return
	}

	// Unmarshal for convenience of the caller
	err = note.JSONUnmarshal(rspJSON, &rsp)
	if err != nil {
		err = fmt.Errorf("error unmarshaling reply from module: %s %s", err, note.ErrCardIo)
		return
	}

	// Done
	return
}

// TransactionJSON performs a card transaction using raw JSON []bytes
func (context *Context) TransactionJSON(reqJSON []byte) (rspJSON []byte, err error) {

	var req Request

	// Handle the special case where we are just processing a response (used by test fixture)
	if len(reqJSON) > 0 {

		// Make sure that it is valid JSON, because the transports won't validate this
		// and they may misbehave if they do not get a valid JSON response back.
		err = note.JSONUnmarshal(reqJSON, &req)
		if err != nil {
			return
		}

		// Make sure that the JSON has a single \n terminator
		for {
			if strings.HasSuffix(string(reqJSON), "\n") {
				reqJSON = []byte(strings.TrimSuffix(string(reqJSON), "\n"))
				continue
			}
			if strings.HasSuffix(string(reqJSON), "\r") {
				reqJSON = []byte(strings.TrimSuffix(string(reqJSON), "\r"))
				continue
			}
			break
		}
		reqJSON = []byte(string(reqJSON) + "\n")
	}

	// Only one caller at a time
	transLock.Lock()

	// Do a reset if one was pending
	if context.resetRequired {
		context.Reset()
	}

	// Reopen if error
	if context.reopenRequired {
		err = context.Reopen()
		if err != nil {
			transLock.Unlock()
			return
		}
	}

	// Debug
	if context.Debug {
		var j []byte
		if context.Pretty {
			j, _ = note.JSONMarshalIndent(req, "", "    ")
		} else {
			j, _ = note.JSONMarshal(req)
		}
		fmt.Printf("%s\n", string(j))
	}

	// Perform the transaction and set ERR if there is an I/O error
	rspJSON, err = context.TransactionFn(context, reqJSON)
	if err != nil {
		context.resetRequired = true
	}

	// Decode the response to create an error if the transaction returned an error.  We
	// do this because it's SUPER inconvenient to always be checking for a response error
	// vs an error on the transaction itself
	var rsp Request
	if err == nil && note.JSONUnmarshal(rspJSON, &rsp) == nil && rsp.Err != "" {
		if req.Req == "" {
			err = fmt.Errorf("%s", rsp.Err)
		} else {
			err = fmt.Errorf("%s: %s", req.Req, rsp.Err)
		}
	}

	// If this was a card restore, we know that a reopen will be required
	if !context.isRemote && (req.Req == ReqCardRestore || req.Req == ReqCardRestart) {
		time.Sleep(8 * time.Second)
		context.reopenRequired = true
	}

	// Debug
	if context.Debug {
		responseJSON := rspJSON
		if context.Pretty {
			prettyJSON, e := note.JSONMarshalIndent(rsp, "    ", "    ")
			if e == nil {
				fmt.Printf("==> ")
				responseJSON = append(prettyJSON, byte('\n'))
			}
		}
		fmt.Printf("%s", string(responseJSON))
	}

	// Done
	transLock.Unlock()
	return

}

// Perform a card transaction over serial under the assumption that request already has '\n' terminator
func cardTransactionSerial(context *Context, reqJSON []byte) (rspJSON []byte, err error) {

	// Exit if not open
	if context.openSerialPort == nil {
		err = fmt.Errorf("port not open " + note.ErrCardIo)
		cardReportError(context, err)
		return
	}

	// Handle the special case where we are looking only for a reply
	if len(reqJSON) > 0 {

		// Transmit the request in segments so as not to overwhelm the notecard's interrupt buffers
		segOff := 0
		segLeft := len(reqJSON)
		for {
			segLen := segLeft
			if segLen > CardRequestSegmentMaxLen {
				segLen = CardRequestSegmentMaxLen
			}
			_, err = context.openSerialPort.Write(reqJSON[segOff : segOff+segLen])
			if err != nil {
				err = fmt.Errorf("error transmitting to module: %s %s", err, note.ErrCardIo)
				cardReportError(context, err)
				return
			}
			segOff += segLen
			segLeft -= segLen
			if segLeft == 0 {
				break
			}
			time.Sleep(CardRequestSegmentDelayMs * time.Millisecond)
		}

	}

	// Read the reply until we get '\n' at the end
	waitBeganSecs := time.Now().Unix()
	for {
		var length int
		buf := make([]byte, 2048)
		readBeganMs := int(time.Now().UnixNano() / 1000000)
		length, err = context.openSerialPort.Read(buf)
		readElapsedMs := int(time.Now().UnixNano()/1000000) - readBeganMs
		if false {
			err2 := err
			if err2 == nil {
				err2 = fmt.Errorf("none")
			}
			fmt.Printf("req: elapsed:%d len:%d err:%s '%s'\n", readElapsedMs, length, err2, string(buf[:length]))
		}
		if readElapsedMs == 0 && length == 0 && err == io.EOF {
			// On Linux, hardware port failures come back simply as immediate EOF
			err = fmt.Errorf("hardware failure")
		}
		if err != nil {
			if err == io.EOF {
				// Just a read timeout
				continue
			}
			// Ignore [flaky, rare, Windows] hardware errors for up to several seconds
			if (time.Now().Unix() - waitBeganSecs) > 2 {
				err = fmt.Errorf("error reading from module: %s %s", err, note.ErrCardIo)
				cardReportError(context, err)
				return
			}
			time.Sleep(1 * time.Second)
			continue
		}
		rspJSON = append(rspJSON, buf[:length]...)
		if strings.HasSuffix(string(rspJSON), "\n") {
			break
		}
	}

	// Done
	return

}

// Perform a card transaction over I2C under the assumption that request already has '\n' terminator
func cardTransactionI2C(context *Context, reqJSON []byte) (rspJSON []byte, err error) {

	// Transmit the request in chunks, but also in segments so as not to overwhelm the notecard's interrupt buffers
	chunkoffset := 0
	jsonbufLen := len(reqJSON)
	sentInSegment := 0
	for jsonbufLen > 0 {
		chunklen := CardI2CMax
		if jsonbufLen < chunklen {
			chunklen = jsonbufLen
		}
		err = i2cWriteBytes(reqJSON[chunkoffset : chunkoffset+chunklen])
		if err != nil {
			err = fmt.Errorf("write error: %s %s", err, note.ErrCardIo)
			return
		}
		chunkoffset += chunklen
		jsonbufLen -= chunklen
		sentInSegment += chunklen
		if sentInSegment > CardRequestSegmentMaxLen {
			sentInSegment -= CardRequestSegmentMaxLen
		}
		time.Sleep(CardRequestSegmentDelayMs * time.Millisecond)
	}

	// Loop, building a reply buffer out of received chunks.  We'll build the reply in the same
	// buffer we used to transmit, and will grow it as necessary.
	jsonbufLen = 0
	receivedNewline := false
	chunklen := 0
	expires := time.Now().Add(time.Duration(10 * time.Second))
	for {

		// Read the next chunk
		readbuf, available, err2 := i2cReadBytes(chunklen)
		if err2 != nil {
			err = fmt.Errorf("read error: %s %s", err2, note.ErrCardIo)
			return
		}

		// Append to the JSON being accumulated
		rspJSON = append(rspJSON, readbuf...)
		readlen := len(readbuf)
		jsonbufLen += readlen

		// If we received something, don't time out
		if readlen > 0 {
			expires = time.Now().Add(time.Duration(10 * time.Second))
		}

		// If the last byte of the chunk is \n, chances are that we're done.  However, just so
		// that we pull everything pending from the module, we only exit when we've received
		// a newline AND there's nothing left available from the module.
		if readlen > 0 && readbuf[readlen-1] == '\n' {
			receivedNewline = true
		}

		// For the next iteration, reaad the min of what's available and what we're permitted to read
		chunklen = available
		if chunklen > CardI2CMax {
			chunklen = CardI2CMax
		}

		// If there's something available on the notecard for us to receive, do it
		if chunklen > 0 {
			continue
		}

		// If there's nothing available and we received a newline, we're done
		if receivedNewline {
			break
		}

		// If we've timed out and nothing's available, exit
		if time.Now().After(expires) {
			err = fmt.Errorf("transaction timeout")
			return
		}

	}

	// Done
	return
}

// OpenRemote opens a remote card
func OpenRemote(farmURL string, farmCheckoutMins int) (context Context, err error) {

	// Set up class functions
	context.isRemote = true
	context.PortEnumFn = remotePortEnum
	context.PortDefaultsFn = remotePortDefault
	context.CloseFn = remoteClose
	context.ReopenFn = remoteReopen
	context.ResetFn = remoteReset
	context.TransactionFn = remoteTransaction

	// Record serial configuration
	context.farmURL = farmURL
	if farmCheckoutMins == 0 {
		farmCheckoutMins = 1
	}
	farmCheckoutMins = (((farmCheckoutMins - 1) / reservationModulusMinutes) + 1) * reservationModulusMinutes
	context.farmCheckoutMins = farmCheckoutMins
	context.farmCheckoutExpires = time.Now().Unix() + int64(context.farmCheckoutMins*60)

	// Open the port
	err = context.ReopenFn(&context)
	if err != nil {
		err = fmt.Errorf("error opening remote %s: %s %s", farmURL, err, note.ErrCardIo)
		return
	}

	// All set
	return

}
