// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// The time when the last read began
var (
	readBeganMs        = 0
	inputHandlerActive = false
)

// Trace the incoming serial output AND connect the input handler
func (context *Context) Trace() (err error) {

	// Tracing only works for USB and AUX ports
	if context.traceOpenFn == nil {
		return fmt.Errorf("tracing is not available on this port")
	}

	// Ensure that we have a reservation
	err = context.ReopenIfRequired(context.portConfig)
	if err != nil {
		return err
	}

	// Open the trace port
	err = context.traceOpenFn(context)
	if err != nil {
		cardReportError(context, err)
		return
	}

	// Spawn the input handler
	if !inputHandlerActive {
		go inputHandler(context)
	}

	// Loop, echoing to the console
	for {

		buf, err := context.traceReadFn(context)
		if err != nil {
			cardReportError(context, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(buf) > 0 {
			fmt.Printf("%s", buf)
		}

	}

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
			if !context.portIsOpen {
				for _, r := range message[1:] {
					switch {
					// 'a' - 'z'
					case 97 <= r && r <= 122:
						ba := make([]byte, 1)
						ba[0] = byte(r - 96)
						context.traceWriteFn(context, ba)
						// 'A' - 'Z'
					case 65 <= r && r <= 90:
						ba := make([]byte, 1)
						ba[0] = byte(r - 64)
						context.traceWriteFn(context, ba)
					}
				}
			}
		} else {
			// Send the command to the module
			if !context.portIsOpen {
				time.Sleep(250 * time.Millisecond)
			} else {
				context.traceWriteFn(context, []byte(message))
				context.traceWriteFn(context, []byte("\n"))
			}
		}
	}
}
