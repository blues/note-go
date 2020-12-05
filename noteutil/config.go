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

// ConfigSettings defines the config file that maintains the command processor's state
type ConfigSettings struct {
	When       string `json:"when,omitempty"`
	Hub        string `json:"hub,omitempty"`
	App        string `json:"project,omitempty"`
	Product    string `json:"product,omitempty"`
	Device     string `json:"device,omitempty"`
	Interface  string `json:"interface,omitempty"`
	TokenUser  string `json:"token_user,omitempty"`
	Token      string `json:"token,omitempty"`
	Port       string `json:"port,omitempty"`
	PortConfig int    `json:"port_config,omitempty"`
}

// Config is the active copy of our configuration file, never dirty.
var flagConfigReset bool
var flagConfigSave bool

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

	if Config.TokenUser != "" && Config.Token != "" {
		fmt.Printf("   account: %s\n", Config.TokenUser)
	}
	if Config.Hub != "" {
		fmt.Printf("       hub: %s\n", Config.Hub)
	}
	if Config.App != "" {
		fmt.Printf("   project: %s\n", Config.App)
	}
	if Config.Product != "" {
		fmt.Printf("   product: %s\n", Config.Product)
	}
	if Config.Device != "" {
		fmt.Printf("    device: %s\n", Config.Device)
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

	// Reset if requested
	if flagConfigReset {
		ConfigReset()
	}

	// Set the flags as desired
	if configFlags.Hub != "" {
		ConfigSetHub(configFlags.Hub)
	}
	if configFlags.App == "-" {
		Config.App = ""
	} else if configFlags.App != "" {
		Config.App = configFlags.App
	}
	if configFlags.Product == "-" {
		Config.Product = ""
	} else if configFlags.Product != "" {
		Config.Product = configFlags.Product
	}
	if configFlags.Device == "-" {
		Config.Device = ""
	} else if configFlags.Device != "" {
		Config.Device = configFlags.Device
	}
	if configFlags.Interface == "-" {
		configResetInterface()
	} else if configFlags.Interface != "" {
		Config.Interface = configFlags.Interface
	}
	if configFlags.Port != "" {
		Config.Port = configFlags.Port
	}
	if configFlags.PortConfig < 0 {
		Config.PortConfig = 0
	} else if configFlags.PortConfig != 0 {
		Config.PortConfig = configFlags.PortConfig
	}

	// Save if requested
	if flagConfigSave {
		ConfigWrite()
		ConfigShow()
	}

	// Override, just for this session, with env vars
	str := os.Getenv("NOTE_INTERFACE")
	if str != "" {
		Config.Interface = str
	}
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
		flag.StringVar(&configFlags.Device, "device", "", "set DeviceUID")
		if false { // no longer necessary because of universal support for productUID
			flag.StringVar(&configFlags.App, "project", "", "set AppUID")
		}
		flag.StringVar(&configFlags.Product, "product", "", "set ProductUID")
		flag.StringVar(&configFlags.Hub, "hub", "", "set notehub domain")
	}

	// Write the config if asked to do so
	flag.BoolVar(&flagConfigReset, "config-reset", false, "reset the cross-tool saved configuration to its defaults")
	flag.BoolVar(&flagConfigSave, "config-save", false, "save changes to the cross-tool saved configuration")

}

// FlagParse is a wrapper around flag.Parse that handles our config flags
func FlagParse(notecardFlags bool, notehubFlags bool) (err error) {
	ConfigFlagsRegister(notecardFlags, notehubFlags)
	flag.Parse()
	return ConfigFlagsProcess()
}

// ConfigSignedIn returns info about whether or not we're signed in
func ConfigSignedIn() (username string, authenticated bool) {

	if Config.Token != "" && Config.TokenUser != "" {
		authenticated = true
		username = Config.TokenUser
	}

	return

}

// ConfigAuthenticationHeader sets the authorization field in the header as appropriate
func ConfigAuthenticationHeader(httpReq *http.Request) (err error) {

	// Read config if not yet read
	if Config.When == "" {
		err = ConfigRead()
		if err != nil {
			return
		}
	}

	// Exit if not signed in
	if Config.Token == "" || Config.TokenUser == "" {
		err = fmt.Errorf("not authenticated: please use 'notehub -signin' to sign into the notehub service")
		return
	}

	// Set the header
	httpReq.Header.Set("X-Session-Token", Config.Token)

	// Done
	return

}

// ConfigAPIHub returns the configured notehub, for use by the HTTP API.  If none is configured it returns
// the default Blues API service.  Regardless, it always makes sure that the host has "api." as a prefix.
// This enables flexibility in what's configured.
func ConfigAPIHub() (hub string) {
	hub = Config.Hub
	if hub == "" {
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
	if strings.HasPrefix(hub, "api.") {
		hub = strings.TrimPrefix(hub, "api.")
	}
	return
}

// ConfigSetHub clears the hub
func ConfigSetHub(hub string) {
	if Config.Hub == "-" {
		hub = ""
	}
	Config.Hub = hub
}
