// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.
// Derived from the "FeedSync" sample code which was covered
// by the Microsoft Public License (Ms-Pl) of December 3, 2007.
// FeedSync itself was derived from the algorithms underlying
// Lotus Notes replication, developed by Ray Ozzie et al c.1985.

package note

import (
    "fmt"
    "time"
    "encoding/json"
)

// Note is the most fundamental data structure, containing
// user data referred to as its "body" and its "payload".  All
// access to these fields, and changes to these fields, must
// be done indirectly through the note API.
type Note struct {
    Body interface{}            `json:"b,omitempty"`
    Payload []byte              `json:"p,omitempty"`
    Change int64                `json:"c,omitempty"`
    Histories *[]NoteHistory    `json:"h,omitempty"`
    Conflicts *[]Note           `json:"x,omitempty"`
    Updates int32               `json:"u,omitempty"`
    Deleted bool                `json:"d,omitempty"`
    Sent bool                   `json:"s,omitempty"`
    Bulk bool                   `json:"k,omitempty"`
}

// NoteHistory records the update history, optimized so that if the most recent entry
// is by the same endpoint as an update/delete, that entry is re-used.  The primary use
// of NoteHistory is for conflict detection, and you don't need to detect conflicts
// against yourself.
type NoteHistory struct {
    When int64                  `json:"w,omitempty"`
    Where string				`json:"l,omitempty"`
    EndpointID string           `json:"e,omitempty"`
    Sequence int32              `json:"s,omitempty"`
}

// NoteInfo is the info returned on a per-note basis on requests
type NoteInfo struct {
    Body *interface{}           `json:"body,omitempty"`
    Payload *[]byte             `json:"payload,omitempty"`
    Deleted bool                `json:"deleted,omitempty"`
}

// CreateNote creates the core data structure for an object
func CreateNote(body []byte, payload []byte) (newNote Note, err error) {
    newNote.Payload = payload
    err = newNote.SetBody(body)
    return
}

// SetBody sets the application-supplied Body field of a given Note
func (note *Note) SetBody(body []byte) (err error) {
    if body == nil {
        note.Body = nil
    } else {
        err := json.Unmarshal(body, &note.Body)
        if err != nil {
            note.Body = nil
            return fmt.Errorf("cannot set body: invalid JSON: %s", err)
        }
    }
    return;
}

// SetPayload sets the application-supplied Payload field of a given Note,
// which must be binary bytes that will ultimately be rendered as base64 in JSON
func (note *Note) SetPayload(payload []byte) {
    note.Payload = payload
}

// Close closes and frees the object on a note {
func (note *Note) Close() {
}

// Dup duplicates the note
func (note *Note) Dup() Note {
    newNote := *note
    return newNote
}

// GetBody retrieves the application-specific Body of a given Note
func (note *Note) GetBody() []byte {
    data, err := json.Marshal(note.Body)
    if err != nil {
        data = []byte{}
    }
    return data
}

// GetPayload retrieves the Payload from a given Note
func (note *Note) GetPayload() []byte {
    return note.Payload
}

// IsDeleted determines whether or not a given Note is deleted
func (note *Note) IsDeleted() bool {
    return note.Deleted
}

// EndpointID determines the endpoint that last modified the note
func (note *Note) EndpointID() string {
    if note.Histories == nil {
        return ""
    }
    histories := *note.Histories
    if len(histories) == 0 {
        return ""
    }
    return histories[0].EndpointID
}

// HasConflicts determines whether or not a given Note has conflicts
func (note *Note) HasConflicts() bool {
    if note.Conflicts == nil {
        return false
    }
    return len(*note.Conflicts) != 0
}

// GetConflicts fetches the conflicts, so that they may be displayed
func (note *Note) GetConflicts() []Note {
    if note.Conflicts == nil {
        return []Note{}
    }
    return *note.Conflicts
}

// GetModified retrieves information about the note's modification
func (note *Note) GetModified() (isAvailable bool, endpointID string, when string, where string, updates int32) {
    if note.Histories == nil || len(*note.Histories) == 0 {
        return;
    }
    histories := *note.Histories
    endpointID = histories[0].EndpointID
    when = time.Unix(0, histories[0].When * 1000000000).UTC().Format("2006-01-02T15:04:05Z")
    where = histories[0].Where
    updates = histories[0].Sequence
    isAvailable = true
    return
}

// Perform the bulk of Note Update, Delete, Merge operations
func (note *Note) UpdateNote(endpointID string, resolveConflicts bool, deleted bool) {
    var when int64
    var where string

    updates := note.Updates + 1
    sequence := updates

    // Purge the history of other updates performed by the endpoint doing this update
    newHistories := copyOrCreateBlankHistory(nil)

    if note.Histories != nil {
        histories := *note.Histories
        for _, History := range histories {

            // Bump sequence if the same endpoint is updating the note
            if History.EndpointID == endpointID && History.Sequence >= sequence {
                sequence = History.Sequence + 1
            }

            //  Sparse purge
            if History.EndpointID != endpointID {
                newHistories = append(newHistories, History)
            }
        }
    }

    note.Histories = &newHistories

    // Create a new history
    newHistory := NewHistory(endpointID, when, where, sequence)

    // Insert newHistory at offset 0, then append the old history
    histories := copyOrCreateBlankHistory(nil)
    histories = append(histories, newHistory)
    if note.Histories != nil && len(*note.Histories) > 0 {
        histories = append(histories, *note.Histories...)
    }
    note.Histories = &histories

    // Update the note with these key flags
    note.Updates = updates
    note.Deleted = deleted

    //  Unless performing implicit conflict resolution, this code resolves all conflicts
    //  but does not accomodate selective conflict resolution
    performConflictResolution := resolveConflicts
    performImplicitConflictResolution := !performConflictResolution

    if performConflictResolution || performImplicitConflictResolution {
        noteConflicts := copyOrCreateBlankConflict(note.Conflicts)
        noteHistories := copyOrCreateNonblankHistory(note.Histories)

        for conflictIndex := len(noteConflicts) - 1; conflictIndex >= 0; conflictIndex-- {
            conflict := noteConflicts[conflictIndex]
            conflictNote := conflict
            conflictNoteHistories := copyOrCreateNonblankHistory(conflictNote.Histories)

            if performImplicitConflictResolution && (conflictNoteHistories[0].EndpointID != endpointID) {
                continue
            }

            for conflictHistoryIndex := 0; conflictHistoryIndex < len(conflictNoteHistories); conflictHistoryIndex++ {
                conflictHistory := conflictNoteHistories[conflictHistoryIndex]

                isSubsumed := false

                for historyIndex := 0; historyIndex < len(noteHistories); historyIndex++ {
                    history := noteHistories[historyIndex]

                    if conflictHistory.isHistorySubsumedBy(&history) {
                        isSubsumed = true
                        break
                    }

                    if (history.EndpointID == conflictHistory.EndpointID) && (conflictHistory.Sequence >= sequence) {
                        sequence = conflictHistory.Sequence + 1
                        noteHistories[historyIndex].Sequence = sequence
                    }
                }

                if !isSubsumed {
                    //  Attempt to sparse purge again before incorporating conflict history
                    for historyIndex := len(noteHistories) - 1; historyIndex >= 0; historyIndex-- {
                        history := noteHistories[historyIndex]
                        if history.isHistorySubsumedBy(&conflictHistory) {
                            // Delete the note at historyIndex
                            if historyIndex < len(noteHistories) {
                                noteHistories = append(noteHistories[:historyIndex], noteHistories[historyIndex+1:]...)
                            }
                        }
                    }

                    // Insert conflictHistory at offset 1
                    noteHistories = append(noteHistories, conflictHistory)
                    if len(noteHistories) > 1 {
                        copy(noteHistories[2:], noteHistories[1:])
                        noteHistories[1] = conflictHistory
                    }
                }
            }

            // Delete the note at conflictIndex
            if conflictIndex < len(noteConflicts) {
                noteConflicts = append(noteConflicts[:conflictIndex], noteConflicts[conflictIndex+1:]...)
            }

        }

        if (performConflictResolution) {
            noteConflicts = copyOrCreateBlankConflict(nil)
        }

        note.Conflicts = &noteConflicts
        note.Histories = &noteHistories

    }

}

//  This function compares two Notes, and returns
//     1 if local note is newer than incoming note
//    -1 if incoming note is newer than local note
//     0 if notes are equal
//  conflictDataDiffers is returned as true if they
//     are equal but their conflict data is different
func (note *Note) CompareModified(incomingNote *Note) (conflictDataDiffers bool, result int) {

    if incomingNote.Updates > note.Updates {
        return false, -1
    }
    if incomingNote.Updates < note.Updates {
        return false, 1
    }

    noteHistories := copyOrCreateNonblankHistory(note.Histories)
    incomingNoteHistories := copyOrCreateNonblankHistory(incomingNote.Histories)
    noteConflicts := copyOrCreateBlankConflict(note.Conflicts)
    incomingNoteConflicts := copyOrCreateBlankConflict(incomingNote.Conflicts)

    localTopMostHistory := noteHistories[0]
    incomingTopMostHistory := incomingNoteHistories[0]

    if localTopMostHistory.When == 0 {
        if incomingTopMostHistory.When != 0 {
            return false, -1
        }
    }

    // Compare when modified
    if localTopMostHistory.When != 0 && incomingTopMostHistory.When != 0 {
        result := compareNoteTimestamps(localTopMostHistory.When, incomingTopMostHistory.When)
        if result != 0 {
            return false, result
        }
    }

    // UTF-8 string comparisons
    if localTopMostHistory.EndpointID < incomingTopMostHistory.EndpointID {
        return false, -1
    }
    if localTopMostHistory.EndpointID > incomingTopMostHistory.EndpointID {
        return false, 1
    }

    if len(noteConflicts) == len(incomingNoteConflicts) {

        if len(noteConflicts) > 0 {

            for index := 0; index < len(noteConflicts); index++ {
                localConflictNote := noteConflicts[index]
                matchingConflict := false

                for index2 := 0; index2 < len(incomingNoteConflicts); index2++ {
                    incomingConflictNote := incomingNoteConflicts[index2]
                    _, compareResult := incomingConflictNote.CompareModified(&localConflictNote)
                    if compareResult == 0 {
                        matchingConflict = true
                        break
                    }
                }

                if !matchingConflict {
                    return true, 0
                }
            }
        }

        return false, 0

    }

    return true, 0

}

// Determine whether or not this Note was subsumed by changes to another
func (note *Note) IsSubsumedBy(incomingNote *Note) bool {

    noteHistories := copyOrCreateNonblankHistory(note.Histories)
    incomingNoteHistories := copyOrCreateNonblankHistory(incomingNote.Histories)
    noteConflicts := copyOrCreateBlankConflict(note.Conflicts)
    incomingNoteConflicts := copyOrCreateBlankConflict(incomingNote.Conflicts)

    isSubsumed := false
    localTopMostHistory := noteHistories[0]

    for index := 0; index < len(incomingNoteHistories); index++ {
        incomingHistory := incomingNoteHistories[index]

        if localTopMostHistory.isHistorySubsumedBy(&incomingHistory) {
            isSubsumed = true
            break
        }

    }

    if !isSubsumed {
        for index := 0; index < len(incomingNoteConflicts); index++ {
            incomingConflict := incomingNoteConflicts[index]
            if note.IsSubsumedBy(&incomingConflict) {
                isSubsumed = true
                break
            }
        }
    }

    if !isSubsumed {
        return false
    }

    for index := 0; index < len(noteConflicts); index++ {
        isSubsumed = false

        localConflict := noteConflicts[index]
        if localConflict.IsSubsumedBy(incomingNote) {
            isSubsumed = true
            break
        }

        for index2 := 0; index2 < len(incomingNoteConflicts); index2++ {
            incomingConflict := incomingNoteConflicts[index2]
            if localConflict.IsSubsumedBy(&incomingConflict) {
                isSubsumed = true
                break
            }
        }

        if !isSubsumed {
            return false
        }
    }

    return true
}

// Copy a list of conflicts, else if blank create a new one that is truly blank, with 0 items
func copyOrCreateBlankConflict(conflictsToCopy *[]Note) []Note {
    if conflictsToCopy != nil {
        return *conflictsToCopy
    }
    return []Note{}
}

// Copy a history, else if blank create a new one that is truly blank, with 0 items
func copyOrCreateBlankHistory(historiesToCopy *[]NoteHistory) []NoteHistory {
    if historiesToCopy != nil {
        return *historiesToCopy
    }
    return []NoteHistory{}
}

// Copy a history, else if blank create a new entry with exactly 1 item
func copyOrCreateNonblankHistory(historiesToCopy *[]NoteHistory) []NoteHistory {
    if historiesToCopy != nil {
        return *historiesToCopy
    }
    histories := []NoteHistory{}
    histories = append(histories, NewHistory("", 0, "", 0))
    return histories
}

// NewHistory creates a history entry for a Note being modified
func NewHistory(endpointID string, when int64, where string, sequence int32) NoteHistory {

    newHistory := NoteHistory{}
    newHistory.EndpointID = endpointID

    if when == 0 {
        newHistory.When = createNoteTimestamp(endpointID)
    } else {
        newHistory.When = when
    }

    if where == "" {
	    // On the service we don't have a location
        newHistory.Where = ""
    } else {
        newHistory.Where = where
    }

    newHistory.Sequence = sequence

    return newHistory
}

// Determine whether or not a Note's history was subsumed by changes to another
func (thisHistory *NoteHistory) isHistorySubsumedBy(incomingHistory *NoteHistory) bool {

    Subsumed := (thisHistory.EndpointID == incomingHistory.EndpointID) && (incomingHistory.Sequence >= thisHistory.Sequence)

    return Subsumed

}

// Merge the contents of two Notes
func mergeTwoNotes(localNote *Note, incomingNote *Note) Note {

    //  Create flattened array for local note & conflicts
    clonedLocalNote := copyNote(localNote)
    localNotes := copyOrCreateBlankConflict(clonedLocalNote.Conflicts)
    clonedLocalNote.Conflicts = nil
    localNotes = append(localNotes, clonedLocalNote)

    //  Create flattened array for incoming note & conflicts
    clonedIncomingNote := copyNote(incomingNote)
    incomingNotes := copyOrCreateBlankConflict(clonedIncomingNote.Conflicts)
    clonedIncomingNote.Conflicts = nil
    incomingNotes = append(incomingNotes, clonedIncomingNote)

    //  Remove duplicates & subsumed notes - also get the winner
    mergeResultNotes := []Note{}
    winnerNote := mergeNotes(&localNotes, &incomingNotes, &mergeResultNotes, nil)
    winnerNote = mergeNotes(&incomingNotes, &localNotes, &mergeResultNotes, winnerNote)
    if len(mergeResultNotes) == 1 {
        return *winnerNote
    }

    //  Reconstruct conflicts for item
    winnerNoteConflicts := copyOrCreateBlankConflict(nil)
    for index := 0; index < len(mergeResultNotes); index++ {
        mergeResultNote := mergeResultNotes[index]
        _, compareResult := winnerNote.CompareModified(&mergeResultNote)
        if 0 == compareResult {
            continue
        }
        winnerNoteConflicts = append(winnerNoteConflicts, mergeResultNote)
    }
    winnerNote.Conflicts = &winnerNoteConflicts

    // Done
    return *winnerNote
}

// The bulk of processing for merging two sets of Notes into a single result set
func mergeNotes(outerNotes *[]Note, innerNotes *[]Note, mergeNotes *[]Note, winnerNote *Note) *Note {

    for outerNotesIndex := 0; outerNotesIndex < len(*outerNotes); outerNotesIndex++ {
        outerNote := (*outerNotes)[outerNotesIndex]
        outerNoteSubsumed := false

        for innerNotesIndex := 0; innerNotesIndex < len(*innerNotes); innerNotesIndex++ {
            innerNote := (*innerNotes)[innerNotesIndex]

            //  Note can be specially flagged due to subsumption below, so check for it
            if innerNote.Updates == -1 {
                continue
            }

            // See if the outer note is subsumed by any changes in the inner notes
            outerNoteHistories := copyOrCreateNonblankHistory(outerNote.Histories)
            innerNoteHistories := copyOrCreateNonblankHistory(innerNote.Histories)
            outerHistory := outerNoteHistories[0]
            for historyIndex := 0; historyIndex < len(innerNoteHistories); historyIndex++ {
                innerHistory := innerNoteHistories[historyIndex]
                if (outerHistory.EndpointID == innerHistory.EndpointID) {
                    outerNoteSubsumed = (innerHistory.Sequence >= outerHistory.Sequence)
                    break
                }
            }
            if outerNoteSubsumed {
                break
            }
        }

        if outerNoteSubsumed {
            //  Place a special flag on the note to indicate that it has been subsumed
            (*outerNotes)[outerNotesIndex].Updates = -1
            continue
        }

        // Free the conflicts on this outer note, and append it to the merge note list
        (*outerNotes)[outerNotesIndex].Conflicts = nil
        (*mergeNotes) = append((*mergeNotes), (*outerNotes)[outerNotesIndex])

        needToAssignWinner := (winnerNote == nil)
        if !needToAssignWinner {
            _, compareResult := winnerNote.CompareModified(&(*outerNotes)[outerNotesIndex])
            needToAssignWinner = (-1 == compareResult)
        }
        if needToAssignWinner {
            winnerNote = &outerNote
        }
    }

    return winnerNote
}

// copyNote duplicates a Note
func copyNote(note *Note) Note {

    newNote := Note{}
    newNote.Updates = note.Updates
    newNote.Deleted = note.Deleted
    newNote.Sent = note.Sent
    newNote.Body = note.Body
    newNote.Payload = note.Payload

    if note.Histories != nil {
        newHistories := copyOrCreateBlankHistory(note.Histories)
        newNote.Histories = &newHistories
    }

    if note.Conflicts != nil {
        newConflicts := copyOrCreateBlankConflict(note.Conflicts)
        newNote.Conflicts = &newConflicts
    }

    return newNote
}

// Return the current timestamp to be used for "when", in milliseconds
var lastTimestamp int64
func createNoteTimestamp(endpointID string) int64 {

    // Return the later of the current time (in seconds) and the last one that we handed out
    thisTimestamp := time.Now().UTC().UnixNano() / 1000000000
    if thisTimestamp <= lastTimestamp {
        lastTimestamp++
        thisTimestamp = lastTimestamp
    } else {
        lastTimestamp = thisTimestamp;
    }

    return thisTimestamp

}

// compareNoteTimestamps is a standard Compare function
func compareNoteTimestamps(left int64, right int64) int {
    if left < right {
        return -1
    } else if left > right {
        return 1
    }
    return 0
}
