// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
)

// ScannedDevice data structure
type ScannedDevice struct {
	DeviceUID   string            `json:"device,omitempty"`
	Factory     notecard.CardTest `json:"factory,omitempty"`
	Hub         string            `json:"hub,omitempty"`
	SN          string            `json:"sn,omitempty"`
	ProductUID  string            `json:"product,omitempty"`
	Firmware    string            `json:"firmware,omitempty"`
	Provisioned int64             `json:"activated,omitempty"`
	BytesUsed   uint32            `json:"bytes_used,omitempty"`
}

// Scan of a set of notecards, appending to JSON file.  Press ^C when done.
func scan(reqfile string, outfile string) (err error) {

	// Load the request file
	var requests []notecard.Request
	if reqfile != "" {
		requests, err = loadRequests(reqfile)
		if err != nil {
			return
		}
	}

	// Require a json file
	if !strings.HasSuffix(outfile, ".json") {
		if strings.Contains(outfile, ".") {
			return fmt.Errorf("only the .json file type is supported")
		}
		outfile += ".json"
	}

	// Start the input handler
	go inputHandler()

	// Turn off debug output
	card.DebugOutput(false, false)

	// Read the file into an array that we'll keep ordered
	var contents []byte
	var records []ScannedDevice
	contents, err = ioutil.ReadFile(outfile)
	if err != nil {
		fmt.Printf("*** new file: %s\n", outfile)
	} else {
		jrecs := bytes.Split(contents, []byte("\n"))
		for _, line := range jrecs {
			if len(line) == 0 {
				continue
			}
			var v ScannedDevice
			err = json.Unmarshal(line, &v)
			if err != nil {
				fmt.Printf("*** error converting record into inventory JSON: %s\n%s\n", err, line)
			} else {
				records = append(records, v)
			}
		}
	}

	// Loop, connecting with the card
	first := true
	sawDisconnected := true
	for {

		// Delay so as not to overwhelm the card
		time.Sleep(1 * time.Second)

		// See if it's available
		var rsp notecard.Request
		rsp, err = card.TransactionRequest(notecard.Request{Req: "service.get"})
		if note.ErrorContains(err, note.ErrCardIo) {
			if !sawDisconnected || first {
				first = false
				sawDisconnected = true
				fmt.Printf("\n*** please insert the next notecard, or enter q to quit\n")
			}
			continue
		}
		if err != nil {
			fmt.Printf("%s\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if !sawDisconnected {
			continue
		}
		first = false
		sawDisconnected = false
		fmt.Printf("\n%s\n", rsp.DeviceUID)

		// If requests were specified, process them
		if len(requests) > 0 {
			for _, req := range requests {
				_, err = card.TransactionRequest(req)
				if err != nil {
					break
				}
			}
			// Re-do the service.get because the setup script may have changed things
			rsp, _ = card.TransactionRequest(notecard.Request{Req: "service.get"})
		}

		// Create a new inventory record
		ir := ScannedDevice{}
		ir.DeviceUID = rsp.DeviceUID
		ir.Hub = rsp.Host
		ir.SN = rsp.SN
		ir.ProductUID = rsp.ProductUID

		rsp, err = card.TransactionRequest(notecard.Request{Req: "card.version"})
		if err == nil {
			ir.Firmware = rsp.Version
		}

		rsp, err = card.TransactionRequest(notecard.Request{Req: "card.usage.get"})
		if err == nil {
			ir.Provisioned = int64(rsp.Time)
			ir.BytesUsed = rsp.BytesSent + rsp.BytesReceived
		}

		rsp, err = card.TransactionRequest(notecard.Request{Req: "card.test"})
		if err == nil {
			note.BodyToObject(rsp.Body, &ir.Factory)
		}

		// Delete this card from the array if it's there
		found := false
		for _, v := range records {
			if v.DeviceUID == ir.DeviceUID {
				found = true
				break
			}
		}
		if found {
			rnew := []ScannedDevice{}
			for _, v := range records {
				if v.DeviceUID != ir.DeviceUID {
					rnew = append(rnew, v)
				}
			}
			records = rnew
		}

		// Append this record
		records = append(records, ir)

		// Write the file
		f, err2 := os.Create(outfile)
		if err2 != nil {
			err = err2
			return
		}
		w := bufio.NewWriter(f)
		for _, v := range records {
			vj, err := json.Marshal(v)
			if err != nil {
				continue
			}
			f.Write(vj)
			f.Write([]byte("\n"))
		}
		w.Flush()
		f.Close()

		// Done
		fmt.Printf("\n*** please remove the notecard\n")

	}
}

// Background input handler
func inputHandler() {

	// Create a scanner to watch stdin
	scanner := bufio.NewScanner(os.Stdin)
	var text string

	for {

		scanner.Scan()
		text = scanner.Text()

		switch strings.ToLower(text) {

		case "":

		default:
			fmt.Printf("Unrecognized: '%s'\n", text)

		case "q":
			os.Exit(0)

		}

	}

}

// Load requests from a JSON request file
func loadRequests(filename string) (requests []notecard.Request, err error) {

	// Require a json file
	if !strings.HasSuffix(filename, ".json") {
		if strings.Contains(filename, ".") {
			err = fmt.Errorf("requests must be in a .json file")
			return
		}
		filename += ".json"
	}

	// Read the file into an array of requests
	var contents []byte
	contents, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	jrecs := bytes.Split(contents, []byte("\n"))
	for _, line := range jrecs {
		if len(line) == 0 {
			continue
		}
		var req notecard.Request
		err = json.Unmarshal(line, &req)
		if err != nil {
			err = fmt.Errorf("error: invalid request JSON: %s", line)
			return
		}
		requests = append(requests, req)
	}

	// Done
	return
}
