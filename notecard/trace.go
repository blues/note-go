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
)

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
			return
		}
	}

	// Loop, echoing to the console
	for {

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
			break
		}
		captured += string(buf[:length])
		if (toEnd != "" && strings.Contains(captured, toEnd)) {
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

	// Turn on tracing on the current port
	req := Request{Req: ReqCardIO}
	req.Mode = "trace-on"
	context.TransactionRequest(req)

	// Enter interactive mode
	return context.interactive()

}

// Enter interactive request/response mode, disabling trace in case
// that was the last mode entered
func (context *Context) Interactive() (err error) {

	// Interaction only works for USB and AUX ports
	if !context.isSerial {
		return fmt.Errorf("interaction mode is only available on USB and AUX ports")
	}

	// Turn on tracing on the current port
	req := Request{Req: ReqCardIO}
	req.Mode = "trace-off"
	context.TransactionRequest(req)

	// Enter interactive mode
	return context.interactive()

}

// Enter interactive request/response mode
func (context *Context) interactive() (err error) {

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
