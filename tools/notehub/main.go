// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/blues/note-go/noteutil"
)

// Exit codes
const exitOk = 0
const exitFail = 1

// Main entry point
func main() {

	// Spawn our signal handler
	go signalHandler()

	// Process command line
	var flagReq string
	flag.StringVar(&flagReq, "req", "", "{json for device-like request}")
	var flagJq bool
	flag.BoolVar(&flagJq, "jq", false, "strip all non json lines from output so that jq can be used")
	var flagIn string
	flag.StringVar(&flagIn, "in", "", "input filename, enabling request to be contained in a file")
	var flagUpload string
	flag.StringVar(&flagUpload, "upload", "", "filename to upload")
	var flagType string
	flag.StringVar(&flagType, "type", "", "indicate file type of image such as 'firmware'")
	var flagTags string
	flag.StringVar(&flagTags, "tags", "", "indicate tags to attach to uploaded image")
	var flagNotes string
	flag.StringVar(&flagNotes, "notes", "", "indicate notes to attach to uploaded image")
	var flagOverwrite bool
	flag.BoolVar(&flagOverwrite, "overwrite", false, "use exact filename in upload and overwrite it on service")
	var flagOut string
	flag.StringVar(&flagOut, "out", "", "output filename")

	// Parse these flags and also the note tool config flags
	err := noteutil.FlagParse(false, true)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(exitFail)
	}

	// If no commands found, just show the config
	if len(os.Args) == 1 {
		fmt.Printf("\nCommand options:\n")
		flag.PrintDefaults()
		fmt.Printf("\nCurrent settings:\n")
		noteutil.ConfigShow()
		os.Exit(exitOk)
	}

	// Misc state flags
	outq := make(chan string)
	go func() {
		for {
			fmt.Printf("%s", <-outq)
		}
	}()

	// Process the main part of the command line as a -req
	argsLeft := len(flag.Args())
	if argsLeft == 1 {
		flagReq = flag.Args()[0]
	} else if argsLeft != 0 {
		remainingArgs := strings.Join(flag.Args()[1:], " ")
		fmt.Printf("Switches must be placed on the command line prior to the request: %s\n", remainingArgs)
		os.Exit(exitFail)
	}

	// Process input filename as a -req
	if flagIn != "" {
		if flagReq != "" {
			fmt.Printf("It's redundant to specify both -in as well as a request. Do one or the other.\n")
			os.Exit(exitFail)
		}
		contents, err := ioutil.ReadFile(flagIn)
		if err != nil {
			fmt.Printf("Can't read input file: %s\n", err)
			os.Exit(exitFail)
		}
		flagReq = string(contents)
	}

	// Process requests
	if flagReq != "" || flagUpload != "" {
		rsp, err := reqHubJSON(noteutil.Config.Hub, []byte(flagReq), flagUpload, flagType, flagTags, flagNotes, flagOverwrite, !noteutil.Config.Unsecure, flagJq, nil)
		if err != nil {
			fmt.Printf("Error processing request: %s\n", err)
			os.Exit(exitFail)
		}
		if flagOut == "" {
			fmt.Printf("%s", rsp)
		} else {
			outfile, err2 := os.Create(flagOut)
			if err2 != nil {
				fmt.Printf("Can't create output file: %s\n", err)
				os.Exit(exitFail)
			}
			outfile.Write(rsp)
			outfile.Close()
		}
	}

	// Success
	os.Exit(exitOk)

}

// Is this a reset command?
func stringOrReset(str string) string {

	switch str {
	case "none":
		fallthrough
	case "(none)":
		fallthrough
	case "reset":
		return ""
	}

	return str

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
