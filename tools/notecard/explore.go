// Copyright 2021 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"fmt"
	"sort"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
)

// Explore the contents of this device
func explore(includeReserved bool) (err error) {

	// Get the list of notefiles
	req := notecard.Request{Req: notecard.ReqFileChanges}
	req.Allow = includeReserved
	var rsp notecard.Request
	rsp, err = card.TransactionRequest(req)
	if err != nil {
		return
	}

	// Exit if no notefiles
	if rsp.FileInfo == nil || len(*rsp.FileInfo) == 0 {
		fmt.Printf("no notefiles\n")
		return
	}

	// Sort the notefiles
	notefileIDs := []string{}
	for notefileID := range *rsp.FileInfo {
		notefileIDs = append(notefileIDs, notefileID)
	}
	sort.Strings(notefileIDs)

	// Iterate over each file
	for _, notefileID := range notefileIDs {

		fmt.Printf("    %s\n", notefileID)

		// Get the notes
		req = notecard.Request{Req: notecard.ReqNoteChanges}
		req.Allow = includeReserved
		req.Deleted = true
		req.NotefileID = notefileID
		rsp, err = card.TransactionRequest(req)
		if err != nil {
			return
		}

		// Exit if no notefiles
		if rsp.Notes == nil || len(*rsp.Notes) == 0 {
			continue
		}

		// Show the notes
		for noteID, n := range *rsp.Notes {
			fmt.Printf("        %s", noteID)
			if n.Deleted {
				fmt.Printf(" (DELETED)")
			}
			fmt.Printf("\n")
			if n.Body != nil {
				bodyJSON, err := note.JSONMarshal(*n.Body)
				if err == nil {
					fmt.Printf("            %s\n", string(bodyJSON))
				}
			}
			if n.Payload != nil {
				fmt.Printf("            Payload: %d bytes\n", len(*n.Payload))
			}
		}

	}

	// Done
	return

}
