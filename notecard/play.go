// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/blues/note-go/note"
)

// Interactive I/O
var iInputHandlerActive = false
var iWatch = false
var uiLock sync.RWMutex

// Interactive enters interactive request/response mode, disabling trace in case
// that was the last mode entered
func (context *Context) Interactive(watch bool, watchLevel int, prompt bool, watchCommand string, quitCommand string) (err error) {
	var rsp Request
	var colWidth int
	var cols int
	var subsystem []string
	var subsystemDisplayName []string

	// Set the watch on/off based upon whether there is a command
	iWatch = watch

	// Initialize for watching
	linesDisplayed := 0

	// Get the template for the trace log results.  We need to get this regardless of whether watch
	// is initially on because it might be turned on later
	rsp, err = context.TransactionRequest(Request{Req: "note.get", NotefileID: "_synclog.qi", Start: true})
	if err != nil {
		return
	}
	for _, entry := range strings.Split(rsp.Status, ",") {
		str := strings.Split(entry, ":")
		if len(str) >= 2 {
			cols++
			subsystem = append(subsystem, str[0])
			subsystemDisplayName = append(subsystemDisplayName, str[1])
			if len(str[1]) > colWidth {
				colWidth = len(str[1])
			}
		}
	}
	colWidth += 4

	if iWatch {

		// Print an opening banner if necessary
		now := time.Now().Local().Format("03:04:05 PM MST")
		rsp, err = context.TransactionRequest(Request{Req: "note.get", NotefileID: "_synclog.qi"})
		if err == nil && rsp.Body == nil {
			fmt.Printf("%s waiting for sync activity\n", now)
		}

	}

	// Now that we know we can speak to the notecard, spawn the input handlers
	if !iInputHandlerActive {
		go interactiveInputHandler(context, prompt, watchCommand, quitCommand)
		for !iInputHandlerActive {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Loop, printing data
	prevTimeSecs := int64(0)
	for err == nil || note.ErrorContains(err, note.ErrNoteNoExist) {

		// Exit if the handler exited
		if !iInputHandlerActive {
			err = nil
			break
		}

		// Loop if not watching
		if !iWatch {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Get the next entry
		rsp, err = context.TransactionRequest(Request{Req: "note.get", NotefileID: "_synclog.qi", Delete: true})
		if err != nil {
			if !note.ErrorContains(err, note.ErrNoteNoExist) {
				uiLock.Lock()
				fmt.Printf("\r%s\n", err)
				uiLock.Unlock()
			}
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		if rsp.Body == nil {
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		var bodyJSON []byte
		bodyJSON, err = note.ObjectToJSON(rsp.Body)
		if err != nil {
			break
		}
		var body SyncLogBody
		err = note.JSONUnmarshal(bodyJSON, &body)
		if err != nil {
			break
		}
		if body.DetailLevel > uint32(watchLevel) {
			continue
		}

		// Lock output for a moment
		uiLock.Lock()

		// Output a header if it will help readability
		if linesDisplayed%250 == 0 {
			fmt.Printf("\n%s ", strings.Repeat(" ", len(time.Now().Local().Format("03:04:05 PM MST"))))
			for i := 0; i < cols; i++ {
				fmt.Printf("%s%s",
					subsystemDisplayName[i],
					strings.Repeat(" ", colWidth-len(subsystemDisplayName[i])))
			}
			fmt.Printf("\n\n")
		} else {
			// Output a spacer if there is a distance in time
			if body.TimeSecs != 0 && body.TimeSecs > prevTimeSecs+30 {
				fmt.Printf("\n")
			}
		}
		linesDisplayed++

		// Display either the time OR the 'secs since boot' if time isn't available
		prevTimeSecs = body.TimeSecs
		timebuf := time.Unix(int64(body.TimeSecs), 0).Local().Format("03:04:05 PM MST")
		if body.TimeSecs == 0 {
			str := fmt.Sprintf("%d", body.BootMs)
			timebuf = fmt.Sprintf("%s%s", str, strings.Repeat(" ", len(timebuf)-len(str)))
		}

		// Display indentation
		fmt.Printf("\r%s ", timebuf)
		indentstr := "." + strings.Repeat(" ", colWidth-1)
		for _, ss := range subsystem {
			if ss == body.Subsystem {
				break
			}
			fmt.Printf("%s", indentstr)
		}

		// Display the message
		if watchLevel < SyncLogLevelProg {
			fmt.Printf("%s\n", note.ErrorClean(fmt.Errorf(body.Text)))
		} else {
			fmt.Printf("%s\n", body.Text)
		}

		// Release the UI
		uiLock.Unlock()

	}

	// Done
	iInputHandlerActive = false
	return

}

// Watch for console input
func interactiveInputHandler(context *Context, prompt bool, watchCommand string, quitCommand string) {

	// Mark as active, in case we invoke this multiple times
	iInputHandlerActive = true

	// Create a scanner to watch stdin
	scanner := bufio.NewScanner(os.Stdin)
	var message string

	// Send the command to the module
	for iInputHandlerActive {
		if prompt {
			uiLock.Lock()
			fmt.Printf("> ")
			uiLock.Unlock()
		}
		scanner.Scan()
		message = scanner.Text()
		if message == quitCommand {
			iInputHandlerActive = false
			break
		}
		if message == "" {
			continue
		}
		uiLock.Lock()
		if watchCommand != "" && message == watchCommand {
			if iWatch {
				iWatch = false
				fmt.Printf("watch off\n")
			} else {
				iWatch = true
				fmt.Printf("watch ON\n")
			}
			uiLock.Unlock()
			continue
		}
		rspJSON, err := context.TransactionJSON([]byte(message))
		if err != nil {
			fmt.Printf("error: %s\n", err)
		} else {
			fmt.Printf("%s", rspJSON)
		}
		uiLock.Unlock()
	}

}
