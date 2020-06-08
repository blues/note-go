// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/blues/note-go/note"
)

// The time when the last read began
var readBeganMs = 0
var promptedMs = 0
var prompted = false
var inputHandlerActive = false
var promptHandlerActive = false

// TraceCapture monitors the trace output until a delimiter is reached
// It then returns the received output to the caller.
func (context *Context) TraceCapture(toSend string, toEnd string) (captured string, err error) {

	// Tracing only works for USB and AUX ports
	if !context.isSerial {
		err = fmt.Errorf("tracing is only available on USB and AUX ports")
		return
	}

	// Send the string, if supplied
	if len(toSend) > 0 {
		_, err = context.openSerialPort.Write(append([]byte(toSend), []byte("\n")...))
		if err != nil {
			err = fmt.Errorf("%s %s", err, note.ErrCardIo)
			return
		}
	}

	// Loop, echoing to the console
	for {

		// Reopen if error
		if cardResetOnNextRequest {
			err = context.Reopen()
			if err != nil {
				continue
			}
		}

		// Read from the card
		buf := make([]byte, 2048)
		readBeganMs := int(time.Now().UnixNano() / 1000000)
		var length int
		length, err = context.openSerialPort.Read(buf)
		readElapsedMs := int(time.Now().UnixNano()/1000000) - readBeganMs

		if err == nil && length == 0 {
			// Nothing to read yet
			// Sleep briefly to be polite yet responsive
			time.Sleep(1 * time.Millisecond)
			continue
		}

		if readElapsedMs == 0 && length == 0 && err == io.EOF {
			// On Linux, hardware port failures come back simply as immediate EOF
			err = fmt.Errorf("hardware failure")
		}
		if err != nil {
			if err == io.EOF {
				// Just a read timeout
				err = nil
				continue
			}
			err = fmt.Errorf("%s %s", err, note.ErrCardIo)
			break
		}
		captured += string(buf[:length])
		if toEnd != "" && strings.Contains(captured, toEnd) {
			break
		}
	}

	return

}

// Trace the incoming serial output AND connect the input handler
func (context *Context) Trace() (err error) {

	// Tracing only works for USB and AUX ports
	if !context.isSerial {
		return fmt.Errorf("tracing is only available on USB and AUX ports")
	}

	// Exit if not open
	if context.openSerialPort == nil {
		err = fmt.Errorf("port not open " + note.ErrCardIo)
		cardReportError(context, err)
		return
	}

	// Turn on tracing on the current port
	debugWas := context.Debug
	context.Debug = false
	req := Request{Req: ReqCardIO}
	req.Mode = "trace-on"
	context.TransactionRequest(req)
	context.Debug = debugWas

	// Spawn the input handler
	if !inputHandlerActive {
		go inputHandler(context)
	}
	if !promptHandlerActive {
		go promptHandler(context)
	}

	// Loop, echoing to the console
	for {

		// Pause if not open
		if context.openSerialPort == nil {
			err = fmt.Errorf("port not open " + note.ErrCardIo)
			cardReportError(context, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Do the read
		var length int
		buf := make([]byte, 2048)
		readBeganMs = int(time.Now().UnixNano() / 1000000)
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
			err = fmt.Errorf("%s %s", err, note.ErrCardIo)
			break
		}

		// Overwrite prompt
		if prompted {
			prompted = false
			fmt.Printf("\r")
		}

		// Echo
		text := string(buf[:length])
		if text != "\n" && text != "\r\n" {
			fmt.Printf("%s", text)
		}

	}

	err = fmt.Errorf("error reading from module: %s", err)
	cardReportError(context, err)
	return

}

// Watch for console input
func inputHandler(context *Context) {

	// Mark as active, in case we invoke this multiple times
	inputHandlerActive = true

	// Create a scanner to watch stdin
	scanner := bufio.NewScanner(os.Stdin)
	var message string

	for {

		scanner.Scan()
		message = scanner.Text()

		if strings.HasPrefix(message, "^") {

			if context.openSerialPort != nil {
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
			}

		} else {

			// Send the command to the module
			if context.openSerialPort == nil {
				time.Sleep(250 * time.Millisecond)
			} else {
				context.openSerialPort.Write([]byte(message))
				context.openSerialPort.Write([]byte("\n"))
			}

		}
	}

}

// Display a prompt
func promptHandler(context *Context) {

	// Mark as active, in case we invoke this multiple times
	promptHandlerActive = true

	// Loop, prompting whenever a read is pending for a period of time
	for {
		if readBeganMs != promptedMs {
			nowMs := int(time.Now().UnixNano() / 1000000)
			if readBeganMs == 0 || nowMs > readBeganMs+500 {
				promptedMs = readBeganMs
				prompted = true
				fmt.Printf("> ")
			}
		} else {
			time.Sleep(150 * time.Millisecond)
		}
	}

}
