// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"os"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blues/note-go/note"
	"go.bug.st/serial"
)

// Debug serial I/O
var debugSerialIO = false

// InitialDebugMode is the debug mode that the context is initialized with
var InitialDebugMode = false

// InitialTraceMode is whether or not we will be entering trace mode, to prevent reservationsa
var InitialTraceMode = false

// InitialResetMode says whether or not we should reset the port on entry
var InitialResetMode = true

// Protect against multiple concurrent callers, because across different operating systems it is
// not at all clear that concurrency is allowed on a single I/O device.  An exception is made
// for the I2C 'multiport' case (exposed by TransactionRequestToPort) where we allow multiple
// concurrent I2C transactions on a single device.  (This capability was needed for the
// Notefarm, but it's unclear if anyone uses this multi-notecard concurrency capability anymore
// now that it's deprecated.)
var (
	transLock          sync.RWMutex
	multiportTransLock [128]sync.RWMutex
)

// SerialTimeoutMs is the response timeout for Notecard serial communications.
var SerialTimeoutMs = 30000

// IgnoreWindowsHWErrSecs is the amount of time to ignore a Windows serial communiction error.
var IgnoreWindowsHWErrSecs = 2

// Module communication interfaces
const (
	NotecardInterfaceSerial = "serial"
	NotecardInterfaceI2C    = "i2c"
	NotecardInterfaceLease  = "lease"
)

// The number of minutes that we'll round up so that notecard reservations don't thrash
const reservationModulusMinutes = 5

// CardI2CMax controls chunk size that's socially appropriate on the I2C bus.
// It must be 1-253 bytes as per spec (which allows space for the 2-byte header in a 255-byte read)
const CardI2CMax = 253

// The notecard is a real-time device that has a fixed size interrupt buffer.  We can push data
// at it far, far faster than it can process it, therefore we push it in segments with a pause
// between each segment.

// CardRequestSerialSegmentMaxLen (golint)
const CardRequestSerialSegmentMaxLen = 250

// CardRequestSerialSegmentDelayMs (golint)
const CardRequestSerialSegmentDelayMs = 250

// CardRequestI2CSegmentMaxLen (golint)
const CardRequestI2CSegmentMaxLen = 250

// CardRequestI2CSegmentDelayMs (golint)
const CardRequestI2CSegmentDelayMs = 250

// RequestSegmentMaxLen (golint)
var RequestSegmentMaxLen = -1

// RequestSegmentDelayMs (golint)
var RequestSegmentDelayMs = -1

var DoNotReterminateJSON = false

// Transaction retry logic
const requestRetriesAllowed = 5

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

	// Disable generation of User Agent object
	DisableUA bool

	// Reset should be done on next transaction
	resetRequired       bool
	reopenRequired      bool
	reopenBecauseOfOpen bool

	// Sequence number
	lastRequestSeqno int

	// Class functions
	PortEnumFn     func() (allports []string, usbports []string, notecardports []string, err error)
	PortDefaultsFn func() (port string, portConfig int)
	CloseFn        func(context *Context)
	ReopenFn       func(context *Context, portConfig int) (err error)
	ResetFn        func(context *Context, portConfig int) (err error)
	TransactionFn  func(context *Context, portConfig int, noResponse bool, reqJSON []byte) (rspJSON []byte, err error)

	// Trace functions
	traceOpenFn  func(context *Context) (err error)
	traceReadFn  func(context *Context) (data []byte, err error)
	traceWriteFn func(context *Context, data []byte)

	// Port data
	iface      string
	isLocal    bool
	port       string
	portConfig int
	portIsOpen bool

	// Serial instance state
	isSerial         bool
	serialPort       serial.Port
	serialUseDefault bool
	serialName       string
	serialConfig     serial.Mode

	// Serial I/O timeout helpers
	ioStartSignal    chan int
	ioCompleteSignal chan bool
	ioTimeoutSignal  chan bool

	// I2C
	i2cMultiport bool

	// Lease state
	leaseScope     string
	leaseExpires   int64
	leaseLessor    string
	leaseDeviceUID string
	leaseTraceConn net.Conn
}

// Report a critical card error
func cardReportError(context *Context, err error) {
	if context == nil {
		return
	}
	if context.Debug {
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
	if context.PortEnumFn == nil {
		return
	}
	return context.PortEnumFn()
}

// PortDefaults gets the defaults for the specified port
func (context *Context) PortDefaults() (port string, portConfig int) {
	if context.PortDefaultsFn == nil {
		return
	}
	return context.PortDefaultsFn()
}

// Identify this Notecard connection
func (context *Context) Identify() (protocol string, port string, portConfig int) {
	if context.isSerial {
		return "serial", context.serialName, context.serialConfig.BaudRate
	}
	return "I2C", context.port, context.portConfig
}

// Defaults gets the default interface, port, and config
func Defaults() (moduleInterface string, port string, portConfig int) {
	moduleInterface = NotecardInterfaceSerial
	port, portConfig = serialDefault()
	return
}

// Open the card to establish communications
func Open(moduleInterface string, port string, portConfig int) (context *Context, err error) {
	if moduleInterface == "" {
		moduleInterface, _, _ = Defaults()
	}

	switch moduleInterface {
	case NotecardInterfaceSerial:
		context, err = OpenSerial(port, portConfig)
		context.isLocal = true
	case NotecardInterfaceI2C:
		context, err = OpenI2C(port, portConfig)
		context.isLocal = true
	case NotecardInterfaceLease:
		context, err = OpenLease(port, portConfig)
	default:
		err = fmt.Errorf("unknown interface: %s", moduleInterface)
	}
	if err != nil {
		cardReportError(nil, err)
		err = fmt.Errorf("error opening port: %s %s", err, note.ErrCardIo)
		return
	}
	context.iface = moduleInterface
	return
}

// Reset serial to a known state
func cardResetSerial(context *Context, portConfig int) (err error) {
	// Exit if not open
	if !context.portIsOpen {
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
		if debugSerialIO {
			fmt.Printf("cardResetSerial: about to write newline\n")
		}
		serialIOBegin(context)
		_, err = context.serialPort.Write([]byte("\n"))
		err = serialIOEnd(context, err)
		if debugSerialIO {
			fmt.Printf("                 back with err = %v\n", err)
		}
		if err != nil {
			err = fmt.Errorf("error transmitting to module: %s %s", err, note.ErrCardIo)
			cardReportError(context, err)
			return
		}
		time.Sleep(750 * time.Millisecond)
		if debugSerialIO {
			fmt.Printf("cardResetSerial: about to read up to %d bytes\n", len(buf))
		}
		readBeganMs := int(time.Now().UnixNano() / 1000000)
		serialIOBegin(context)
		length, err = context.serialPort.Read(buf)
		err = serialIOEnd(context, err)
		readElapsedMs := int(time.Now().UnixNano()/1000000) - readBeganMs
		if debugSerialIO {
			fmt.Printf("                 back after %d ms with len = %d err = %v\n", readElapsedMs, length, err)
		}
		if readElapsedMs == 0 && length == 0 && err == io.EOF {
			// On Linux, hardware port failures come back simply as immediate EOF
			err = fmt.Errorf("hardware failure")
		}
		if err != nil {
			err = fmt.Errorf("error reading from module after reset: %s %s", err, note.ErrCardIo)
			cardReportError(context, err)
			return
		}
		somethingFound := false
		nonCRLFFound := false
		for i := 0; i < length && !nonCRLFFound; i++ {
			if false {
				fmt.Printf("chr: 0x%02x '%c'\n", buf[i], buf[i])
			}
			if buf[i] != '\r' {
				somethingFound = true
				if buf[i] != '\n' {
					nonCRLFFound = true
				}
			}
		}
		if somethingFound && !nonCRLFFound {
			break
		}
	}

	// Done
	return
}

// Serial I/O timeout helper function for Windows
func serialTimeoutHelper(context *Context, portConfig int) {
	for {
		timeoutMs := <-context.ioStartSignal
		timeout := false
		select {
		case <-context.ioCompleteSignal:
		case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
			timeout = true
			if debugSerialIO {
				fmt.Printf("serialTimeoutHelper: timeout\n")
			}
			cardCloseSerial(context)
		}
		context.ioTimeoutSignal <- timeout
	}
}

// Begin a serial I/O
func serialIOBegin(context *Context) {
	timeoutMs := SerialTimeoutMs
	context.ioStartSignal <- timeoutMs
	if debugSerialIO {
		if !context.portIsOpen {
			fmt.Printf("serialIoBegin: WARNING: PORT NOT OPEN\n")
		}
		fmt.Printf("serialIOBegin: begin timeout of %d ms\n", timeoutMs)
	}
}

// End a serial I/O
func serialIOEnd(context *Context, errIn error) (errOut error) {
	errOut = errIn
	context.ioCompleteSignal <- true
	timeout := <-context.ioTimeoutSignal
	select {
	case <-context.ioCompleteSignal:
		if debugSerialIO {
			fmt.Printf("serialIOEnd: ioComplete ate the completed signal (timeout: %v)\n", timeout)
		}
	default:
		if debugSerialIO {
			fmt.Printf("serialIOEnd: ioComplete nothing to eat (timeout: %v)\n", timeout)
		}
	}
	if timeout {
		errOut = fmt.Errorf("serial I/O timeout %s", note.ErrCardIo)
	}
	return
}

// OpenSerial opens the card on serial
func OpenSerial(port string, portConfig int) (context *Context, err error) {
	// Create the context structure
	context = &Context{}
	context.Debug = InitialDebugMode
	context.port = port
	context.portConfig = portConfig
	context.lastRequestSeqno = 0

	// Set up class functions
	context.PortEnumFn = serialPortEnum
	context.PortDefaultsFn = serialDefault
	context.CloseFn = cardCloseSerial
	context.ReopenFn = cardReopenSerial
	context.ResetFn = cardResetSerial
	context.TransactionFn = cardTransactionSerial
	context.traceOpenFn = serialTraceOpen
	context.traceReadFn = serialTraceRead
	context.traceWriteFn = serialTraceWrite

	// Record serial configuration, and whether or not we are using the default
	context.isSerial = true
	if port == "" {
		context.serialUseDefault = true
		context.serialName, context.serialConfig.BaudRate = serialDefault()
	} else {
		context.serialName = port
		context.serialConfig.BaudRate = portConfig
	}

	// Set up I/O port close channels, because Windows needs a bit of help in timing out I/O's.
	context.ioStartSignal = make(chan int, 1)
	context.ioCompleteSignal = make(chan bool, 1)
	context.ioTimeoutSignal = make(chan bool, 1)
	go serialTimeoutHelper(context, portConfig)

	// For serial, we defer the port open until the first transaction so that we can
	// support the concept of dynamically inserted devices, as in "notecard -scan" mode.
	context.reopenBecauseOfOpen = true
	context.reopenRequired = true

	// All set
	return
}

// Reset I2C to a known good state
func cardResetI2C(context *Context, portConfig int) (err error) {
	// Synchronize by guaranteeing not only that I2C works, but that we drain the remainder of any
	// pending partial reply from a previously-aborted session.
	chunklen := 0
	for {

		// Read the next chunk of available data
		_, available, err2 := i2cReadBytes(chunklen, portConfig)
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
func OpenI2C(port string, portConfig int) (context *Context, err error) {

	// Create the context structure
	context = &Context{}
	context.Debug = InitialDebugMode
	context.lastRequestSeqno = 0

	// Open
	context.portIsOpen = false

	// Use default if not specified
	if port == "" {
		port, portConfig = i2cDefault()
	}
	context.port = port
	context.portConfig = portConfig

	// Set up class functions
	context.PortEnumFn = i2cPortEnum
	context.PortDefaultsFn = i2cDefault
	context.CloseFn = cardCloseI2C
	context.ReopenFn = cardReopenI2C
	context.ResetFn = cardResetI2C
	context.TransactionFn = cardTransactionI2C

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

	// Open
	context.portIsOpen = true

	// Done
	return
}

// Reset the port
func (context *Context) Reset(portConfig int) (err error) {
	context.resetRequired = false
	if context.ResetFn == nil {
		return
	}
	return context.ResetFn(context, portConfig)
}

// Close the port
func (context *Context) Close() {
	context.CloseFn(context)
}

// Close serial
func cardCloseSerial(context *Context) {
	if !context.portIsOpen {
		if debugSerialIO {
			fmt.Printf("cardCloseSerial: port not open\n")
		}
	} else {
		if debugSerialIO {
			fmt.Printf("cardCloseSerial: closed\n")
		}
		context.serialPort.Close()
		context.portIsOpen = false
	}
}

// Close I2C
func cardCloseI2C(context *Context) {
	i2cClose()
	context.portIsOpen = false
}

// ReopenIfRequired reopens the port but only if required
func (context *Context) ReopenIfRequired(portConfig int) (err error) {
	if context.reopenRequired {
		err = context.ReopenFn(context, portConfig)
	}
	return
}

// Reopen the port
func (context *Context) Reopen(portConfig int) (err error) {
	context.reopenRequired = false
	err = context.ReopenFn(context, portConfig)
	return
}

// Reopen serial
func cardReopenSerial(context *Context, portConfig int) (err error) {

	// Close if open
	cardCloseSerial(context)

	// Handle deferred insertion
	if context.serialUseDefault {
		context.serialName, context.serialConfig.BaudRate = serialDefault()
	}
	if context.serialName == "" {
		return fmt.Errorf("error opening serial port: serial device not available %s", note.ErrCardIo)
	}

	// Set default speed if not set
	if context.serialConfig.BaudRate == 0 {
		_, context.serialConfig.BaudRate = serialDefault()
	}

	// Open the serial port
	if debugSerialIO {
		fmt.Printf("cardReopenSerial: about to open '%s'\n", context.serialName)
	}
	context.serialPort, err = serial.Open(context.serialName, &context.serialConfig)
	if debugSerialIO {
		fmt.Printf("                  back with err = %v\n", err)
	}
	if err != nil {
		return fmt.Errorf("error opening serial port %s at %d: %s %s", context.serialName, context.serialConfig.BaudRate, err, note.ErrCardIo)
	}

	context.portIsOpen = true

	// Done with the reopen
	context.reopenRequired = false

	// Unless we've been instructed not to reset on open, reset serial to a known good state
	if context.reopenBecauseOfOpen && InitialResetMode {
		err = cardResetSerial(context, portConfig)
	}
	context.reopenBecauseOfOpen = false

	// Done
	return err
}

// Reopen I2C
func cardReopenI2C(context *Context, portConfig int) (err error) {
	fmt.Printf("error i2c reopen not yet supported since I can't test it yet\n")
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
	return context.transactionRequest(req, false, 0)
}

// TransactionRequestToPort performs a card transaction with a Req structure, to a specified port
func (context *Context) TransactionRequestToPort(req Request, portConfig int) (rsp Request, err error) {
	return context.transactionRequest(req, true, portConfig)
}

// transactionRequest performs a card transaction with a Req structure, to the current or specified port
func (context *Context) transactionRequest(req Request, multiport bool, portConfig int) (rsp Request, err error) {
	reqJSON, err2 := note.JSONMarshal(req)
	if err2 != nil {
		err = fmt.Errorf("error marshaling request for module: %s", err2)
		return
	}
	var rspJSON []byte
	rspJSON, err = context.transactionJSON(reqJSON, multiport, portConfig)
	if err != nil {
		// Give transaction's error precedence, except that if we get an error unmarshaling
		// we want to make sure that we indicate to the caller that there was an I/O error (corruption)
		err2 := note.JSONUnmarshal(rspJSON, &rsp)
		if err2 != nil {
			err = fmt.Errorf("%s %s", err, note.ErrCardIo)
		}
		return
	}
	err = note.JSONUnmarshal(rspJSON, &rsp)
	if err != nil {
		err = fmt.Errorf("error unmarshaling reply from module: %s %s: %s", err, note.ErrCardIo, rspJSON)
	}
	return
}

// NewRequest creates a new request that is guaranteed to get a response
// from the Notecard.  Note that this method is provided merely as syntactic sugar, as of the form
// req := notecard.NewRequest("note.add")
func NewRequest(reqType string) (req map[string]interface{}) {
	return map[string]interface{}{
		"req": reqType,
	}
}

// NewCommand creates a new command that requires no response from the notecard.
func NewCommand(reqType string) (cmd map[string]interface{}) {
	return map[string]interface{}{
		"cmd": reqType,
	}
}

// NewBody creates a new body.  Note that this method is provided
// merely as syntactic sugar, as of the form
// body := note.NewBody()
func NewBody() (body map[string]interface{}) {
	return make(map[string]interface{})
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
		err = fmt.Errorf("error unmarshaling reply from module: %s %s: %s", err, note.ErrCardIo, rspJSON)
		return
	}

	// Done
	return
}

// ReceiveBytes receives arbitrary Bytes from the Notecard
func (context *Context) ReceiveBytes() (rspBytes []byte, err error) {
	return context.receiveBytes(0)
}

// SendBytes sends arbitrary Bytes to the Notecard
func (context *Context) SendBytes(reqBytes []byte) (err error) {

	// Only operate on port 0
	portConfig := 0

	// Only one caller at a time accessing the I/O port
	lockTrans(false, portConfig)

	// Reopen if error
	err = context.ReopenIfRequired(portConfig)
	if err != nil {
		unlockTrans(false, portConfig)
		return
	}

	// Do a reset if one was pending
	if context.resetRequired {
		context.Reset(portConfig)
	}

	// Do the send, with no response requested
	_, err = context.TransactionFn(context, portConfig, true, reqBytes)

	// Done
	unlockTrans(false, portConfig)
	return

}

// receiveBytes receives arbitrary Bytes from the Notecard, using  the current or specified port
func (context *Context) receiveBytes(portConfig int) (rspBytes []byte, err error) {
	// Only one caller at a time accessing the I/O port
	lockTrans(false, portConfig)

	// Reopen if error
	err = context.ReopenIfRequired(portConfig)
	if err != nil {
		unlockTrans(false, portConfig)
		if context.Debug {
			fmt.Printf("%s\n", err)
		}
		return
	}

	// Do a reset if one was pending
	if context.resetRequired {
		context.Reset(portConfig)
	}

	// Request is empty
	var reqBytes []byte
	// Perform the transaction
	rspBytes, err = context.TransactionFn(context, portConfig, false, reqBytes)

	unlockTrans(false, portConfig)

	// Done
	return
}

// TransactionJSON performs a card transaction using raw JSON []bytes
func (context *Context) TransactionJSON(reqJSON []byte) (rspJSON []byte, err error) {
	return context.transactionJSON(reqJSON, false, 0)
}

// TransactionJSONToPort performs a card transaction using raw JSON []bytes to a specified port
func (context *Context) TransactionJSONToPort(reqJSON []byte, portConfig int) (rspJSON []byte, err error) {
	return context.transactionJSON(reqJSON, true, portConfig)
}

// transactionJSON performs a card transaction using raw JSON []bytes, to the current or specified port
func (context *Context) transactionJSON(reqJSON []byte, multiport bool, portConfig int) (rspJSON []byte, err error) {
	// Remember in the context if we've ever seen multiport I/O, for timeout computation
	if multiport {
		context.i2cMultiport = true
	}

	// Unmarshal the request to peek inside it.  Also, accept a zero-length request as a valid case
	// because we use this in the test fixture where  we just accept pure responses w/o requests.
	var req Request
	var noResponseRequested bool
	if len(reqJSON) > 0 {

		// Make sure that it is valid JSON, because the transports won't validate this
		// and they may misbehave if they do not get a valid JSON response back.
		err = note.JSONUnmarshal(reqJSON, &req)
		if err != nil {
			return
		}

		// If this is a hub.set, generate a user agent object if one hasn't already been supplied
		if !context.DisableUA && (req.Req == ReqHubSet || req.Cmd == ReqHubSet) && req.Body == nil {
			ua := context.UserAgent()
			if ua != nil {
				req.Body = &ua
				reqJSON, _ = note.JSONMarshal(req)
			}
		}

		// Determine whether or not a response will be expected from the notecard by
		// examining the req and cmd fields
		noResponseRequested = req.Req == "" && req.Cmd != ""

		if !DoNotReterminateJSON {
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

	// If it is a request (as opposed to a command), include a CRC so that the
	// request might be retried if it is received in a corrupted state.  (We can
	// only do this on requests because for cmd's there is no 'response channel'
	// where we can find out that the cmd failed.  Note that a Seqno is included
	// as part of the CRC data so that two identical requests occurring within the
	// modulus of seqno never are mistaken as being the same request being retried.
	lastRequestRetries := 0
	lastRequestCrcAdded := false
	if !noResponseRequested {
		reqJSON = crcAdd(reqJSON, context.lastRequestSeqno)
		lastRequestCrcAdded = true
	}

	// Only one caller at a time accessing the I/O port
	lockTrans(multiport, portConfig)

	// Transaction retry loop.  Note that "err" must be set before breaking out of loop
	err = nil
	for {
		if lastRequestRetries > requestRetriesAllowed {
			break
		}

		// Only do reopen/reset in the single-port case, because we may not be talking to the port in error
		if !multiport {

			// Reopen if error
			err = context.ReopenIfRequired(portConfig)
			if err != nil {
				unlockTrans(multiport, portConfig)
				if context.Debug {
					fmt.Printf("%s\n", err)
				}
				return
			}

			// Do a reset if one was pending
			if context.resetRequired {
				context.Reset(portConfig)
			}

		}

		// Perform the transaction
		rspJSON, err = context.TransactionFn(context, portConfig, noResponseRequested, reqJSON)
		if err != nil {
			// We can defer the error if a single port, but we need to reset it NOW if multiport
			if multiport {
				if context.ResetFn != nil {
					context.ResetFn(context, portConfig)
				}
			} else {
				context.resetRequired = true
			}
		}

		// If no response expected, we won't be retrying
		if noResponseRequested {
			break
		}

		// Decode the response to create an error if the response JSON was badly formatted.
		// do this because it's SUPER inconvenient to always be checking for a response error
		// vs an error on the transaction itself
		if err == nil {
			var rsp Request
			err = note.JSONUnmarshal(rspJSON, &rsp)
			if err != nil {
				err = fmt.Errorf("error unmarshaling reply from module: %s %s: %s", err, note.ErrCardIo, rspJSON)
			} else {
				if rsp.Err != "" {
					if req.Req == "" {
						err = fmt.Errorf("%s", rsp.Err)
					} else {
						err = fmt.Errorf("%s: %s", req.Req, rsp.Err)
					}
				}
			}
		}

		// Don't retry these transactions for obvious reasons
		if req.Req == ReqCardRestore || req.Req == ReqCardRestart {
			break
		}

		// If an I/O error, retry
		if note.ErrorContains(err, note.ErrCardIo) && !note.ErrorContains(err, note.ErrReqNotSupported) {
			// We can defer the error if a single port, but we need to reset it NOW if multiport
			if multiport {
				if context.ResetFn != nil {
					context.ResetFn(context, portConfig)
				}
			} else {
				context.resetRequired = true
			}
			lastRequestRetries++
			if context.Debug {
				fmt.Printf("retrying I/O error detected by host: %s\n", err)
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// If an error, stop transaction processing here
		if err != nil {
			break
		}

		// If we sent a CRC in the request, examine the response JSON to see if
		// it has a CRC error.  Note that the CRC is stripped from the
		// rspJSON as a side-effect of this method.
		if lastRequestCrcAdded {
			rspJSON, err = crcError(rspJSON, context.lastRequestSeqno)
			if err != nil {
				lastRequestRetries++
				if context.Debug {
					fmt.Printf("retrying: %s\n", err)
				}
				time.Sleep(500 * time.Millisecond)
				continue
			}

		}

		// Transaction completed
		break

	}

	// Bump the request sequence number now that we've processed this request, success or error
	context.lastRequestSeqno++

	// If this was a card restore, we want to hold everyone back if we reset the card if it
	// isn't a multiport case.  But in multiport, we only want to hold this caller back.
	if (req.Req == ReqCardRestore) && req.Reset {
		// Special case card.restore, reset:true does not cause a reboot.
		unlockTrans(multiport, portConfig)
	} else if context.isLocal && (req.Req == ReqCardRestore || req.Req == ReqCardRestart) {
		if multiport {
			unlockTrans(multiport, portConfig)
			time.Sleep(12 * time.Second)
		} else {
			context.reopenRequired = true
			time.Sleep(8 * time.Second)
			unlockTrans(multiport, portConfig)
		}
	} else {
		unlockTrans(multiport, portConfig)
	}

	// If no response, we're done
	if noResponseRequested {
		rspJSON = []byte("{}")
		return
	}

	// Debug
	if context.Debug {
		responseJSON := rspJSON
		if context.Pretty {
			var rsp Request
			e := note.JSONUnmarshal(responseJSON, &rsp)
			if e == nil {
				prettyJSON, e := note.JSONMarshalIndent(rsp, "    ", "    ")
				if e == nil {
					fmt.Printf("==> ")
					responseJSON = append(prettyJSON, byte('\n'))
				}
			}
		}
		fmt.Printf("%s", string(responseJSON))
	}

	// Done
	return
}

// Perform a card transaction over serial under the assumption that request already has '\n' terminator
func cardTransactionSerial(context *Context, portConfig int, noResponse bool, reqJSON []byte) (rspJSON []byte, err error) {
	// Exit if not open
	if !context.portIsOpen {
		err = fmt.Errorf("port not open " + note.ErrCardIo)
		cardReportError(context, err)
		return
	}

	// Initialize timing parameters
	if RequestSegmentMaxLen < 0 {
		RequestSegmentMaxLen = CardRequestSerialSegmentMaxLen
	}
	if RequestSegmentDelayMs < 0 {
		RequestSegmentDelayMs = CardRequestSerialSegmentDelayMs
	}

	// Handle the special case where we are looking only for a reply
	if len(reqJSON) > 0 {

		// Transmit the request in segments so as not to overwhelm the notecard's interrupt buffers
		segOff := 0
		segLeft := len(reqJSON)
		for {
			segLen := segLeft
			if segLen > RequestSegmentMaxLen {
				segLen = RequestSegmentMaxLen
			}
			if debugSerialIO {
				fmt.Printf("cardTransactionSerial: about to write %d bytes\n", segLen)
			}
			serialIOBegin(context)
			_, err = context.serialPort.Write(reqJSON[segOff : segOff+segLen])
			err = serialIOEnd(context, err)
			if debugSerialIO {
				fmt.Printf("                       back with err = %v\n", err)
			}
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
			time.Sleep(time.Duration(RequestSegmentDelayMs) * time.Millisecond)
		}

	}

	// If no response, we're done
	if noResponse {
		return
	}

	// Read the reply until we get '\n' at the end
	waitBeganSecs := time.Now().Unix()
	for {
		var length int
		buf := make([]byte, 2048)
		if debugSerialIO {
			fmt.Printf("cardTransactionSerial: about to read up to %d bytes\n", len(buf))
		}
		readBeganMs := int(time.Now().UnixNano() / 1000000)
		serialIOBegin(context)
		length, err = context.serialPort.Read(buf)
		err = serialIOEnd(context, err)
		readElapsedMs := int(time.Now().UnixNano()/1000000) - readBeganMs
		if debugSerialIO {
			fmt.Printf("                       back after %d ms with len = %d err = %v\n", readElapsedMs, length, err)
		}
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
			if (time.Now().Unix() - waitBeganSecs) > int64(IgnoreWindowsHWErrSecs) {
				err = fmt.Errorf("error reading from module: %s %s", err, note.ErrCardIo)
				cardReportError(context, err)
				return
			}
			time.Sleep(1 * time.Second)
			continue
		}
		rspJSON = append(rspJSON, buf[:length]...)
		if strings.Contains(string(rspJSON), "\n") {

			// At this point, if we split the string at \n its len must be >= 2
			lines := strings.Split(string(rspJSON), "\n")
			lastLine := lines[len(lines)-1]
			secondToLastLine := lines[len(lines)-2]

			// The reply should be only a single line.  However, if the user had been
			// in trace mode (likely on USB) we may be receiving trace lines that
			// were sent to us and inserted into the serial buffer prior to the JSON reply.
			if lastLine != "" {
				// If the json didn't END in \n, we are still collecting a partial line
				rspJSON = []byte(lastLine)
			} else if len(reqJSON) == 0 {
				// If we're just gathering a reply, we accept binary because it may be COBS
				break
			} else {

				// We're done if and only if the response looks like JSON
				if secondToLastLine[0] == '{' {
					break
				}

				// Drop it, because the line doesn't look like JSON
				rspJSON = []byte{}

			}

		}

	}

	// Done
	return
}

// Perform a card transaction over I2C under the assumption that request already has '\n' terminator
func cardTransactionI2C(context *Context, portConfig int, noResponse bool, reqJSON []byte) (rspJSON []byte, err error) {
	// Initialize timing parameters
	if RequestSegmentMaxLen < 0 {
		RequestSegmentMaxLen = CardRequestI2CSegmentMaxLen
	}
	if RequestSegmentDelayMs < 0 {
		RequestSegmentDelayMs = CardRequestI2CSegmentDelayMs
	}

	// Transmit the request in chunks, but also in segments so as not to overwhelm the notecard's interrupt buffers
	chunkoffset := 0
	jsonbufLen := len(reqJSON)
	sentInSegment := 0
	for jsonbufLen > 0 {
		chunklen := CardI2CMax
		if jsonbufLen < chunklen {
			chunklen = jsonbufLen
		}
		err = i2cWriteBytes(reqJSON[chunkoffset:chunkoffset+chunklen], portConfig)
		if err != nil {
			err = fmt.Errorf("write error: %s %s", err, note.ErrCardIo)
			return
		}
		chunkoffset += chunklen
		jsonbufLen -= chunklen
		sentInSegment += chunklen
		if sentInSegment > RequestSegmentMaxLen {
			sentInSegment = 0
			time.Sleep(time.Duration(RequestSegmentDelayMs) * time.Millisecond)
		}
		time.Sleep(time.Duration(RequestSegmentDelayMs) * time.Millisecond)
	}

	// If no response, we're done
	if noResponse {
		return
	}

	// Loop, building a reply buffer out of received chunks.  We'll build the reply in the same
	// buffer we used to transmit, and will grow it as necessary.
	jsonbufLen = 0
	receivedNewline := false
	chunklen := 0
	expireSecs := 60
	expires := time.Now().Add(time.Duration(expireSecs) * time.Second)
	longExpireSecs := 240
	longexpires := time.Now().Add(time.Duration(longExpireSecs) * time.Second)
	for {

		// Read the next chunk
		readbuf, available, err2 := i2cReadBytes(chunklen, portConfig)
		if err2 != nil {
			err = fmt.Errorf("read error: %s %s", err2, note.ErrCardIo)
			return
		}

		// Append to the JSON being accumulated
		rspJSON = append(rspJSON, readbuf...)
		readlen := len(readbuf)
		jsonbufLen += readlen

		// If we received something, reset the expiration
		if readlen > 0 {
			expires = time.Now().Add(time.Duration(90) * time.Second)
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
		expired := false
		timeoutSecs := 0
		if !context.i2cMultiport || jsonbufLen == 0 {
			expired = time.Now().After(expires)
			timeoutSecs = expireSecs
		} else {
			expired = time.Now().After(longexpires)
			timeoutSecs = longExpireSecs
		}
		if expired {
			err = fmt.Errorf("transaction timeout (received %d bytes in %d secs) %s", jsonbufLen, timeoutSecs, note.ErrCardIo+note.ErrTimeout)
			return
		}

	}

	// Done
	return
}

// OpenLease opens a remote card with a lease
func OpenLease(leaseScope string, leaseMins int) (context *Context, err error) {

	// Create the context structure
	context = &Context{}
	context.Debug = InitialDebugMode
	context.port = leaseScope
	context.portConfig = 0
	context.lastRequestSeqno = 0

	// Prevent accidental reservation for excessive durations e.g. 115200 minutes
	if leaseMins > 120 {
		err = fmt.Errorf("leasing a notecard has a 120 minute limit, but got %d", leaseMins)
		return
	}

	// Set up class functions
	context.CloseFn = leaseClose
	context.ReopenFn = leaseReopen
	context.TransactionFn = leaseTransaction
	context.traceOpenFn = leaseTraceOpen
	context.traceReadFn = leaseTraceRead
	context.traceWriteFn = leaseTraceWrite

	// Record serial configuration
	context.leaseScope = leaseScope
	if leaseMins == 0 {
		leaseMins = 1
	}
	leaseMins = (((leaseMins - 1) / reservationModulusMinutes) + 1) * reservationModulusMinutes
	context.leaseExpires = time.Now().Unix() + int64(leaseMins*60)

	// Open the port
	err = context.ReopenFn(context, context.portConfig)
	if err != nil {
		err = fmt.Errorf("error taking out lease: %s %s", err, note.ErrCardIo)
		return
	}

	// All set
	return
}

// Lock the appropriate mutex for the transaction
func lockTrans(multiport bool, portConfig int) {
	if multiport && portConfig >= 0 && portConfig < 128 {
		multiportTransLock[portConfig].Lock()
	} else {
		transLock.Lock()
	}
}

// Unlock the appropriate mutex for the transaction
func unlockTrans(multiport bool, portConfig int) {
	if multiport && portConfig >= 0 && portConfig < 128 {
		multiportTransLock[portConfig].Unlock()
	} else {
		transLock.Unlock()
	}
}

// Get the CallerID for this requestor, increasing the likelihood of getting the same
// reservation between tests which may be run across different machines and across
// different processes on the same machine.
func callerID() (id string) {

	// See if it's specified in the environment
	id = os.Getenv("NOTEFARM_CALLERID")
	if id != "" {
		return
	}

	user, err := user.Current()
	if user != nil && err == nil {
		id = user.Username
	}

	hostname, err := os.Hostname()
	if hostname != "" && err == nil {
		id += "@" + hostname
	}

	if id == "" {
		// Use the mac address if we have no other name
		interfaces, err := net.Interfaces()
		if err == nil {
			for _, i := range interfaces {
				if i.Flags&net.FlagUp != 0 && !bytes.Equal(i.HardwareAddr, nil) {
					// Don't use random as we have a real address
					id = i.HardwareAddr.String()
					break
				}
			}
		}
	}

	return
}

// Serial trace open
func serialTraceOpen(context *Context) (err error) {
	return
}

// Serial trace read function
func serialTraceRead(context *Context) (data []byte, err error) {

	// Exit if not open
	if !context.portIsOpen {
		return data, fmt.Errorf("port not open " + note.ErrCardIo)
	}

	// Do the read
	var length int
	buf := make([]byte, 2048)
	readBeganMs = int(time.Now().UnixNano() / 1000000)
	length, err = context.serialPort.Read(buf)
	readElapsedMs := int(time.Now().UnixNano()/1000000) - readBeganMs
	if false {
		fmt.Printf("mon: elapsed:%d len:%d err:%s '%s'\n", readElapsedMs, length, err, string(buf[:length]))
	}
	if readElapsedMs == 0 && length == 0 && err == io.EOF {
		// On Linux, hardware port failures come back simply as immediate EOF
		err = fmt.Errorf("hardware failure")
	}
	if readElapsedMs == 0 && length == 0 {
		// On Linux, sudden unplug comes back simply as immediate ''
		err = fmt.Errorf("hardware unplugged or rebooted probably")
	}
	if err != nil {
		if err == io.EOF {
			// Just a read timeout
			return data, nil
		}
		return data, fmt.Errorf("%s %s", err, note.ErrCardIo)
	}

	return buf[:length], nil
}

// Serial trace write function
func serialTraceWrite(context *Context, data []byte) {
	context.serialPort.Write(data)
}

// Add a crc to the JSON transaction
func crcAdd(reqJson []byte, seqno int) []byte {

	// Exit if invalid
	if len(reqJson) < 2 {
		return reqJson
	}

	// Extract any terminator present so it isn't included in the checksum
	reqJsonStr := string(reqJson)
	terminator := ""
	temp := strings.Split(reqJsonStr, "}")
	if len(temp) > 1 {
		terminator = temp[len(temp)-1]
		reqJsonStr = strings.Join(temp[0:len(temp)-1], "}") + "}"
	}

	// Compute the CRC of the JSON
	crc := crc32.ChecksumIEEE([]byte(reqJsonStr))

	// Strip the suffix and prepare for crc concatenation.  Note that
	// the decode side assumes that either a space or comma was added
	reqJsonStr = strings.TrimSuffix(reqJsonStr, "}")
	if !strings.Contains(reqJsonStr, ":") {
		reqJsonStr += " "
	} else {
		reqJsonStr += ","
	}

	// Append the CRC
	reqJsonStr += fmt.Sprintf("\"crc\":\"%04X:%08X\"}", seqno, crc)

	// Done
	return []byte(reqJsonStr + terminator)
}

// Test and remove CRC from transaction JSON
// Note that if a CRC field is not present in the JSON, it is considered
// a valid transaction because old Notecards do not have the code
// with which to calculate and piggyback a CRC field.  Note that the
// CRC is stripped from the input JSON regardless of whether or not
// there was an error.
func crcError(rspJson []byte, shouldBeSeqno int) (retJson []byte, err error) {

	// Exit silently if invalid
	if len(rspJson) < 2 {
		return rspJson, nil
	}

	// Extract any terminator present so it isn't included in the checksum
	rspJsonStr := string(rspJson)
	terminator := ""
	temp := strings.Split(rspJsonStr, "}")
	if len(temp) > 1 {
		terminator = temp[len(temp)-1]
		rspJsonStr = strings.Join(temp[0:len(temp)-1], "}") + "}"
	}

	// Minimum valid JSON is "{}" (2 bytes) and must end with a closing "}".  If
	// it's not there, it's ok because it could be an old notecard
	crcFieldLength := 22 // ,"crc":"SSSS:CCCCCCCC"
	if len(rspJsonStr) < crcFieldLength+2 || !strings.HasSuffix(rspJsonStr, "}") {
		return rspJson, nil
	}

	// Split the string into its json and non-json components
	t1 := strings.Split(rspJsonStr, " \"crc\":\"")
	if len(t1) != 2 {
		t1 = strings.Split(rspJsonStr, ",\"crc\":\"")
	}
	if len(t1) != 2 {
		return rspJson, nil
	}
	stripped := t1[0] + "}"
	t2 := strings.Split(t1[1], ":")
	if len(t2) != 2 {
		return rspJson, fmt.Errorf("badly formatted CRC seqno")
	}
	seqnoHex := t2[0]
	seqno, err := strconv.ParseInt(seqnoHex, 16, 64)
	if err != nil {
		return rspJson, fmt.Errorf("badly formatted hex CRC seqno")
	}
	shouldBeCrcHex := strings.TrimSuffix(t2[1], "\"}")
	shouldBeCrc, err := strconv.ParseInt(shouldBeCrcHex, 16, 64)
	if err != nil {
		return rspJson, fmt.Errorf("badly formatted hex CRC")
	}

	// Compute the CRC of the JSON
	crc := crc32.ChecksumIEEE([]byte(stripped))

	// Test values
	if shouldBeSeqno != int(seqno) {
		return rspJson, fmt.Errorf("sequence number mismatch (%d != %d)", seqno, shouldBeSeqno)
	}
	if uint32(shouldBeCrc) != crc {
		return rspJson, fmt.Errorf("CRC mismatch")
	}

	// Done
	return []byte(stripped + terminator), nil
}
