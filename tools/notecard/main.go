// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"time"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"encoding/json"
	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
	"github.com/blues/note-go/noteutil"
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
	flag.StringVar(&actionRequest, "req", "", "perform the specified request")
	var actionTrace bool
	flag.BoolVar(&actionTrace, "trace", false, "watch Notecard's trace output")
	var actionPlayground bool
	flag.BoolVar(&actionPlayground, "play", false, "enter JSON request/response playground")
	var actionSync bool
	flag.BoolVar(&actionSync, "sync", false, "manually initiate a sync")
	var actionInfo bool
	flag.BoolVar(&actionInfo, "info", false, "show information about the Notecard")
	var actionWatch bool
	flag.BoolVar(&actionWatch, "watch", false, "watch ongoing sync status")
	var actionWatchAll bool
	flag.BoolVar(&actionWatchAll, "watchall", false, "watch ongoing sync status with full details")
	var actionWatchLevel int
	flag.IntVar(&actionWatchLevel, "watchlevel", -1, "watch ongoing sync status of a given level (0-2)")

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

	if err == nil && actionInfo {

		cardVersion := ""
		cardDeviceUID := ""
		cardName := ""
		cardICCID := ""
		cardIMSI := ""
		cardIMEI := ""
		cardModem := ""
		rsp, err := card.TransactionRequest(notecard.Request{Req: "card.version"})
		if err == nil {
			cardName = rsp.Name
			cardDeviceUID = rsp.DeviceUID
			cardVersion = rsp.Version
			cardWireless := strings.Split(rsp.Wireless, ",")
			if len(cardWireless) >= 4 {
				cardModem = cardWireless[3]
			}
			if len(cardWireless) >= 3 {
				cardIMEI = cardWireless[2]
			}
			if len(cardWireless) >= 2 {
				cardIMSI = cardWireless[1]
			}
			if len(cardWireless) >= 1 {
				cardICCID = cardWireless[0]
			}
		}

		cardSN := ""
		cardHost := ""
		cardProductUID := ""
		cardSyncMode := "periodic"
		cardUploadMins := 60
		cardDownloadHrs := 0
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "service.get"})
			if err == nil {
				cardSN = rsp.SN
				cardHost = rsp.Host
				cardProductUID = rsp.ProductUID
				cardSyncMode = rsp.Mode
				cardUploadMins = int(rsp.Minutes)
				cardDownloadHrs = int(rsp.Hours)
			}
		}

		if cardProductUID == "" {
			cardProductUID = "*** PRODUCT UID NOT YET SET. PLEASE USE NOTEHUB.IO TO CREATE A PROJECT AND A PRODUCT UID ***"
		}

		cardVoltage := 0.0
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "card.voltage"})
			if err == nil {
				cardVoltage = rsp.Value
			}
		}

		cardTemp := 0.0
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "card.temp"})
			if err == nil {
				cardTemp = rsp.Value
			}
		}

		cardGPSMode := ""
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "card.location.mode"})
			if err == nil {
				if rsp.Status == "" {
					cardGPSMode = rsp.Mode
				} else {
					cardGPSMode = rsp.Mode + " (" + rsp.Status + ")"
				}
			}
		}

		cardTime := ""
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "card.time"})
			if err == nil {
				cardTime = time.Unix(int64(rsp.Time), 0).Format("2006-01-02T15:04:05Z") + " (" +
					time.Unix(int64(rsp.Time), 0).Local().Format("2006-01-02 3:04:05 PM MST") + ")"
			}
		}

		cardLocation := ""
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "card.location"})
			if err == nil {
				if rsp.Latitude != 0 || rsp.Longitude != 0 {
					cardLocation = fmt.Sprintf("%f,%f (%s)", rsp.Latitude, rsp.Longitude, rsp.LocationOLC)
				}
			}
		}

		cardBootedTime := ""
		cardStorageUsedPct := 0
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "card.status"})
			if err == nil {
				cardBootedTime = time.Unix(int64(rsp.Time), 0).Format("2006-01-02T15:04:05Z") + " (" +
					time.Unix(int64(rsp.Time), 0).Local().Format("2006-01-02 3:04:05 PM MST") + ")"
				cardStorageUsedPct = int(rsp.Storage)
			}
		}

		cardSyncedTime := ""
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "service.sync.status"})
			if err == nil {
				cardSyncedTime = time.Unix(int64(rsp.Time), 0).Format("2006-01-02T15:04:05Z") + " (" +
					time.Unix(int64(rsp.Time), 0).Local().Format("2006-01-02 3:04:05 PM MST") + ")"
			}
		}

		cardServiceStatus := ""
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "service.status"})
			if err == nil {
				cardServiceStatus = rsp.Status
				if rsp.Connected {
					cardServiceStatus += " (connected)"
				}
			}
		}

		cardProvisionedTime := ""
		cardUsedBytes := 0
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "card.usage.get"})
			if err == nil {
				cardProvisionedTime = time.Unix(int64(rsp.Time), 0).Format("2006-01-02T15:04:05Z") + " (" +
					time.Unix(int64(rsp.Time), 0).Local().Format("2006-01-02 3:04:05 PM MST") + ")"
				cardUsedBytes = int(rsp.BytesSent + rsp.BytesReceived)
			}
		}

		cardEnv := ""
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "env.get"})
			if err == nil {
				cardEnv = rsp.Text
				cardEnv = strings.TrimSuffix(cardEnv, "\n")
				cardEnv = strings.Replace(cardEnv, "\n", ", ", -1)
			}
		}

		cardNotefiles := ""
		if err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "file.changes"})
			if err == nil {
				if rsp.FileInfo != nil {
					for notefileID, info := range(*rsp.FileInfo) {
						if cardNotefiles != "" {
							cardNotefiles += ", "
						}
						if info.Changes > 0 {
							cardNotefiles += fmt.Sprintf("%s (%d)", notefileID, info.Changes)
						} else {
							cardNotefiles += notefileID
						}
					}
				}
			}
		}

		fmt.Printf("\n%s\n", cardName)
		fmt.Printf("			  ProductUID: %s\n", cardProductUID)
		fmt.Printf("		   Serial Number: %s\n", cardSN)
		fmt.Printf("			   DeviceUID: %s\n", cardDeviceUID)
		fmt.Printf("			Notehub Host: %s\n", cardHost)
		fmt.Printf("				 Version: %s\n", cardVersion)
		fmt.Printf("				   Modem: %s\n", cardModem)
		fmt.Printf("				   ICCID: %s\n", cardICCID)
		fmt.Printf("					IMSI: %s\n", cardIMSI)
		fmt.Printf("					IMEI: %s\n", cardIMEI)
		fmt.Printf("			 Provisioned: %s\n", cardProvisionedTime)
		fmt.Printf("	   Used Over-the-Air: %d bytes\n", cardUsedBytes)
		fmt.Printf("			   Sync Mode: %s\n", cardSyncMode)
		fmt.Printf("	  Sync Upload Period: %d mins\n", cardUploadMins)
		fmt.Printf("		 Download Period: %d hours\n", cardDownloadHrs)
		fmt.Printf("		  Notehub Status: %s\n", cardServiceStatus)
		fmt.Printf("			 Last Synced: %s\n", cardSyncedTime)
		fmt.Printf("				 Voltage: %0.02fV\n", cardVoltage)
		fmt.Printf("			 Temperature: %0.02fC\n", cardTemp)
		fmt.Printf("				GPS Mode: %s\n", cardGPSMode)
		fmt.Printf("				Location: %s\n", cardLocation)
		fmt.Printf("			   Currently: %s\n", cardTime)
		fmt.Printf("				  Booted: %s\n", cardBootedTime)
		fmt.Printf("			   Notefiles: %s\n", cardNotefiles)
		fmt.Printf("   Notefile Storage Used: %d%%\n", cardStorageUsedPct)
		fmt.Printf("					 Env: %s\n", cardEnv)

	}

	if err == nil && actionPlayground {
		fmt.Printf("You may now enter Notecard JSON requests interactively:\n");
		err = card.Interactive()
	}

	if err == nil && actionRequest != "" {
		card.TransactionJSON([]byte(actionRequest))
	}

	if err == nil && actionSync {
		_, err = card.TransactionRequest(notecard.Request{Req: "service.sync"})
	}

	if err == nil && actionTrace {
		err = card.Trace()
	}

	if err == nil && actionWatch {
		actionWatchLevel = notecard.SyncLogLevelMajor
	}
	if err == nil && actionWatchAll {
		actionWatchLevel = notecard.SyncLogLevelDetail
	}
	if err == nil && actionWatchLevel != -1 {
		var rsp notecard.Request
		var colWidth int
		var cols int
		var subsystem []string
		var subsystemDisplayName []string

		// Turn off Notecard library debug output
		card.DebugOutput(false, false)

		// Get the template for the trace log results
		rsp, err = card.TransactionRequest(notecard.Request{Req: "note.get", NotefileID: "_synclog.qi", Start: true})
		if err == nil {
			for _, entry := range(strings.Split(rsp.Status, ",")) {
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

		// Print an opening banner if necessary
		linesDisplayed := 0
		rsp, err = card.TransactionRequest(notecard.Request{Req: "note.get", NotefileID: "_synclog.qi"})
		if err == nil && rsp.Body == nil {
			fmt.Printf("%s waiting for sync activity\n", now)
		}

		// Loop, printing data
		prevTimeSecs := int64(0)
		for err == nil {

			// Get the next entry
			rsp, err = card.TransactionRequest(notecard.Request{Req: "note.get", NotefileID: "_synclog.qi", Delete: true})
			if err != nil {
				if note.ErrorContains(err, "invalid character") {
					err = fmt.Errorf("can't enter commands in a different trace window while 'watching'")
					break
				}
				continue
			}
			if rsp.Body == nil {
				time.Sleep(500 * time.Millisecond)
				continue;
			}
			var bodyJSON []byte
			bodyJSON, err = note.ObjectToJSON(rsp.Body)
			if err != nil {
				break
			}
			var body notecard.SyncLogBody
			err = json.Unmarshal(bodyJSON, &body)
			if err != nil {
				break
			}
			if body.DetailLevel > uint32(actionWatchLevel) {
				continue
			}

			// Output a header if it will help readability
			if linesDisplayed % 250 == 0 {
				fmt.Printf("\n%s ", strings.Repeat(" ", len(now)))
				for i:=0; i<cols; i++ {
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
			for _, ss := range(subsystem) {
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

		}

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
