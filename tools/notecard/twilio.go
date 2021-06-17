// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func twilioProvision(creds string, iccid string, key string) (err error) {
	var httpReq *http.Request
	var httpRsp *http.Response
	var rspJSON []byte
	httpClient := &http.Client{}

	// Prefix all error returns
	prefix := "twilio"
	defer func(p string) {
		if err != nil {
			err = fmt.Errorf("%s: %s", p, err)
		}
	}(prefix)

	// Get the existing device.  If it's there, we're done.
	httpURL := fmt.Sprintf("https://%s@supersim.twilio.com/v1/Sims?Iccid=%s", creds, iccid)
	httpReq, err = http.NewRequest("GET", httpURL, nil)
	if err != nil {
		return
	}
	httpReq.Header.Set("User-Agent", "notecard utility")
	httpRsp, err = httpClient.Do(httpReq)
	if err != nil {
		return
	}
	if httpRsp.StatusCode == http.StatusOK {
		rspJSON, err = ioutil.ReadAll(httpRsp.Body)
		if err == nil {
			rsp := map[string]interface{}{}
			err = json.Unmarshal(rspJSON, &rsp)
			if err != nil {
				err = fmt.Errorf("unmarshaling GET response: %s", err)
				return
			}
			sims, present := rsp["sims"]
			if present {
				simArray := sims.([]interface{})
				if len(simArray) > 0 {
					s := simArray[0].(map[string]interface{})
					simname := ""
					if s["unique_name"] != nil {
						simname = s["unique_name"].(string)
					}
					fmt.Printf("%s: device already provisioned: status:%s %s\n", prefix, s["status"], simname)
					return
				}
			}
		}
	}

	// Provision the SIM
	body := url.Values{}
	body.Set("Iccid", iccid)
	body.Set("RegistrationCode", key)
	bodyEncoded := body.Encode()
	httpURL = fmt.Sprintf("https://%s@supersim.twilio.com/v1/Sims", creds)
	httpReq, err = http.NewRequest(http.MethodPost, httpURL, strings.NewReader(bodyEncoded))
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Content-Length", strconv.Itoa(len(bodyEncoded)))
	httpRsp, err = httpClient.Do(httpReq)
	if err != nil {
		return
	}
	if httpRsp.StatusCode != http.StatusOK && httpRsp.StatusCode != http.StatusCreated {
		fmt.Printf("status: %d %s\n", httpRsp.StatusCode, httpRsp.Status)
		rspJSON, _ = ioutil.ReadAll(httpRsp.Body)
		fmt.Printf("%s\n", rspJSON)
		rsp := map[string]interface{}{}
		err = json.Unmarshal(rspJSON, &rsp)
		if err != nil {
			return
		}
		if rsp["message"] != nil {
			err = fmt.Errorf("%s", rsp["message"].(string))
			return
		}
	}
	fmt.Printf("\n%s: successfully provisioned:\n%s\n", prefix, rspJSON)
	return

}
