// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package noteutil

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/blues/note-go/notehub"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// ConfigSettings defines the config file that maintains the command processor's state
type ConfigSettings struct {
	When       string `json:"when,omitempty"`
	Hub        string `json:"hub,omitempty"`
	App        string `json:"app,omitempty"`
	Device     string `json:"device,omitempty"`
	Root       string `json:"root,omitempty"`
	Cert       string `json:"cert,omitempty"`
	Key        string `json:"key,omitempty"`
	Secure     bool   `json:"secure,omitempty"`
	Interface  string `json:"interface,omitempty"`
	Port       string `json:"port,omitempty"`
	PortConfig int    `json:"port_config,omitempty"`
}

// Config is the active copy of our configuration file, never dirty.
var flagConfigReset bool
var flagConfigSave bool
var flagConfigHTTP bool
var flagConfigHTTPS bool

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
		err = json.Unmarshal(contents, &Config)
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
	configJSON, _ := json.MarshalIndent(Config, "", "    ")

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
	Config.Interface = ""
	Config.Port = ""
	Config.PortConfig = 0
}

// ConfigReset updates the file with the default info
func ConfigReset() {
	configResetInterface()
	Config.Hub = notehub.DefaultAPIService
	Config.When = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	return
}

// ConfigShow displays all current config parameters
func ConfigShow() error {

	fmt.Printf("\nCurrently saved values:\n")

	if Config.Secure {
		fmt.Printf("   -https\n")
	} else {
		fmt.Printf("   -http\n")
	}

	if Config.Hub != "" && Config.Hub != notehub.DefaultAPIService {
		fmt.Printf("   -hub %s\n", Config.Hub)
	}
	if Config.App != "" {
		fmt.Printf("   -app %s\n", Config.App)
	}
	if Config.Device != "" {
		fmt.Printf("   -device %s\n", Config.Device)
	}
	if Config.Root != "" {
		fmt.Printf("   -root %s\n", Config.Root)
	}
	if Config.Cert != "" {
		fmt.Printf("   -cert %s\n", Config.Cert)
	}
	if Config.Key != "" {
		fmt.Printf("   -key %s\n", Config.Key)
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
	if flagConfigHTTP {
		configFlags.Secure = false
		Config.Secure = false
	}
	if flagConfigHTTPS {
		configFlags.Secure = true
		Config.Secure = true
	}
	if configFlags.Hub == "-" {
		Config.Hub = notehub.DefaultAPIService
	} else if configFlags.Hub != "" {
		Config.Hub = configFlags.Hub
	}
	if configFlags.Root == "-" {
		Config.Root = ""
	} else if configFlags.Root != "" {
		Config.Root = configFlags.Root
	}
	if configFlags.Key == "-" {
		Config.Key = ""
	} else if configFlags.Key != "" {
		Config.Key = configFlags.Key
	}
	if configFlags.Cert == "-" {
		Config.Cert = ""
	} else if configFlags.Cert != "" {
		Config.Cert = configFlags.Cert
	}
	if configFlags.App == "-" {
		Config.App = ""
	} else if configFlags.App != "" {
		Config.App = configFlags.App
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
	if configFlags.PortConfig != -1 {
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
func ConfigFlagsRegister() {

	// Start by setting to default if requested
	flag.BoolVar(&flagConfigReset, "config-reset", false, "reset the note tool config to its defaults")

	// Process the commands
	flag.StringVar(&configFlags.Interface, "interface", "", "select 'serial' or 'i2c' interface")
	flag.StringVar(&configFlags.Port, "port", "", "select serial or i2c port")
	flag.IntVar(&configFlags.PortConfig, "portconfig", -1, "set serial device speed or i2c address")
	flag.BoolVar(&flagConfigHTTP, "http", false, "use http instead of https")
	flag.BoolVar(&flagConfigHTTPS, "https", false, "use https instead of http")
	flag.StringVar(&configFlags.Hub, "hub", "", "set notehub request service URL")
	flag.StringVar(&configFlags.Device, "device", "", "set DeviceUID")
	flag.StringVar(&configFlags.App, "app", "", "set AppUID (the Project UID)")
	flag.StringVar(&configFlags.Root, "root", "", "set path to service's root CA certificate file")
	flag.StringVar(&configFlags.Key, "key", "", "set path to local private key file")
	flag.StringVar(&configFlags.Cert, "cert", "", "set path to local cert file")

	// Write the config if asked to do so
	flag.BoolVar(&flagConfigSave, "config-save", false, "save changes to note tool config")

}

// FlagParse is a wrapper around flag.Parse that handles our config flags
func FlagParse() (err error) {
	ConfigFlagsRegister()
	flag.Parse()
	return ConfigFlagsProcess()
}
