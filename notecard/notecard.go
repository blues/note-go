// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/blues/serial"
	"io"
	"os"
	"strings"
	"time"
)

// Module communication interfaces
const (
	NotecardInterfaceSerial = "serial"
	NotecardInterfaceI2C	= "i2c"
)

// CardI2CMax controls chunk size that's socially appropriate on the I2C bus.
// It must be 1-255 bytes as per spec
const CardI2CMax = 255

// The notecard is a real-time device that has a fixed size interrupt buffer.  We can push data
// at it far, far faster than it can process it, therefore we push it in segments with a pause
// between each segment.

// CardRequestSegmentMaxLen (golint)
const CardRequestSegmentMaxLen = 1000

// CardRequestSegmentDelayMs (golint)
const CardRequestSegmentDelayMs = 250

// Context for the port that is open
type Context struct {

	// True to emit trace output
	Debug			bool

	// Class functions
	PortEnumFn	   func() (ports []string, err error)
	PortDefaultsFn func() (port string, portConfig int)
	CloseFn		   func(context *Context)
	ReopenFn	   func(context *Context) (err error)
	ResetFn		   func(context *Context) (err error)
	TransactionFn  func(context *Context, reqJSON []byte) (rspJSON []byte, err error)

	// Serial instance state
	isSerial	   bool
	openSerialPort *serial.Port
	serialConfig 	serial.Config
	i2cName		string
	i2cAddress	int
}

// Report a critical card error
func cardReportError(context *Context, err error) {
	if context.Debug {
		fmt.Printf("***\n")
		fmt.Printf("*** %s\n", err)
		fmt.Printf("***\n")
		time.Sleep(10 * time.Second)
	}
}

// DebugOutput enables/disables debug output
func (context *Context) DebugOutput(enabled bool) {
	context.Debug = enabled
}

// EnumPorts returns the list of all available ports on the specified interface
func (context *Context) EnumPorts() (ports []string, err error) {
	return context.PortEnumFn()
}

// PortDefaults gets the defaults for the specified port
func (context *Context) PortDefaults() (port string, portConfig int) {
	return context.PortDefaultsFn()
}

// Identify this Notecard connection
func (context *Context) Identify() (protocol string, port string, portConfig int) {
	if context.isSerial {
		return "serial", context.serialConfig.Name, context.serialConfig.Baud
	} else {
		return "I2C", context.i2cName, context.i2cAddress
	}
}

// Open the card to establish communications
func Open(moduleInterface string, port string, portConfig int) (context Context, err error) {

	switch moduleInterface {
	case NotecardInterfaceSerial:
		context, err = OpenSerial(port, portConfig)
		break
	case NotecardInterfaceI2C:
		context, err = OpenI2C(port, portConfig)
		break
	default:
		err = fmt.Errorf("unknown interface: %s", moduleInterface)
		break
	}
	if err != nil {
		err = fmt.Errorf("error opening port: %s", err)
		return
	}

	return

}

// Reset serial to a known state
func cardResetSerial(context *Context) (err error) {

	// In order to ensure that we're not getting the reply to a failed
	// transaction from a prior session, drain any pending input prior
	// to transmitting a command.  Note that we use this technique of
	// looking for a known reply to \n\n, rather than just "draining
	// anything pending on serial", because the nature of read() is
	// that it blocks (until timeout) if there's nothing available.
	var length int
	buf := make([]byte, 2048)
	for {
		_, err = context.openSerialPort.Write([]byte("\n\n"))
		if err != nil {
			err = fmt.Errorf("error transmitting to module: %s", err)
			cardReportError(context, err)
			return
		}
		time.Sleep(500 * time.Millisecond)
		readBeganMs := int(time.Now().UnixNano() / 1000000)
		length, err = context.openSerialPort.Read(buf)
		readElapsedMs := int(time.Now().UnixNano()/1000000) - readBeganMs
		if readElapsedMs == 0 && length == 0 && err == io.EOF {
			// On Linux, hardware port failures come back simply as immediate EOF
			err = fmt.Errorf("hardware failure")
		}
		if err != nil {
			err = fmt.Errorf("error reading from module: %s", err)
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

	// Set up class functions
	context.PortEnumFn = serialPortEnum
	context.PortDefaultsFn = serialDefault
	context.CloseFn = cardCloseSerial
	context.ReopenFn = cardReopenSerial
	context.ResetFn = cardResetSerial
	context.TransactionFn = cardTransactionSerial

	// Record serial configuration
	context.isSerial = true
	context.serialConfig.Name = port
	context.serialConfig.Baud = portConfig
	context.serialConfig.ReadTimeout = time.Millisecond * 500

	// Open the serial port
	err = cardReopenSerial(&context)
	if err != nil {
		err = fmt.Errorf("error opening serial port %s at %d: %s", port, portConfig, err)
		return
	}

	// Reset serial to a known good state
	err = cardResetSerial(&context)

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
			err = fmt.Errorf("error reading chunk: %s", err2)
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
	err = i2cOpen(uint8(portConfig), port, portConfig)
	if err != nil {
		if false {
			ports, _ := I2CPorts()
			fmt.Printf("Available ports: %v\n", ports)
		}
		err = fmt.Errorf("i2c init error: %s", err)
		return
	}

	// Reset it to a known good state
	err = cardResetI2C(&context)

	// Done
	return

}

// Reset the port
func (context *Context) Reset() (err error) {
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



// Reopen the port
func (context *Context) Reopen() (err error) {
	return context.ReopenFn(context)
}

// Reopen serial
func cardReopenSerial(context *Context) (err error) {
	if context.openSerialPort != nil {
		return fmt.Errorf("error serial port is already open")
	}

	// Open the serial port
	context.openSerialPort, err = serial.OpenPort(&context.serialConfig)
	if err != nil {
		return fmt.Errorf("error opening serial port %s at %d: %s", context.serialConfig.Name, context.serialConfig.Baud, err)
	}

	// Reset serial to a known good state
	return cardResetSerial(context)
}

// Reopen I2C
func cardReopenI2C(context *Context) (err error) {
	return fmt.Errorf("error i2c reopen not yet supported since I can't test it yet")
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
func SerialPorts() (ports []string, err error) {
	return serialPortEnum()
}

// I2CPorts returns the list of available I2C ports
func I2CPorts() (ports []string, err error) {
	return i2cPortEnum()
}

// TraceOutput monitors the trace output for a certain period of time
// quiescentSecs says "if nonzero, return after N secs of no output activity"
// maximumSecs says "if nonzero, return after N secs even if output activity is continuing"
func (context *Context) TraceOutput(quiescentSecs int, maximumSecs int) (err error) {

	// Tracing only works for USB and AUX ports
	if !context.isSerial {
		return fmt.Errorf("tracing is only available on USB and AUX ports")
	}

	// Turn on tracing on the current port
	req := Request{Req: ReqCardIO}
	req.Mode = "trace-on"
	context.TransactionRequest(req)

	// Loop, echoing to the console
	timeStarted := time.Now().Unix()
	timeOutput := time.Now().Unix()
	for {

		now := time.Now().Unix()
		if quiescentSecs > 0 && now >= timeOutput + int64(quiescentSecs) {
			return nil
		}
		if maximumSecs > 0 && now >= timeStarted + int64(maximumSecs) {
			return nil
		}

		buf := make([]byte, 2048)
		readBeganMs := int(time.Now().UnixNano() / 1000000)
		length, err := context.openSerialPort.Read(buf)
		readElapsedMs := int(time.Now().UnixNano()/1000000) - readBeganMs

		if err == nil && length == 0 {
			// Nothing to read yet
			// Sleep briefly to be polite yet responsive
			time.Sleep(1 * time.Millisecond)
			continue
		}

		if false {
			fmt.Printf("mon: elapsed:%d len:%d err:%s '%s'\n", readElapsedMs, length, err, string(buf[:length]))
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
			break
		}
		fmt.Printf("%s", buf[:length])
		timeOutput = time.Now().Unix()
	}

	err = fmt.Errorf("error reading from module: %s", err)
	cardReportError(context, err)
	return

}

// Trace the incoming serial output AND connect the input handler
func (context *Context) Trace() (err error) {

	// Tracing only works for USB and AUX ports
	if !context.isSerial {
		return fmt.Errorf("tracing is only available on USB and AUX ports")
	}

	// Turn on tracing on the current port
	req := Request{Req: ReqCardIO}
	req.Mode = "trace-on"
	context.TransactionRequest(req)

	// Spawn the input handler
	go inputHandler(context)

	// Loop, echoing to the console
	for {
		var length int
		buf := make([]byte, 2048)
		readBeganMs := int(time.Now().UnixNano() / 1000000)
		length, err = context.openSerialPort.Read(buf)
		readElapsedMs := int(time.Now().UnixNano()/1000000) - readBeganMs
		if false {
			fmt.Printf("mon: elapsed:%d len:%d err:%s '%s'\n", readElapsedMs, length, err, string(buf[:length]))
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
			break
		}
		fmt.Printf("%s", buf[:length])
	}

	err = fmt.Errorf("error reading from module: %s", err)
	cardReportError(context, err)
	return

}

// Watch for console input
func inputHandler(context *Context) {

	// Create a scanner to watch stdin
	scanner := bufio.NewScanner(os.Stdin)
	var message string

	for {

		scanner.Scan()
		message = scanner.Text()

		if strings.HasPrefix(message, "^") {

			for _, r := range message[1:] {
				switch {
					// 'a' - 'z'
				case 97 <= r && r <= 122:
					ba := make([]byte, 1)
					ba[0] = byte(r - 96)
					context.openSerialPort.Write(ba)
					// 'A' - 'Z'
				case 65 <= r && r <= 90:
					ba := make([]byte, 1)
					ba[0] = byte(r - 64)
					context.openSerialPort.Write(ba)
				}
			}
		} else {
			context.openSerialPort.Write([]byte(message))
			context.openSerialPort.Write([]byte("\n"))
		}
	}

}

// TransactionRequest performs a card transaction with a Req structure
func (context *Context) TransactionRequest(req Request) (rsp Request, err error) {

	// Marshal the request to JSON
	reqJSON, err2 := json.Marshal(req)
	if err2 != nil {
		err = fmt.Errorf("error marshaling request for module: %s", err2)
		return
	}

	// Perform the transaction
	rspJSON, err2 := context.TransactionJSON(reqJSON)
	if err2 != nil {
		err = fmt.Errorf("error marshaling request for module: %s", err2)
		return
	}

	// Unmarshal for convenience of the caller
	err = json.Unmarshal(rspJSON, &rsp)
	if err != nil {
		err = fmt.Errorf("error unmarshaling reply from module: %s", err)
		return
	}

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

// NewBody creates a new body.	Note that this method is provided
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

// Transaction performs a card transaction with a JSON structure
func (context *Context) Transaction(req map[string]interface{}) (rsp map[string]interface{}, err error) {

	// Marshal the request to JSON
	reqJSON, err2 := json.Marshal(req)
	if err2 != nil {
		err = fmt.Errorf("error marshaling request for module: %s", err2)
		return
	}

	// Perform the transaction
	rspJSON, err2 := context.TransactionJSON(reqJSON)
	if err2 != nil {
		err = fmt.Errorf("error marshaling request for module: %s", err2)
		return
	}

	// Unmarshal for convenience of the caller
	err = json.Unmarshal(rspJSON, &rsp)
	if err != nil {
		err = fmt.Errorf("error unmarshaling reply from module: %s", err)
		return
	}

	// Done
	return
}

// TransactionJSON performs a card transaction using raw JSON []bytes
func (context *Context) TransactionJSON(reqJSON []byte) (rspJSON []byte, err error) {

	if context.Debug {
		fmt.Printf("%s\n", reqJSON)
	}

	// Make sure that the JSON has a terminator
	reqJSON = []byte(string(reqJSON) + "\n")

	// Perform the transaction
	rspJSON, err = context.TransactionFn(context, reqJSON)
	if err != nil {
		context.ResetFn(context)
	}

	if context.Debug {
		fmt.Printf("%s", string(rspJSON))
	}
	return

}

// Perform a card transaction over serial under the assumption that request already has '\n' terminator
func cardTransactionSerial(context *Context, reqJSON []byte) (rspJSON []byte, err error) {

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
			err = fmt.Errorf("error transmitting to module: %s", err)
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

	// Read the reply until we get '\n' at the end
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
			err = fmt.Errorf("error reading from module: %s", err)
			cardReportError(context, err)
			return
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
			err = fmt.Errorf("chunk write error: %s", err)
			return
		}
		chunkoffset += chunklen
		jsonbufLen -= chunklen
		sentInSegment += chunklen
		if sentInSegment > CardRequestSegmentMaxLen {
			sentInSegment -= CardRequestSegmentMaxLen
			time.Sleep(CardRequestSegmentDelayMs * time.Millisecond)
		}
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
			err = fmt.Errorf("error reading chunk: %s", err2)
			return
		}

		// Append to the JSON being accumulated
		rspJSON = append(rspJSON, readbuf...)
		readlen := len(readbuf)
		jsonbufLen += readlen

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

		// Delay, simply waiting for the Notecard to process the request
		time.Sleep(50 * time.Millisecond)

	}

	// Done
	return
}
