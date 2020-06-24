// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
	"github.com/blues/note-go/noteutil"
)

// Exit codes
const exitOk = 0
const exitFail = 1

// The open notecard
var card *notecard.Context

// Main entry
func main() {

	// Spawn our signal handler
	go signalHandler()

	// Process actions
	var actionRequest string
	flag.StringVar(&actionRequest, "req", "", "perform the specified request")
	var actionWhenConnected bool
	flag.BoolVar(&actionWhenConnected, "when-connected", false, "wait until connected")
	var actionWhenDisconnected bool
	flag.BoolVar(&actionWhenDisconnected, "when-disconnected", false, "wait until disconnected")
	var actionWhenDisarmed bool
	flag.BoolVar(&actionWhenDisarmed, "when-disarmed", false, "wait until ATTN is disarmed")
	var actionWhenSynced bool
	flag.BoolVar(&actionWhenSynced, "when-synced", false, "sync if needed and wait until sync completed")
	var actionLog string
	flag.StringVar(&actionLog, "log", "", "add a text string to the _log.qo notefile")
	var actionTrace bool
	flag.BoolVar(&actionTrace, "trace", false, "watch Notecard's trace output")
	var actionPlayground bool
	flag.BoolVar(&actionPlayground, "play", false, "enter JSON request/response playground")
	var actionPlaytime int
	flag.IntVar(&actionPlaytime, "playtime", 0, "enter number of minutes to play")
	var actionSync bool
	flag.BoolVar(&actionSync, "sync", false, "manually initiate a sync")
	var actionProduct string
	flag.StringVar(&actionProduct, "product", "", "set product UID")
	var actionSN string
	flag.StringVar(&actionSN, "sn", "", "set serial number")
	var actionHost string
	flag.StringVar(&actionHost, "host", "", "set notehub to be used")
	var actionInfo bool
	flag.BoolVar(&actionInfo, "info", false, "show information about the Notecard")
	var actionWatchLevel int
	flag.IntVar(&actionWatchLevel, "watch", -1, "watch ongoing sync status of a given level (0-5)")
	var actionCommtest bool
	flag.BoolVar(&actionCommtest, "commtest", false, "perform repetitive request/response test to validate comms with the Notecard")
	var actionSetup string
	flag.StringVar(&actionSetup, "setup", "", "issue requests sequentially as stored in the specified .json file")
	var actionScan string
	flag.StringVar(&actionScan, "scan", "", "scan a batch of notecards to collect info into a json file")

	// Parse these flags and also the note tool config flags
	err := noteutil.FlagParse(true, false)
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
	configVal := noteutil.Config.PortConfig
	if actionPlaytime != 0 {
		configVal = actionPlaytime
		actionPlayground = true
	}
	card, err = notecard.Open(noteutil.Config.Interface, noteutil.Config.Port, configVal)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(exitFail)

	}

	// Process non-config commands

	err = nil

	// Wait until disconnected
	if err == nil && actionWhenDisconnected {
		for {
			rsp, err := card.TransactionRequest(notecard.Request{Req: "service.status", NotefileID: notecard.SyncLogNotefile, Delete: true})
			if err != nil {
				fmt.Printf("%s\n", err)
				break
			}
			if strings.Contains(rsp.Status, note.ErrTransportDisconnected) {
				break
			}
			fmt.Printf("%s\n", rsp.Status)
			time.Sleep(3 * time.Second)
		}
	}

	// Wait until connected
	if err == nil && actionWhenConnected {
		for {
			delay := true
			rsp, err := card.TransactionRequest(notecard.Request{Req: "note.get", NotefileID: notecard.SyncLogNotefile, Delete: true})
			if err != nil && note.ErrorContains(err, note.ErrNoteNoExist) {
				delay = true
				err = nil
			}
			if err != nil {
				fmt.Printf("%s\n", err)
				break
			}
			if rsp.Connected {
				break
			} else if rsp.Body != nil {
				var body notecard.SyncLogBody
				note.BodyToObject(rsp.Body, &body)
				fmt.Printf("%s\n", body.Text)
			}
			if delay {
				time.Sleep(3 * time.Second)
			}
		}
	}

	// Wait until disarmed
	if err == nil && actionWhenDisarmed {
		for {
			rsp, err := card.TransactionRequest(notecard.Request{Req: "card.attn"})
			if err != nil {
				fmt.Printf("%s\n", err)
			} else if rsp.Set {
				break
			}
			time.Sleep(3 * time.Second)
		}
	}

	// Wait until synced
	if err == nil && actionWhenSynced {
		var rsp notecard.Request
		req := notecard.Request{Req: "service.sync.status"}
		req.Sync = true // Initiate sync if sync is needed
		rsp, err = card.TransactionRequest(req)
		for err == nil {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "service.sync.status"})
			if err != nil {
				fmt.Printf("%s\n", err)
				break
			}
			if rsp.Alert {
				fmt.Printf("sync error\n")
				break
			}
			if rsp.Completed > 0 {
				break
			}
			fmt.Printf("%s\n", rsp.Status)
			time.Sleep(3 * time.Second)
		}
	}

	// Turn on Notecard library debug output
	card.DebugOutput(true, false)

	if err == nil && actionInfo {

		card.DebugOutput(false, false)

		cardVersion := ""
		cardDeviceUID := ""
		cardName := ""
		rsp, err := card.TransactionRequest(notecard.Request{Req: "card.version"})
		if err == nil {
			cardName = rsp.Name
			cardDeviceUID = rsp.DeviceUID
			cardVersion = rsp.Version
		}

		cardICCID := ""
		cardIMSI := ""
		cardIMEI := ""
		cardICCIDX := ""
		cardIMSIX := ""
		cardModem := ""
		rsp, err = card.TransactionRequest(notecard.Request{Req: "card.wireless"})
		if err == nil {
			cardModem = rsp.Net.ModemFirmware
			cardIMEI = rsp.Net.Imei
			cardIMSI = rsp.Net.Imsi
			cardICCID = rsp.Net.Iccid
			cardIMSIX = rsp.Net.ImsiExternal
			cardICCIDX = rsp.Net.IccidExternal
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
			cardProductUID = "*** Product UID is not set. Please use notehub.io to create a project and a product UID ***"
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
					for notefileID, info := range *rsp.FileInfo {
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
		fmt.Printf("              ProductUID: %s\n", cardProductUID)
		fmt.Printf("           Serial Number: %s\n", cardSN)
		fmt.Printf("               DeviceUID: %s\n", cardDeviceUID)
		fmt.Printf("            Notehub Host: %s\n", cardHost)
		fmt.Printf("                 Version: %s\n", cardVersion)
		fmt.Printf("                   Modem: %s\n", cardModem)
		fmt.Printf("                   ICCID: %s\n", cardICCID)
		fmt.Printf("                    IMSI: %s\n", cardIMSI)
		fmt.Printf("                    IMEI: %s\n", cardIMEI)
		if cardICCIDX != "" {
			fmt.Printf("          External ICCID: %s\n", cardICCIDX)
			fmt.Printf("           External IMSI: %s\n", cardIMSIX)
		}
		fmt.Printf("             Provisioned: %s\n", cardProvisionedTime)
		fmt.Printf("       Used Over-the-Air: %d bytes\n", cardUsedBytes)
		fmt.Printf("               Sync Mode: %s\n", cardSyncMode)
		fmt.Printf("      Sync Upload Period: %d mins\n", cardUploadMins)
		fmt.Printf("         Download Period: %d hours\n", cardDownloadHrs)
		fmt.Printf("          Notehub Status: %s\n", cardServiceStatus)
		fmt.Printf("             Last Synced: %s\n", cardSyncedTime)
		fmt.Printf("                 Voltage: %0.02fV\n", cardVoltage)
		fmt.Printf("             Temperature: %0.02fC\n", cardTemp)
		fmt.Printf("                GPS Mode: %s\n", cardGPSMode)
		fmt.Printf("                Location: %s\n", cardLocation)
		fmt.Printf("               Currently: %s\n", cardTime)
		fmt.Printf("                  Booted: %s\n", cardBootedTime)
		fmt.Printf("               Notefiles: %s\n", cardNotefiles)
		fmt.Printf("   Notefile Storage Used: %d%%\n", cardStorageUsedPct)
		fmt.Printf("                     Env: %s\n", cardEnv)

	}

	if err == nil && actionProduct != "" {
		_, err = card.TransactionRequest(notecard.Request{Req: "service.set", ProductUID: actionProduct})
	}

	if err == nil && actionSN != "" {
		_, err = card.TransactionRequest(notecard.Request{Req: "service.set", SN: actionSN})
	}

	if err == nil && actionHost != "" {
		_, err = card.TransactionRequest(notecard.Request{Req: "service.set", Host: actionHost})
	}

	if err == nil && actionRequest != "" {
		_, err = card.TransactionJSON([]byte(actionRequest))
	}

	if err == nil && actionLog != "" {
		_, err = card.TransactionRequest(notecard.Request{Req: "service.log", Text: actionLog})
	}

	if err == nil && actionSync {
		_, err = card.TransactionRequest(notecard.Request{Req: "service.sync"})
	}

	if err == nil && actionSetup != "" && actionScan == "" {
		var requests []notecard.Request
		requests, err = loadRequests(actionSetup)
		if err == nil {
			repeat := false
			repeatForever := false
			countLeft := uint32(0)
			done := false
			for !done {
				for _, req := range requests {
					if req.Req == "delay" {
						time.Sleep(time.Duration(req.Seconds) * time.Second)
						continue
					}
					if req.Req == "repeat" {
						if !repeat {
							repeat = true
							countLeft = req.Count
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
					_, err = card.TransactionRequest(req)
					if err != nil {
						break
					}
				}
				if !repeat {
					break
				}
			}
		}
	}

	if err == nil && actionScan != "" {
		err = scan(actionSetup, actionScan)
	}

	if err == nil && actionCommtest {

		// Turn off debug output
		card.DebugOutput(false, false)

		// Turn off tracing because it can interfere with our rapid transaction I/O
		card.TransactionRequest(notecard.Request{Req: "card.io", Mode: "trace-off"})

		// Go into a high-frequency transaction loop
		transactions := 0
		began := time.Now()
		lastMessage := time.Now()
		for {
			_, err = card.TransactionRequest(notecard.Request{Req: "card.version"})
			if err != nil {
				break
			}
			transactions++
			if time.Now().Sub(lastMessage).Seconds() > 2 {
				lastMessage = time.Now()
				fmt.Printf("%d successful transactions (%0.2f/sec)\n", transactions, float64(transactions)/time.Now().Sub(began).Seconds())
			}
		}
	}

	if err == nil && actionTrace {
		err = card.Trace()
	}

	if err == nil && actionPlayground {
		fmt.Printf("You may now enter Notecard JSON requests interactively.\nType w to toggle Sync Watch, or q to quit.\n")
		for {
			card.DebugOutput(false, false)
			err = card.Interactive(false, actionWatchLevel, true, "w", "q")
			if !note.ErrorContains(err, note.ErrCardIo) || !notecard.IoErrorIsRecoverable {
				break
			}
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
