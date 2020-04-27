// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
)

// Watch the notecard until sync is completed
func watch(detailLevel int, timeoutSecs int) {

	var err error
	var rsp notecard.Request
	var colWidth int
	var cols int
	var subsystem []string
	var subsystemDisplayName []string

	// Turn off Notecard library debug output regardless of prior user preference
	card.DebugOutput(false, false)

	// Turn off tracing because it can interfere with our rapid transaction I/O
	card.TransactionRequest(notecard.Request{Req: "card.io", Mode: "trace-off"})

	// Get the template for the trace log results
	rsp, err = card.TransactionRequest(notecard.Request{Req: "note.get", NotefileID: "_synclog.qi", Start: true})
	if err == nil {
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
	}

	// Align into columns
	colWidth += 4
	now := time.Now().Local().Format("03:04:05 PM MST")

	// Loop, printing data
	linesDisplayed := 0
	prevTimeSecs := int64(0)
	expireTime := time.Now().Add(time.Duration(timeoutSecs) * time.Second)
	for err == nil && time.Now().Before(expireTime) {

		// Delay so that we aren't hammering the notecard
		time.Sleep(750 * time.Millisecond)

		// Get the next entry
		rsp, err = card.TransactionRequest(notecard.Request{Req: "note.get", NotefileID: "_synclog.qi", Delete: true})
		if err != nil {
			if note.ErrorContains(err, "invalid character") {
				fmt.Printf("WARNING: can't enter commands in a different trace window while 'watching': %s\n", err)
			}
			err = nil
			continue
		}
		if rsp.Body == nil {
			continue
		}
		var bodyJSON []byte
		bodyJSON, err = note.ObjectToJSON(rsp.Body)
		if err != nil {
			break
		}
		var body notecard.SyncLogBody
		err = note.JSONUnmarshal(bodyJSON, &body)
		if err != nil {
			break
		}
		if body.DetailLevel > uint32(detailLevel) {
			continue
		}

		// Output a header if it will help readability
		if linesDisplayed%250 == 0 {
			fmt.Printf("\n%s ", strings.Repeat(" ", len(now)))
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
		fmt.Printf("%s ", timebuf)
		indentstr := "." + strings.Repeat(" ", colWidth-1)
		for _, ss := range subsystem {
			if ss == body.Subsystem {
				break
			}
			fmt.Printf("%s", indentstr)
		}

		// Display the message
		if body.DetailLevel == notecard.SyncLogLevelMajor {
			fmt.Printf("%s\n", note.ErrorClean(fmt.Errorf(body.Text)))
		} else {
			fmt.Printf("%s\n", body.Text)
		}

		// Remember when
		expireTime = time.Now().Add(time.Duration(timeoutSecs) * time.Second)

	}

}
