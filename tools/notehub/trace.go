// Copyright 2021 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/blues/note-go/noteutil"
)

// Command definitions
type cmdDef struct {
	Command string
	Desc    string
}

func validCommands() []cmdDef {
	return []cmdDef{{"product", "set productUID for requests made in this session"},
		{"project", "set projectUID (appUID) for requests made in this session"},
		{"device", "set deviceUID for requests made in this session"},
		{"hub", "set notehub domain for requests made in this session"},
		{"get", "HTTPS GET from specified URL"},
		{"put", "HTTPS PUT to specified URL of the JSON prompted for on next line"},
		{"post", "HTTPS POST of specified URL of the JSON prompted for on next line"},
		{"delete", "HTTPS DELETE of resource at specified URL"},
		{"ping", "ping the notehub"},
		{"q", "quit"},
	}
}

// Enter a diagnostic trace mode
func trace() error {

	// Create a scanner to watch stdin
	scanner := bufio.NewScanner(os.Stdin)
	var cmd string

traceloop:
	for {
		// Get next text line
		fmt.Print("> ")
		scanner.Scan()
		cmd = scanner.Text()

		// Parse into arguments
		r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
		args := r.FindAllString(cmd, -1)
		for i := 0; i < 10; i++ {
			args = append(args, "")
		}
		cmdAfter0 := strings.TrimPrefix(cmd, args[0]+" ")

		// Process JSON requests
		if strings.HasPrefix(cmd, "{") {
			_, err := reqHubJSON(true, noteutil.ConfigAPIHub(), []byte(cmd), "", "", "", "", false, false, nil)
			if err != nil {
				fmt.Printf("error: %s\n", err)
			}
			continue
		}

		// Create clean IDs to work with in the commands
		cleanProduct := flagProduct
		if cleanProduct != "" && !strings.HasPrefix(cleanProduct, "product:") {
			cleanProduct = "product:" + flagProduct
		}
		cleanApp := flagApp
		if !strings.HasPrefix(cleanApp, "app:") {
			if cleanApp == "" {
				cleanApp = cleanProduct
			} else {
				cleanApp = "app:" + flagApp
			}
		}
		cleanDevice := flagDevice
		if !strings.HasPrefix(cleanDevice, "dev:") {
			cleanDevice = "dev:" + flagDevice
		}
		cmdAfter0 = strings.Replace(cmdAfter0, "{productUID}", cleanProduct, -1)
		cmdAfter0 = strings.Replace(cmdAfter0, "{projectUID}", cleanApp, -1)
		cmdAfter0 = strings.Replace(cmdAfter0, "{deviceUID}", cleanDevice, -1)

		// Dispatch command
		switch args[0] {
		case "?":
			fmt.Printf("Trace commands:\n")
			for _, c := range validCommands() {
				fmt.Printf("%s: %s\n", c.Command, c.Desc)
			}
			fmt.Printf("{\"req\":...} for a JSON request\n")
		case "product":
			if args[1] != "" {
				if args[1] == "-" {
					args[1] = ""
				}
				flagProduct = args[1]
			}
			fmt.Printf("productUID is %s\n", flagProduct)

		case "project":
			fallthrough
		case "app":
			if args[1] != "" {
				if args[1] == "-" {
					args[1] = ""
				}
				flagApp = args[1]
			}
			fmt.Printf("projectUID is %s\n", flagApp)

		case "device":
			if args[1] != "" {
				if args[1] == "-" {
					args[1] = ""
				}
				flagDevice = args[1]
			}
			fmt.Printf("deviceUID is %s\n", flagDevice)

		case "hub":
			if args[1] != "" {
				if args[1] == "-" {
					args[1] = ""
				}
				noteutil.Config.Hub = args[1]
			}
			fmt.Printf("hub is %s\n", flagApp)

		case "get":
			fallthrough
		case "delete":
			fallthrough
		case "put":
			fallthrough
		case "post":
			// Get the body to post/put
			var bodyJSON []byte
			if args[0] == "put" || args[0] == "post" {
				fmt.Print("JSON> ")
				scanner.Scan()
				bodyJSON = []byte(scanner.Text())
			}

			// Make sure that it's a well-formed URL for our API
			url := cmdAfter0
			if !strings.HasPrefix(url, "/") {
				url = "/" + url
			}
			if !strings.HasPrefix(url, "/v1/") {
				url = "/v1" + url
			}

			// Perform the transaction
			_, err := reqHubHTTP(true, noteutil.ConfigAPIHub(), args[0], url, bodyJSON)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return err
			}
		case "ping":
			_, err := reqHubHTTP(true, noteutil.ConfigAPIHub(), "GET", "/ping", nil)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return err
			}
			if cleanApp != "" {
				url := "/v1/products/" + cleanApp + "/products"
				_, err = reqHubHTTP(true, noteutil.ConfigAPIHub(), "GET", url, nil)
				if err != nil {
					fmt.Printf("error: %s\n", err)
					return err
				}
			}
		case "":
			break
		case "q":
			break traceloop
		default:
			fmt.Printf("%s ???\n", args[0])
		}
	}
	return nil
}

// Process an HTTPS request
func reqHubHTTP(verbose bool, hub string, verb string, url string, body []byte) (response []byte, err error) {

	verb = strings.ToUpper(verb)

	httpurl := fmt.Sprintf("https://%s%s", hub, url)
	buffer := &bytes.Buffer{}
	if body != nil {
		buffer = bytes.NewBuffer(body)
	}
	httpReq, err := http.NewRequest(verb, httpurl, buffer)
	if err != nil {
		return
	}
	httpReq.Header.Set("User-Agent", "notehub-client")
	httpReq.Header.Set("Content-Type", "application/json")
	err = noteutil.ConfigAuthenticationHeader(httpReq)
	if err != nil {
		return
	}

	if verbose {
		fmt.Printf("%s %s\n", verb, httpurl)
		if len(body) != 0 {
			fmt.Printf("%s\n", string(body))
		}
	}

	httpClient := &http.Client{}
	httpRsp, err2 := httpClient.Do(httpReq)
	if err2 != nil {
		err = err2
		return
	}
	if httpRsp.StatusCode == http.StatusUnauthorized {
		err = fmt.Errorf("please use -signin to authenticate")
		return
	}

	if verbose {
		fmt.Printf("STATUS %d\n", httpRsp.StatusCode)
	}

	var rspJSON []byte
	rspJSON, err = ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return
	}

	if verbose && len(rspJSON) != 0 {
		fmt.Printf("%s\n", string(rspJSON))
	}

	return

}
