// Copyright 2019 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package noteutil

import (
    "os"
    "fmt"
    "time"
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
var Config = ConfigSettings{}

// configRead reads the current info from config file
func ConfigRead() error {

    // Read the file
    contents, err := ioutil.ReadFile(configSettingsPath())

    // Unmarshal if no error
    if err == nil {
        err = json.Unmarshal(contents, &Config)
    }

    // If error reading or unmarshaling, just reinitialize it
    if err != nil || Config.When == "" {

        // Reset the configuration
        err = ConfigReset()
        if err != nil {
            err = fmt.Errorf("can't read configuration: %s", err)
        }

    }

    // Done
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
func ConfigReset() error {
    Config = ConfigSettings{}
    Config.Interface = "serial"
    Config.Port, Config.PortConfig = notecard.PortDefaults(Config.Interface)
    Config.Hub = notehub.DefaultAPIService
    Config.When = time.Now().UTC().Format("2006-01-02T15:04:05Z")
    return nil
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
