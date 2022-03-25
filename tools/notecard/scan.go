// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
)

// ICCID prefix-to-carrier mapping
var simPrefixToCarrier = []string{
	"898830", "twilio",
}

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

// ScannedSIM data structure
type ScannedSIM struct {
	Order string `json:"order,omitempty"`
	ICCID string `json:"iccid,omitempty"`
	Key   string `json:"key,omitempty"`
}

// Scan of a set of notecards, appending to JSON file.  Press ^C when done.
func scan(debugEnabled bool, init bool, fnSetup string, fnSetupSKU string, carrierProvision string, factoryReset bool, outfile string) (err error) {

	// Only allow one of the two
	if fnSetup != "" && fnSetupSKU != "" {
		err = fmt.Errorf("only one of setup or sku setup can be performed at a time")
		return
	}

	// Load the requests file
	var requests []map[string]interface{}
	if fnSetup != "" {
		requests, err = loadRequests(fnSetup)
		if err != nil {
			return
		}
	}

	// Load the requests string
	var requestsString string
	if fnSetupSKU != "" {
		requestsString, err = loadRequestsString(fnSetupSKU)
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

	// Generate a SIM file with a CSV extension
	simfile := strings.TrimSuffix(outfile, ".json") + ".csv"

	// Start the input handler
	go inputHandler()

	// Turn off debug output
	card.DebugOutput(debugEnabled, false)

	// Read the file into an array that we'll keep ordered
	var contents []byte
	var scannedDevices []ScannedDevice
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
			err = note.JSONUnmarshal(line, &v)
			if err != nil {
				fmt.Printf("*** error converting record into inventory JSON: %s\n%s\n", err, line)
			} else {
				scannedDevices = append(scannedDevices, v)
			}
		}
	}

	// Read the SIM file into an array that we'll keep ordered
	var scannedSIMs []ScannedSIM
	contents, err = ioutil.ReadFile(simfile)
	if err != nil {
		fmt.Printf("*** new file: %s\n", simfile)
	} else {
		jrecs := bytes.Split(contents, []byte("\n"))
		for i, line := range jrecs {
			if i == 0 { // header row
				continue
			}
			if len(line) == 0 {
				continue
			}
			var v ScannedSIM
			cols := strings.Split(string(line), ",")
			if len(cols) == 3 {
				v.Order = cols[0]
				v.ICCID = cols[1]
				v.Key = cols[2]
				scannedSIMs = append(scannedSIMs, v)
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
		card.DebugOutput(false, false)
		rsp, err = card.TransactionRequest(notecard.Request{Req: "hub.get"})
		card.DebugOutput(debugEnabled, false)
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

		// If requests string was specified, process it
		if requestsString != "" {
			req := notecard.Request{Req: "card.setup"}
			req.Text = requestsString
			rsp, err = card.TransactionRequest(req)
			if err != nil {
				break
			}
			if !factoryReset {
				card.TransactionRequest(notecard.Request{Req: "card.restart"})
				for i := 0; i < 5; i++ {
					_, err = card.TransactionRequest(notecard.Request{Req: "hub.get"})
					if err == nil {
						break
					}
				}
				if err != nil {
					break
				}
			}
		}

		// If they desired a factory reset, do so.  Note that this must be after the
		// card.setup so that we do the reset in the context of post-script execution
		if factoryReset {
			req := notecard.Request{Req: "card.restore"}
			req.Delete = true
			card.TransactionRequest(req)
			for i := 0; i < 5; i++ {
				_, err = card.TransactionRequest(notecard.Request{Req: "hub.get"})
				if err == nil {
					break
				}
			}
			if err != nil {
				break
			}
		}

		// If requests were specified, process them
		if len(requests) > 0 {
			// Process the requests
			err = processRequests(init, requests)
			if err != nil {
				break
			}
			// Re-do the hub.get because the setup script may have changed things
			card.DebugOutput(false, false)
			rsp, _ = card.TransactionRequest(notecard.Request{Req: "hub.get"})
			card.DebugOutput(debugEnabled, false)
		}

		// Create a new inventory record
		sir := ScannedSIM{}
		ir := ScannedDevice{}
		ir.DeviceUID = rsp.DeviceUID
		ir.Hub = rsp.Host
		ir.SN = rsp.SN
		ir.ProductUID = rsp.ProductUID

		// Take inventory

		card.DebugOutput(false, false)

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

		sir.ICCID = ir.Factory.ICCID
		sir.Key = ir.Factory.SIMActivationKey

		card.DebugOutput(debugEnabled, false)

		// Provision the device if requested
		if carrierProvision != "" {
			carrier := ""
			for i := 0; i < len(simPrefixToCarrier)/2; i++ {
				if strings.HasPrefix(sir.ICCID, simPrefixToCarrier[i*2]) {
					carrier = simPrefixToCarrier[i*2+1]
					break
				}
			}

			// Perform per-carrier provisioning
			switch carrier {

			case "twilio":
				err = twilioProvision(carrierProvision, sir.ICCID, sir.Key)
				if err != nil {
					return
				}

			default:
				fmt.Printf("\nPROVISIONING CARRIER NOT FOUND for SIM %s\n", sir.ICCID)
				return

			}

		}

		// Delete this card from the array if it's there, and append it
		found := false
		for _, v := range scannedDevices {
			if v.DeviceUID == ir.DeviceUID {
				found = true
				break
			}
		}
		if found {
			rnew := []ScannedDevice{}
			for _, v := range scannedDevices {
				if v.DeviceUID != ir.DeviceUID {
					rnew = append(rnew, v)
				}
			}
			scannedDevices = rnew
		}
		scannedDevices = append(scannedDevices, ir)

		// Delete this card from the SIM array if it's there, and append it
		for i, v := range scannedSIMs {
			if v.ICCID == sir.ICCID {
				scannedSIMs = append(scannedSIMs[0:i], scannedSIMs[i+1:]...)
				break
			}
		}
		scannedSIMs = append(scannedSIMs, sir)

		// Write the file
		f, err2 := os.Create(outfile)
		if err2 != nil {
			err = err2
			return
		}
		w := bufio.NewWriter(f)
		for _, v := range scannedDevices {
			vj, err := note.JSONMarshal(v)
			if err != nil {
				continue
			}
			f.Write(vj)
			f.Write([]byte("\r\n"))
		}
		w.Flush()
		f.Close()

		// Write the SIM file
		f, err2 = os.Create(simfile)
		if err2 != nil {
			err = err2
			return
		}
		w = bufio.NewWriter(f)
		f.WriteString("order_sid,iccid,registration_code\r\n")
		for _, v := range scannedSIMs {
			f.WriteString(fmt.Sprintf("%s,%s,%s\r\n", v.Order, v.ICCID, v.Key))
		}
		w.Flush()
		f.Close()

		// Done
		fmt.Printf("\n*** please remove the notecard\n")

	}

	// Done
	return
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

		case "q":
			os.Exit(0)

		default:
			fmt.Printf("Unrecognized: '%s'\n", text)

		}

	}

}

// Load requests from a JSON request file
func loadRequests(filename string) (requests []map[string]interface{}, err error) {

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
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		// Allow comments in either C or Python style
		s := string(line)
		if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "#") {
			continue
		}
		var req map[string]interface{}
		err = note.JSONUnmarshal(line, &req)
		if err != nil {
			err = fmt.Errorf("error: invalid request JSON: %s", line)
			return
		}
		requests = append(requests, req)
	}

	// Done
	return
}

// Load requests from a JSON request file, validating them and newline-separating into a string
func loadRequestsString(filename string) (requests string, err error) {

	// If the caller is resetting the requests, do it
	if filename == "-" {
		requests = filename
		return
	}

	// Iterate over the requests, converting them into a newline-delimited string
	var reqv []map[string]interface{}
	reqv, err = loadRequests(filename)
	for _, req := range reqv {
		var jsondata []byte
		jsondata, err = note.JSONMarshal(req)
		if err != nil {
			return
		}
		if requests != "" {
			requests += "\n"
		}
		requests += string(jsondata)
	}

	// Done
	return
}

// Process a set of requests
func processRequests(init bool, requests []map[string]interface{}) (err error) {
	repeat := false
	repeatForever := false
	countLeft := uint32(0)
	done := false
	for !done {
		if init {
			req := notecard.Request{Req: "card.restore"}
			req.Delete = true
			_, err = card.TransactionRequest(req)
			if err != nil {
				break
			}
		}
		for _, req := range requests {
			if req["req"] == "delay" {
				time.Sleep(time.Duration(req["seconds"].(int)) * time.Second)
				continue
			}
			if req["req"] == "repeat" {
				if !repeat {
					repeat = true
					countLeft = req["count"].(uint32)
					if countLeft == 0 {
						repeatForever = true
					}
				} else {
					if countLeft > 0 {
						countLeft--
					}
					if countLeft == 0 && !repeatForever {
						done = true
					}
				}
				continue
			}
			var reqJSON []byte
			reqJSON, err = note.JSONMarshal(req)
			if err != nil {
				break
			}
			_, err = card.TransactionJSON(reqJSON)
			if err != nil {
				break
			}
		}
		_, err = card.TransactionRequest(notecard.Request{Req: "card.checkpoint"})
		if err != nil {
			break
		}
		if !repeat {
			break
		}
	}
	return
}
