// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
	"github.com/google/uuid"
)

// Send text or a JSON object to a list of addressees (email or path)
func messageSend(to string, sendASCII string, bodyJSON string) (err error) {

	// Create a new message
	var message note.Message
	message.UID = uuid.New().String()
	message.Sent = uint32(time.Now().Unix())

	// Update our own contact, which has a side-effect of updating our own device addresses
	ourContact, err := contactSet("", "", 0)
	if err != nil {
		return
	}
	message.From = ourContact

	// Break down the addressee list into its components, and fail if we can't find them
	message.To, err = contactExpand(ourContact, to)
	if err != nil {
		return
	}
	if len(message.To) == 0 {
		err = fmt.Errorf("no recipient address supplied")
		return
	}

	// Set the body
	message.ContentType = note.MessageContentASCII
	message.Content = sendASCII
	if len(bodyJSON) > 0 {
		var b map[string]interface{}
		b, err = note.JSONToBody([]byte(bodyJSON))
		if err != nil {
			return
		}
		message.Body = &b
	}

	// Convert the message to JSON
	var body map[string]interface{}
	body, err = note.ObjectToBody(message)
	if err != nil {
		return
	}

	// Send the message
	rsp := notecard.Request{}
	req := notecard.Request{Req: notecard.ReqNoteAdd}
	req.NotefileID = note.MessageOutbox
	req.Body = &body
	rsp, err = card.TransactionRequest(req)
	if err != nil {
		fmt.Printf("message enqueueing %s error: %s\n", req.Req, err)
		return
	}
	numStaged := rsp.Total

	// Convert the message to JSON that will be ready for the message store
	message.StoreTags = append(message.StoreTags, note.MessageSTagSent)
	body, err = note.ObjectToBody(message)
	if err != nil {
		return
	}

	// After being sent, keep a copy in the message store
	req = notecard.Request{Req: notecard.ReqNoteAdd}
	req.NotefileID = note.MessageStore
	req.Body = &body
	_, err = card.TransactionRequest(req)
	if err != nil {
		fmt.Printf("message store %s error: %s\n", req.Req, err)
		return
	}

	// Done
	fmt.Printf("message enqueued (%d staged for sync)\n", numStaged)
	return

}

type byMessageWhen []note.Info

func (a byMessageWhen) Len() int      { return len(a) }
func (a byMessageWhen) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byMessageWhen) Less(i, j int) bool {
	var im, jm note.Message
	note.BodyToObject(a[i].Body, &im)
	note.BodyToObject(a[j].Body, &jm)
	iMinWhen := im.Sent
	if iMinWhen == 0 || (im.Received != 0 && iMinWhen > im.Received) {
		iMinWhen = im.Received
	}
	jMinWhen := jm.Sent
	if jMinWhen == 0 || (jm.Received != 0 && jMinWhen > jm.Received) {
		jMinWhen = jm.Received
	}
	return iMinWhen < jMinWhen
}

// Print a separator line
func printSep() {
	fmt.Printf("------------------------------------------------------------------------\n")
}

// Print a word-wrapped message with a per-line prefix
func printWrapped(prefix string, width int, data string) {

	width += len(prefix)
	datalen := len(data)
	col := 0
	off := 0
	for {
		next := strings.IndexAny(data[off:], " ")
		thisWidth := datalen - off
		if next != -1 {
			for data[off+next:off+next+1] == " " {
				next++
			}
			thisWidth = next
		}
		if thisWidth == 0 {
			break
		}
		if col+thisWidth > width {
			fmt.Printf("\n")
			col = 0
		}
		if col == 0 {
			fmt.Printf(prefix)
			col += len(prefix)
		}
		fmt.Printf("%s", data[off:off+thisWidth])
		col += thisWidth
		off += thisWidth
	}
	if col != 0 {
		fmt.Printf("\n")
		col = 0
	}
}

// Show messages (show all if tail == 0)
func messageShow(tail int) (err error) {

	// Get our contact for address optimization
	ourContact, err := contactGet()

	// Get all notes in the contact database
	var notes []note.Info
	notes, err = getAllNotes(note.MessageStore)
	if err != nil {
		return
	}

	// Sort notes by date
	sort.Sort(byMessageWhen(notes))

	// Show them
	numNotes := len(notes)
	firstNote := 0
	if tail != 0 && tail < numNotes {
		firstNote = numNotes - tail
		fmt.Printf("%d/%d messages\n", numNotes-firstNote, numNotes)
	} else {
		fmt.Printf("%d messages\n", numNotes)
	}
	if firstNote >= numNotes {
		return
	}
	for i := firstNote; i < numNotes; i++ {

		// Fetch the message from the note
		if notes[i].Body == nil {
			continue
		}
		var m note.Message
		err = note.BodyToObject(notes[i].Body, &m)
		if err != nil {
			continue
		}

		// Print a separator
		printSep()

		// Display the message
		fmt.Printf("%-10s      From: %s\n", "#"+notes[i].NoteID, contactName(ourContact, m.From))
		if len(m.To) >= 1 {
			fmt.Printf("%-10s        To: %s\n", strings.Join(m.StoreTags, ","), contactName(ourContact, m.To[0]))
			for j := 1; j < len(m.To); j++ {
				fmt.Printf("%-10s			%s\n", strings.Join(m.Tags, ","), contactName(ourContact, m.To[j]))
			}
			if m.Sent != 0 {
				fmt.Printf("%-10s      Sent: %s\n", "",
					time.Unix(int64(m.Sent), 0).Local().Format("2006-01-02 15:04:05 MST"))
			}
			if m.Received != 0 {
				fmt.Printf("%-10s  Received: %s\n", "",
					time.Unix(int64(m.Received), 0).Local().Format("2006-01-02 15:04:05 MST"))
			}
			fmt.Printf("\n")
			switch m.ContentType {
			case note.MessageContentASCII:
				printWrapped("                      ", 40, m.Content)
			}
		}

	}

	// Done
	printSep()
	return
}

// Delete a message or messages
func messageDelete(noteID string) (err error) {
	noteID = strings.Replace(noteID, ",", ";", -1)
	noteIDList := strings.Split(noteID, ";")
	for _, id := range noteIDList {
		if strings.HasPrefix(id, "#") {
			id = strings.TrimPrefix(id, "#")
		}
		req := notecard.Request{Req: notecard.ReqNoteDelete}
		req.NotefileID = note.MessageStore
		req.NoteID = id
		_, err = card.TransactionRequest(req)
		if err != nil {
			return
		}
	}
	return
}

// Delete all but the N most recent messages
func messageDeleteTail(tail int) (err error) {

	// Get all notes in the contact database
	var notes []note.Info
	notes, err = getAllNotes(note.MessageStore)
	if err != nil {
		return
	}

	// Sort notes by date
	sort.Sort(byMessageWhen(notes))

	// Delete them
	numNotes := len(notes)
	firstNoteToSave := 0
	if tail != 0 && tail < numNotes {
		firstNoteToSave = numNotes - tail
	}
	for i := 0; i < firstNoteToSave; i++ {
		req := notecard.Request{Req: notecard.ReqNoteDelete}
		req.NotefileID = note.MessageStore
		req.NoteID = notes[i].NoteID
		_, err = card.TransactionRequest(req)
		if err != nil {
			return
		}
		fmt.Printf("#%s deleted\n", req.NoteID)
	}

	// Done
	return
}

// Get the addresees from a message in the inbox
func messageContacts(ourContact note.MessageContact, noteID string) (contacts []note.MessageContact, err error) {

	// Get the message
	req := notecard.Request{Req: notecard.ReqNoteGet}
	req.NotefileID = note.MessageStore
	req.NoteID = strings.TrimPrefix(noteID, "#")
	var rsp notecard.Request
	rsp, err = card.TransactionRequest(req)
	if err != nil {
		if note.ErrorContains(err, note.ErrNoteNoExist) {
			err = fmt.Errorf("reply-to message not found: %s", noteID)
		}
		return
	}

	// Get the message body
	if rsp.Body == nil {
		err = fmt.Errorf("reply-to message is invalid")
		return
	}
	var m note.Message
	err = note.BodyToObject(rsp.Body, &m)
	if err != nil {
		err = fmt.Errorf("reply-to note is not a message")
		return
	}

	// Append the From contact
	contacts = append(contacts, m.From)

	// Append all the To contacts other than ourselves
	for _, c := range m.To {
		same, _, _, _ := contactNewer(ourContact, c)
		if !same {
			contacts = append(contacts, c)
		}
	}

	// Done
	return

}

// Migrate messages from the QI to the message store
func messageMigrate() (err error) {

	// Loop, getting notes from the inbound queue
	numMoved := 0
	for {

		// Dequeue
		req := notecard.Request{Req: notecard.ReqNoteGet}
		req.NotefileID = note.MessageInbox
		req.Delete = true
		var rsp notecard.Request
		rsp, err = card.TransactionRequest(req)
		if err != nil {
			if note.ErrorContains(err, note.ErrNoteNoExist) {
				err = nil
			}
			break
		}

		// Get the message
		if rsp.Body == nil {
			continue
		}
		var m note.Message
		err = note.BodyToObject(rsp.Body, &m)
		if err != nil {
			break
		}

		// Set the store tag as "received"
		m.StoreTags = append([]string{}, note.MessageSTagReceived)
		var body map[string]interface{}
		body, err = note.ObjectToBody(m)
		if err != nil {
			break
		}

		// Move it into the message store
		req = notecard.Request{Req: notecard.ReqNoteAdd}
		req.NotefileID = note.MessageStore
		req.Body = &body
		_, err = card.TransactionRequest(req)
		if err != nil {
			fmt.Printf("migration: message store %s error: %s\n", req.Req, err)
		}
		numMoved++

		// Rummage through the message's contacts, adding them to our contact list
		var contacts []note.MessageContact
		contacts = append(contacts, m.From)
		contacts = append(contacts, m.To...)
		err = contactListUpdate(contacts)
		if err != nil {
			fmt.Printf("updating contact list: %s\n", err)
		}

	}

	// Done
	if numMoved > 0 {
		fmt.Printf("%d messages migrated from inbox\n", numMoved)
	}
	return

}

// Validate that the outbox exists and that it has the correct access privs for anonymous deposit
func messageValidateInbox() (err error) {

	// Form a NotefileInfo structure with the correct info
	info := note.NotefileInfo{}
	info.AnonAddAllowed = true

	// Store the NotefileInfo into a map that is indexed by the notefile name
	req := notecard.Request{Req: notecard.ReqFileAdd}
	fileInfoMap := map[string]note.NotefileInfo{}
	fileInfoMap[note.MessageInbox] = info
	req.FileInfo = &fileInfoMap

	// Perform the request
	_, err = card.TransactionRequest(req)
	if err != nil {
		fmt.Printf("message inbox %s error: %s\n", req.Req, err)
	}

	// Done
	return

}
