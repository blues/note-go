// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/blues/note-go/notecard"
	"github.com/blues/note-go/noteutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Exit codes
const exitOk = 0
const exitFail = 1

// The open notecard
var card notecard.Context

// Main entry
func main() {

	// Substitute args if they're specified in an env var, which is handy both when using this on resin.io
	// and also when debugging with VS Code
	args := os.Getenv("ARGS")
	if args != "" {
		os.Args = strings.Split(os.Args[0]+" "+args, " ")
	}

	// Spawn our signal handler
	go signalHandler()

	// Process actions
	var actionRequest string
	flag.StringVar(&actionRequest, "req", "", "perform a request directly")
	var actionTrace bool
	flag.BoolVar(&actionTrace, "trace", false, "trace serial port output")

	// Parse these flags and also the note tool config flags
	err := noteutil.FlagParse()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(exitFail)
	}

	// If no action specified (i.e. just -port x), exit so that we don't touch the wrong port
	if len(os.Args) == 1 {
		fmt.Printf("Command arguments:\n")
		flag.PrintDefaults()
		noteutil.ConfigShow()
		fmt.Printf("\n")
		nInterface, nPort, _ := notecard.Defaults()
		if noteutil.Config.Interface != "" {
			nInterface = noteutil.Config.Interface
			nPort = noteutil.Config.Port
		}
		var ports []string
		if nInterface == notecard.NotecardInterfaceSerial {
			ports, _, _, _ = notecard.SerialPorts()
		}
		if nInterface == notecard.NotecardInterfaceI2C {
			ports, _, _, _ = notecard.I2CPorts()
		}
		if len(ports) != 0 {
			fmt.Printf("Ports on '%s':\n", nInterface)
			for _, port := range ports {
				if port == nPort {
					fmt.Printf("   %s ***\n", port)
				} else {
					fmt.Printf("   %s\n", port)
				}
			}
		}
		return
	}

	// Process the main part of the command line as a -req
	argsLeft := len(flag.Args())
	if argsLeft == 1 {
		actionRequest = flag.Args()[0]
	} else if argsLeft != 0 {
		remainingArgs := strings.Join(flag.Args()[1:], " ")
		fmt.Printf("Switches must be placed on the command line prior to the request: %s\n", remainingArgs)
		os.Exit(exitFail)
	}

	// Open the card, just to make sure errors are reported early
	card, err = notecard.Open(noteutil.Config.Interface, noteutil.Config.Port, noteutil.Config.PortConfig)
	if err != nil {
		fmt.Printf("Can't open card: %s\n", err)
		os.Exit(exitFail)

	}

	// Turn on Notecard library debug output
	card.DebugOutput(true, false)

	// Process non-config commands

	err = nil

	if actionRequest != "" {
		card.TransactionJSON([]byte(actionRequest))
	}

	if actionTrace {
		err = card.Trace()
	}

	// Process errors
	if err != nil {
		fmt.Printf("test error: %s\n", err)
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
