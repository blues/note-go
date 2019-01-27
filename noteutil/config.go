// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package noteutil

import (
    "os"
    "fmt"
    "flag"
    "time"
	"strconv"
    "io/ioutil"
    "encoding/json"
	"github.com/rayozzie/note-go/notecard"
	"github.com/rayozzie/note-go/notehub"
)

// ConfigSettings defines the config file that maintains the command processor's state
type ConfigSettings struct {
    When            string      `json:"when,omitempty"`
    Hub             string      `json:"hub,omitempty"`
    App             string      `json:"app,omitempty"`
    Device          string      `json:"device,omitempty"`
    Product         string      `json:"product,omitempty"`
    Root            string      `json:"root,omitempty"`
    Cert            string      `json:"cert,omitempty"`
    Key             string      `json:"key,omitempty"`
    Secure          bool        `json:"secure,omitempty"`
    Interface       string      `json:"interface,omitempty"`
    Port            string      `json:"port,omitempty"`
    PortConfig      int         `json:"port_config,omitempty"`
}

// Config is the active copy of our configuration file, never dirty.
var flagConfigReset bool
var flagConfigSave bool
var Config = ConfigSettings{}

// configRead reads the current info from config file
func ConfigRead() error {

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

// configWrite updates the file with the current config info
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

// configReset updates the file with the default info
func ConfigReset() {
    Config.Interface = "serial"
    Config.Port, Config.PortConfig = notecard.PortDefaults(Config.Interface)
    Config.Hub = notehub.DefaultAPIService
    Config.When = time.Now().UTC().Format("2006-01-02T15:04:05Z")
    return
}

// configShow displays all current config parameters
func ConfigShow() error {

    fmt.Printf("\nCurrently saved values:\n")

    if Config.Hub != "" && Config.Hub != notehub.DefaultAPIService {
        if (Config.Secure) {
            fmt.Printf("   -https\n")
        } else {
            fmt.Printf("   -http\n")
        }
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
	    fmt.Printf("  -interface %s\n", Config.Interface)
		port, portConfig := notecard.PortDefaults(Config.Interface)
		if port == "" {
			port = "-"
		}
	    fmt.Printf("  -port %s\n", port)
	    fmt.Printf("  -portconfig %d\n", portConfig)
	}

    return nil

}

// ConfigFlagsProcess processes the registered config flags
func ConfigFlagsProcess() (err error) {

	if Config.When == "" {
		err = ConfigRead()
		if err != nil {
			return
		}
	}

	if flagConfigReset {
		ConfigReset()
	}

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
    flag.StringVar(&Config.Interface, "interface", Config.Interface, "select 'serial' or 'i2c' interface")
    flag.StringVar(&Config.Port, "port", Config.Port, "select serial or i2c port")
    flag.IntVar(&Config.PortConfig, "portconfig", Config.PortConfig, "set serial device speed or i2c address")

	// Write the config if asked to do so
    flag.BoolVar(&flagConfigSave, "config-save", false, "save changes to note tool config")
	
}
