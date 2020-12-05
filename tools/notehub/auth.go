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
	"net/http"
	"os"
	"strings"

	"github.com/blues/note-go/noteutil"
	"golang.org/x/crypto/ssh/terminal"
)

// Sign into the notehub account
func authSignIn() (err error) {

	// Print banner
	fmt.Printf(banner())

	// Read the account
	var username string
	for username == "" {
		fmt.Printf("account email@address.com > ")
		scanner := bufio.NewScanner(os.Stdin)
		ok := scanner.Scan()
		if ok {
			username = strings.TrimRight(scanner.Text(), "\r\n")
		}
	}

	// Read the password
	var password string
	for password == "" {
		fmt.Printf("account password > ")
		var pw []byte
		pw, err = terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Printf("\n")
		if err != nil {
			return
		}
		password = string(pw)
	}

	// Do the sign-in HTTP request
	req := map[string]interface{}{}
	req["username"] = username
	req["password"] = password
	reqJSON, err2 := json.Marshal(req)
	if err2 != nil {
		err = err2
		return
	}
	httpURL := "https://" + noteutil.ConfigAPIHub() + "/auth/login"
	httpReq, err2 := http.NewRequest("POST", httpURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		err = err2
		return
	}
	httpReq.Header.Set("User-Agent", "notehub-client")
	httpReq.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	httpRsp, err2 := httpClient.Do(httpReq)
	if err2 != nil {
		err = err2
		return
	}
	if httpRsp.StatusCode == http.StatusUnauthorized {
		err = fmt.Errorf("unrecognized username or password")
		return
	}
	rspJSON, err2 := ioutil.ReadAll(httpRsp.Body)
	if err2 != nil {
		err = err2
		return
	}
	rsp := map[string]interface{}{}
	err = json.Unmarshal(rspJSON, &rsp)
	if err != nil {
		return
	}
	token := ""
	if rsp["session_token"] != nil {
		token = rsp["session_token"].(string)
	}
	if token == "" {
		err = fmt.Errorf("%s authentication error", noteutil.ConfigAPIHub())
		return
	}

	// Extract the token and save it
	noteutil.ConfigRead()
	noteutil.Config.Token = token
	noteutil.Config.TokenUser = username
	err = noteutil.ConfigWrite()
	if err != nil {
		return
	}

	// Done
	fmt.Printf("signed in successfully as %s\n", username)
	return

}

// Sign out of the API
func authSignOut() (err error) {
	if noteutil.Config.Token == "" || noteutil.Config.TokenUser == "" {
		fmt.Printf("not currently signed in\n")
		return
	}
	// TODO hit the logout endpoint in the API to revoke the session
	fmt.Printf("%s signed out successfully\n", noteutil.Config.TokenUser)
	noteutil.ConfigRead()
	noteutil.Config.Token = ""
	err = noteutil.ConfigWrite()
	if err != nil {
		return
	}
	return
}

// Banner for authentication
// http://patorjk.com/software/taag
// "Big" font

func banner() (s string) {
	s += "             _       _           _       \r\n"
	s += "            | |     | |         | |      \r\n"
	s += " _ __   ___ | |_ ___| |__  _   _| |__    \r\n"
	s += "| '_ \\ / _ \\| __/ _ \\ '_ \\| | | | '_ \\   \r\n"
	s += "| | | | (_) | ||  __/ | | | |_| | |_) |  \r\n"
	s += "|_| |_|\\___/ \\__\\___|_| |_|\\__,_|_.__/   \r\n"
	s += "\r\n"
	return
}
