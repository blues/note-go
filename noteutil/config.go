// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package noteutil

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notehub"
)

// ConfigCreds are the credentials for a given notehub
type ConfigCreds struct {
	User  string `json:"user,omitempty"`
	Token string `json:"token,omitempty"`
}

// ConfigSettings defines the config file that maintains the command processor's state
type ConfigSettings struct {
	When       string                 `json:"when,omitempty"`
	Hub        string                 `json:"hub,omitempty"`
	HubCreds   map[string]ConfigCreds `json:"creds,omitempty"`
	Interface  string                 `json:"interface,omitempty"`
	Port       string                 `json:"port,omitempty"`
	PortConfig int                    `json:"port_config,omitempty"`
}

// Config are the master config settings
var Config ConfigSettings
var configFlags ConfigSettings

// ConfigRead reads the current info from config file
func ConfigRead() error {

	// As a convenience to all tools, generate a new random seed for each iteration
	rand.Seed(time.Now().UnixNano())
	rand.Seed(rand.Int63() ^ time.Now().UnixNano())

	// Read the config file
	contents, err := ioutil.ReadFile(configSettingsPath())
	if os.IsNotExist(err) {
		ConfigReset()
		err = nil
	} else if err == nil {
		err = note.JSONUnmarshal(contents, &Config)
		if err != nil || Config.When == "" {
			ConfigReset()
			if err != nil {
				err = fmt.Errorf("can't read configuration: %s", err)
			}
		}
	}

	return err

}

// ConfigWrite updates the file with the current config info
func ConfigWrite() error {

	// Marshal it
	configJSON, _ := note.JSONMarshalIndent(Config, "", "    ")

	// Write the file
	fd, err := os.OpenFile(configSettingsPath(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	fd.Write(configJSON)
	fd.Close()

	// Done
	return err

}

// Reset the comms to default
func configResetInterface() {
	Config = ConfigSettings{}
}

// ConfigReset updates the file with the default info
func ConfigReset() {
	configResetInterface()
	ConfigSetHub("-")
	Config.When = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	return
}

// ConfigShow displays all current config parameters
func ConfigShow() error {

	fmt.Printf("\nCurrently saved values:\n")

	if Config.Hub != "" {
		fmt.Printf("       hub: %s\n", Config.Hub)
	}
	if Config.HubCreds == nil {
		Config.HubCreds = map[string]ConfigCreds{}
	}
	if len(Config.HubCreds) != 0 {
		fmt.Printf("     creds:\n")
		for hub, cred := range Config.HubCreds {
			if hub == "" {
				hub = "api.notefile.net"
			}
			fmt.Printf("            %s: %s\n", hub, cred.User)
		}
	}
	if Config.Interface != "" {
		fmt.Printf("   -interface %s\n", Config.Interface)
		if Config.Port == "" {
			fmt.Printf("   -port -\n")
			fmt.Printf("   -portconfig -\n")
		} else {
			fmt.Printf("   -port %s\n", Config.Port)
			fmt.Printf("   -portconfig %d\n", Config.PortConfig)
		}
	}

	return nil

}

// ConfigFlagsProcess processes the registered config flags
func ConfigFlagsProcess() (err error) {

	// Read if not yet read
	if Config.When == "" {
		err = ConfigRead()
		if err != nil {
			return
		}
	}

	// Set or reset the flags as desired
	if configFlags.Hub != "" {
		ConfigSetHub(configFlags.Hub)
	}
	if configFlags.Interface == "-" {
		configResetInterface()
	} else if configFlags.Interface != "" {
		Config.Interface = configFlags.Interface
	}
	if configFlags.Port == "-" {
		Config.Port = ""
	} else if configFlags.Port != "" {
		Config.Port = configFlags.Port
	}
	if configFlags.PortConfig < 0 {
		Config.PortConfig = 0
	} else if configFlags.PortConfig != 0 {
		Config.PortConfig = configFlags.PortConfig
	}
	if Config.Interface == "" {
		configFlags.Port = ""
		configFlags.PortConfig = 0
	}

	// Done
	return nil

}

// ConfigFlagsRegister registers the config-related flags
func ConfigFlagsRegister(notecardFlags bool, notehubFlags bool) {

	// Process the commands
	if notecardFlags {
		flag.StringVar(&configFlags.Interface, "interface", "", "select 'serial' or 'i2c' interface for notecard")
		flag.StringVar(&configFlags.Port, "port", "", "select serial or i2c port for notecard")
		flag.IntVar(&configFlags.PortConfig, "portconfig", 0, "set serial device speed or i2c address for notecard")
	}
	if notehubFlags {
		flag.StringVar(&configFlags.Hub, "hub", "", "set notehub domain")
	}

}

// FlagParse is a wrapper around flag.Parse that handles our config flags
func FlagParse(notecardFlags bool, notehubFlags bool) (err error) {

	// Register our flags
	ConfigFlagsRegister(notecardFlags, notehubFlags)

	// Parse them
	flag.Parse()

	// Process our flags
	err = ConfigFlagsProcess()
	if err != nil {
		return
	}

	// If our flags were the only ones present, save them
	configOnly := true
	if len(os.Args) == 1 {
		configOnly = false
	} else {
		for i, arg := range os.Args {
			// Even arguments are parameters, odd args are flags
			if (i & 1) != 0 {
				switch arg {
				case "-interface":
				case "-port":
				case "-portconfig":
				case "-hub":
				// any odd argument that isn't one of our switches
				default:
					configOnly = false
					break
				}
			}
		}
	}
	if configOnly {
		fmt.Printf("*** saving configuration ***")
		ConfigWrite()
		ConfigShow()
	}

	// Override, just for this session, with env vars
	str := os.Getenv("NOTE_INTERFACE")
	if str != "" {
		Config.Interface = str
	}

	// Override via env vars if specified
	str = os.Getenv("NOTE_PORT")
	if str != "" {
		Config.Port = str
		str := os.Getenv("NOTE_PORT_CONFIG")
		strint, err2 := strconv.Atoi(str)
		if err2 != nil {
			strint = Config.PortConfig
		}
		Config.PortConfig = strint
	}

	// Done
	return

}

// ConfigSignedIn returns info about whether or not we're signed in
func ConfigSignedIn() (username string, token string, authenticated bool) {
	if Config.HubCreds == nil {
		Config.HubCreds = map[string]ConfigCreds{}
	}
	creds, present := Config.HubCreds[Config.Hub]
	if present {
		if creds.Token != "" && creds.User != "" {
			authenticated = true
			username = creds.User
			token = creds.Token
		}
	}

	return

}

// ConfigAuthenticationHeader sets the authorization field in the header as appropriate
func ConfigAuthenticationHeader(httpReq *http.Request) (err error) {

	// Exit if not signed in
	_, token, authenticated := ConfigSignedIn()
	if !authenticated {
		err = fmt.Errorf("not authenticated to %s: please use 'notehub -signin' to sign into the notehub service", Config.Hub)
		return
	}

	// Set the header
	httpReq.Header.Set("X-Session-Token", token)

	// Done
	return

}

// ConfigAPIHub returns the configured notehub, for use by the HTTP API.  If none is configured it returns
// the default Blues API service.  Regardless, it always makes sure that the host has "api." as a prefix.
// This enables flexibility in what's configured.
func ConfigAPIHub() (hub string) {
	hub = Config.Hub
	if hub == "" || hub == "-" {
		hub = notehub.DefaultAPIService
	}
	if !strings.HasPrefix(hub, "api.") {
		hub = "api." + hub
	}
	return
}

// ConfigNotecardHub returns the configured notehub, for use as the Notecard host.  If none is configured
// it returns "".  Regardless, it always makes sure that the host does NOT have "api." as a prefix.
func ConfigNotecardHub() (hub string) {
	hub = Config.Hub
	if hub == "" || hub == "-" {
		hub = notehub.DefaultAPIService
	}
	if strings.HasPrefix(hub, "api.") {
		hub = strings.TrimPrefix(hub, "api.")
	}
	return
}

// ConfigSetHub clears the hub
func ConfigSetHub(hub string) {
	if hub == "-" {
		hub = ""
	}
	Config.Hub = hub
}
