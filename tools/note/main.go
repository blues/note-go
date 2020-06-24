// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
	"github.com/blues/note-go/noteutil"
)

// Show Notecard trace output, used for debugging
var flagTrace = true

// Exit codes
const exitOk = 0
const exitFail = 1

// The open notecard
var card *notecard.Context

// Main entry
func main() {

	// Spawn our signal handler
	go signalHandler()

	// Flag for visualization of notecard I/O
	flag.BoolVar(&flagTrace, "trace", false, "trace notecard requests/responses")

	// Flags for your own contact
	var flagName string
	flag.StringVar(&flagName, "name", "", "set name in your contact")
	var flagEmail string
	flag.StringVar(&flagEmail, "email", "", "set email address in your contact")
	var flagContact bool
	flag.BoolVar(&flagContact, "contact", false, "show your own contact")
	var flagRemoveInactive int
	flag.IntVar(&flagRemoveInactive, "contact-clean-devices", 0, "remove devices that have been inactive for N days")
	var flagAddDevice string
	flag.StringVar(&flagAddDevice, "contact-add-device", "", "add a new device by device path")
	var flagRemoveDevice string
	flag.StringVar(&flagRemoveDevice, "contact-remove-device", "", "add a new device by device ID")

	// Flags for others' contacts
	var flagContacts bool
	flag.BoolVar(&flagContacts, "contacts", false, "list all contacts in contact list")
	var flagRemoveContact string
	flag.StringVar(&flagRemoveContact, "contact-remove", "", "remove a contact from contact list (email or DeviceUID)")

	// Flags to send a message, which can occur in any order
	var flagTo string
	flag.StringVar(&flagTo, "to", "", "recipient(s) for message (email or address list or reply message #)")
	var flagText string
	flag.StringVar(&flagText, "text", "", "ASCII text content of a note to be sent")
	var flagBody string
	flag.StringVar(&flagBody, "body", "", "JSON to be attached to a note to be sent")

	// Flags to sync messages
	var flagSync bool
	flag.BoolVar(&flagSync, "sync", false, "send and receive messages and watch progress")
	var flagWatch bool
	flag.BoolVar(&flagWatch, "watch", false, "watch sync progress")
	var flagOnline bool
	flag.BoolVar(&flagOnline, "online", false, "remain continuously connected to the notehub")
	var flagOffline bool
	flag.BoolVar(&flagOffline, "offline", false, "only periodically connect to the notehub")

	// Flags to manage the message store
	var flagShow bool
	flag.BoolVar(&flagShow, "show", false, "show all messages in message store")
	var flagTail int
	flag.IntVar(&flagTail, "tail", 0, "show the most recent N messages in message store")
	var flagDelete string
	flag.StringVar(&flagDelete, "delete", "", "delete specific # message(s)")
	var flagSave int
	flag.IntVar(&flagSave, "save", 0, "delete all but the most recent N messages in message store")

	// Parse these flags and also the note tool config flags
	err := noteutil.FlagParse(true, false)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(exitFail)
	}
	if len(os.Args) == 1 {
		flag.PrintDefaults()
		fmt.Printf("\nExamples:\n")
		fmt.Printf("    note -to \"larry@gmail.com,mark@fb.com\" -send \"howdy, partner\"\n")
		fmt.Printf("    note -send \"it's me\" -to \"api.a.notefile.net/messages!org.redcross.pager!imei:864475040013962\"\n")
		fmt.Printf("    note -sync\n")
		fmt.Printf("    note -count\n")
		fmt.Printf("    note -show\n")
		fmt.Printf("    note -tail 10\n")
		fmt.Printf("    note -save 5\n")
		fmt.Printf("    note -name \"Lisa Smith\" -email \"lsmith@msn.com\"\n")
		fmt.Printf("    note -contact\n")
		fmt.Printf("    note -contacts\n")
		fmt.Printf("    note -contact-remove spammer@aol.com\n")
		os.Exit(exitOk)
	}

	// Open the notecard API
	card, err = notecard.Open(noteutil.Config.Interface, noteutil.Config.Port, noteutil.Config.PortConfig)
	if err != nil {
		fmt.Printf("Can't open notecard: %s\n", err)
		os.Exit(exitFail)
	}

	// Turn on Notecard library debug output during debugging
	card.DebugOutput(flagTrace, false)

	// Process commands
	err = nil

	for {

		// Set your contact
		if flagName != "" || flagEmail != "" {
			_, err := contactSet(flagName, flagEmail, contactDefaultPurgeDays)
			if err != nil {
				break
			}
			flagContact = true
			err = messageValidateInbox()
			if err != nil {
				break
			}
		}

		// Add a device to your contact
		if flagAddDevice != "" {
			err = contactAddDevice(flagAddDevice)
			if err != nil {
				break
			}
			flagContact = true
		}

		// Remove a device from your contact
		if flagRemoveDevice != "" {
			err = contactRemoveDevice(flagRemoveDevice)
			if err != nil {
				break
			}
			flagContact = true
		}

		// Show your contact
		if flagContact {
			var contact note.MessageContact
			contact, err = contactGet()
			if err != nil {
				break
			}
			contactJSON, _ := note.JSONMarshalIndent(contact, "", "  ")
			fmt.Printf("%s\n", contactJSON)
		}

		// Remove inactive contacts
		if flagRemoveInactive != 0 {
			var contact note.MessageContact
			contact, err := contactSet("", "", flagRemoveInactive)
			if err != nil {
				break
			}
			contactJSON, _ := note.JSONMarshalIndent(contact, "", "  ")
			fmt.Printf("%s\n", contactJSON)
		}

		// Remove specified contact
		if flagRemoveContact != "" {
			err = contactRemove(flagRemoveContact)
			if err != nil {
				break
			}
		}

		// Show all contacts
		if flagContacts {
			err = contactShowOthers()
			if err != nil {
				break
			}
		}

		// Send a message
		if flagText != "" || flagBody != "" {
			if flagTo != "" {
				err = messageSend(flagTo, flagText, flagBody)
			} else {
				err = fmt.Errorf("no recipients specified with -to")
			}
			if err != nil {
				break
			}
		} else {
			if flagTo != "" {
				err = fmt.Errorf("no message specified to send to recipients")
				if err != nil {
					break
				}
			}
		}

		// Delete one or more messages
		if flagDelete != "" {
			err = messageDelete(flagDelete)
			if err != nil {
				break
			}
		}

		// Delete all but the N most recent messages
		if flagSave != 0 {
			err = messageDeleteTail(flagSave)
			if err != nil {
				break
			}
		}

		// Online/offline
		if flagOnline || flagOffline {
			req := notecard.Request{Req: "service.set"}
			if flagOnline {
				req.Mode = "continuous"
			} else {
				req.Mode = "periodic"
			}
			_, err = card.TransactionRequest(req)
			if err != nil {
				break
			}
		}

		// Sync and watch progress
		if flagSync {
			err = messageValidateInbox()
			if err != nil {
				break
			}
			_, err = card.TransactionRequest(notecard.Request{Req: "service.sync"})
			if err != nil {
				break
			}
			flagWatch = true
		}
		if flagWatch {
			watch(notecard.SyncLogLevelMinor, 10)
			messageMigrate()
		}

		// Show messages
		if flagTail != 0 || flagShow {
			messageMigrate()
			err = messageShow(flagTail)
			if err != nil {
				break
			}
		}

		// Done processing commands
		break

		// Sync
		if flagSync {
		}

	}

	// Process errors
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(exitFail)
	}

	// Success
	os.Exit(exitOk)

}

// Our app's signal handler
func signalHandler() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM)
	signal.Notify(ch, syscall.SIGINT)
	signal.Notify(ch, syscall.SIGSEGV)
	for {
		switch <-ch {
		case syscall.SIGINT:
			fmt.Printf(" (interrupted)\n")
			os.Exit(exitFail)
		case syscall.SIGTERM:
			break
		}
	}
}
