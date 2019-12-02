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

// TraceCapture monitors the trace output for a certain period of time
// quiescentSecs says "if nonzero, return after N secs of no output activity"
// maximumSecs says "if nonzero, return after N secs even if output activity is continuing"
// It then returns the received output to the caller.
func (context *Context) TraceCapture(toSend []byte, quiescentSecs int, maximumSecs int) (captured []byte, err error) {

	// Tracing only works for USB and AUX ports
	if !context.isSerial {
		err = fmt.Errorf("tracing is only available on USB and AUX ports")
		return
	}

	// Send the string, if supplied
	if len(toSend) > 0 {
		_, err = context.openSerialPort.Write(append(toSend, []byte("\n")...))
		if err != nil {
			return
		}
	}

	// Loop, echoing to the console
	timeStarted := time.Now().Unix()
	timeOutput := time.Now().Unix()
	for {

		now := time.Now().Unix()
		if quiescentSecs > 0 && now >= timeOutput+int64(quiescentSecs) {
			return 
		}
		if maximumSecs > 0 && now >= timeStarted+int64(maximumSecs) {
			return 
		}

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
				continue
			}
			break
		}
		captured = append(captured, buf[:length]...)
		timeOutput = time.Now().Unix()
	}

	return

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
		if quiescentSecs > 0 && now >= timeOutput+int64(quiescentSecs) {
			return nil
		}
		if maximumSecs > 0 && now >= timeStarted+int64(maximumSecs) {
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
