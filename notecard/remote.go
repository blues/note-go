// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/blues/note-go/note"
)

// The number of minutes that we'll round up so that card reservations don't thrash
const reservationModulusMinutes = 5

// RemoteCard is the full description of notecards managed by the farm
type RemoteCard struct {
	Instance        string `json:"instance,omitempty"`
	Address         int    `json:"address,omitempty"`
	DirectURL       string `json:"direct,omitempty"`
	ProxyURL        string `json:"proxy,omitempty"`
	Version         string `json:"version,omitempty"`
	DeviceUID       string `json:"device,omitempty"`
	ProductUID      string `json:"product,omitempty"`
	SN              string `json:"sn,omitempty"`
	Transactions    uint32 `json:"transactions,omitempty"`
	TransactionTime uint32 `json:"transaction_time,omitempty"`
	Refreshed       uint32 `json:"refreshed,omitempty"`
}

// RemoteCards are the objects that get published
type RemoteCards struct {
	Cards []RemoteCard `json:"notecards,omitempty"`
}

// Get the default remote farm and notecard checkout period
func remotePortDefault() (farmURL string, farmCheckoutMins int) {
	return
}

// Enumerate remote farms
func remotePortEnum() (allFarms []string, unused []string, notecardFarms []string, err error) {
	return
}

// Reset communications with the remote notecard
func remoteReset(context *Context) (err error) {
	return
}

// Close a remote notecard
func remoteClose(context *Context) {
	return
}

// Get the CallerID for this requestor, increasing the likelihood of getting the same
// reservation between tests which may be run across different machines and across
// different processes on the same machine.
func callerID() (id string) {

	// Get the mac address
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				// Don't use random as we have a real address
				id = i.HardwareAddr.String()
				break
			}
		}
	}

	// Append the parent process ID
	id += fmt.Sprintf(":%d", os.Getppid())

	return
}

// Get the caller ID with expiration
func callerIDWithExpiration(expires int64) string {
	return fmt.Sprintf("%s=%d", callerID(), expires)
}

// Get the timeout from the caller ID
func extractCallerID(sn string) (callerid string, expires int64) {
	c := strings.Split(sn, "=")
	if len(c) == 1 {
		return
	}
	n, err := strconv.ParseInt(c[len(c)-1], 10, 64)
	if err != nil {
		return
	}
	callerid = c[0]
	expires = n
	return
}

// Get the remote notecard list
func cardList(context *Context) (cards []RemoteCard, err error) {

	notefarm := os.Getenv("NOTEFARM")
	if context.farmURL != "" {
		notefarm = context.farmURL
	}
	if notefarm == "" {
		err = fmt.Errorf("cannot use remote notecards without hosting a NOTEFARM")
		return
	}

	req, err1 := http.NewRequest("GET", notefarm, nil)
	if err1 != nil {
		err = err1
		return
	}

	httpclient := &http.Client{Timeout: time.Second * 90}
	resp, err2 := httpclient.Do(req)
	if err2 != nil {
		err = fmt.Errorf("notefarm: can't get device list: %s", err2)
		return
	}

	rspbuf, err3 := ioutil.ReadAll(resp.Body)
	if err3 != nil {
		err = fmt.Errorf("notefarm: can't read device list: %s", err3)
		return
	}

	remoteCards := RemoteCards{}
	err3 = note.JSONUnmarshal(rspbuf, &remoteCards)
	if err3 != nil {
		// If a web URL is mistakenly used, clarify the error
		if strings.Contains(string(rspbuf), "DOCTYPE") {
			err = fmt.Errorf("notefarm: not a device list")
			return
		}
		err = fmt.Errorf("notefarm: can't unmarshal device list: %sn", err3)
		return
	}

	if len(remoteCards.Cards) == 0 {
		err = fmt.Errorf("notefarm: empty farm")
		return
	}

	cards = remoteCards.Cards
	return

}

// Open or reopen the remote card
func remoteReopen(context *Context) (err error) {

	// Wait indefinitely for a reservation
	for {

		// Read the device list from the farm
		cards, err0 := cardList(context)
		if err0 != nil {
			err = err0
			return
		}

		// Look at the cards for a prior reservation, because we only allow a single
		// notecard for each caller
		ourCallerID := callerID()
		ourCard := RemoteCard{}
		for _, c := range cards {
			cid, expires := extractCallerID(c.SN)
			if cid == ourCallerID {
				// We don't need to reserve the card if it expires after what we need
				if context.farmCheckoutExpires < expires {
					context.farmCard = c
					now := time.Now().Unix()
					secs := int(expires-now) % 60
					mins := int(expires-now) / 60
					fmt.Printf("%s reserved for %dm %2ds\n", c.DeviceUID, mins, secs)
					return
				}
				ourCard = c
				break
			}
		}

		// If we didn't find it, get the LRU card that hasn't expired
		first := true
		oursExpires := int64(0)
		now := time.Now().Unix()
		if ourCard.SN == "" {
			for _, c := range cards {
				_, expires := extractCallerID(c.SN)
				if oursExpires > now {
					continue
				}
				if first || expires < oursExpires {
					first = false
					ourCard = c
					oursExpires = expires
				}
			}
			if first {
				err = fmt.Errorf("notefarm: all cards are currently reserved")
				return
			}
		}

		// On an interim basis claim the card
		context.farmCard = ourCard

		// Reserve the card
		req := Request{Req: "service.set"}
		checkoutMins := ((context.farmCheckoutMins / reservationModulusMinutes) + 1) * reservationModulusMinutes
		req.SN = callerIDWithExpiration(now + int64(checkoutMins*60))
		reqJSON, err1 := note.ObjectToJSON(req)
		if err1 != nil {
			err = err1
			return
		}
		_, err = remoteTransaction(context, reqJSON)
		if err != nil {
			err = fmt.Errorf("notefarm reservation error: %s", err)
			return
		}

		// Wait until the reservation succeeds
		for i := 0; i < 10; i++ {

			time.Sleep(2 * time.Second)

			cards, err = cardList(context)
			if err != nil {
				return
			}

			for _, c := range cards {
				if c.DeviceUID == ourCard.DeviceUID {
					if c.SN == req.SN {
						fmt.Printf("%s reserved for %d minutes\n", c.DeviceUID, checkoutMins)
						return
					}
				}
			}

			fmt.Printf("waiting for reservation confirmation\n")

		}

		// We no longer have the card
		context.farmCard = RemoteCard{}

	}

	return
}

// Perform a remote transaction
func remoteTransaction(context *Context, reqJSON []byte) (rspJSON []byte, err error) {

	// If our reservation has expired, fail the transaction
	if time.Now().Unix() > context.farmCheckoutExpires {
		err = fmt.Errorf("notefarm reservation of %d min has expired", context.farmCheckoutMins)
		return
	}

	// Perform the transaction several times to cover the case of exceeding the
	// transaction rate of the notefarm's proxy infrastructure
	var rspbuf []byte
	var resp *http.Response
	var maxRetries = 5
	for i := 0; ; i++ {

		var req *http.Request
		req, err = http.NewRequest("POST", context.farmCard.DirectURL, bytes.NewBuffer(reqJSON))
		if err != nil {
			rspJSON = []byte(fmt.Sprintf("{\"err\":\"%s\"}", err))
			break
		}

		httpclient := &http.Client{Timeout: time.Second * 90}
		resp, err = httpclient.Do(req)
		if err != nil {
			rspJSON = []byte(fmt.Sprintf("{\"err\":\"%s\"}", err))
			break
		}

		rspbuf, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			rspJSON = []byte(fmt.Sprintf("{\"err\":\"%s\"}", err))
			break
		}

		// Validate that it's compliant JSON
		var jobj map[string]interface{}
		err = note.JSONUnmarshal(rspbuf, &jobj)
		if err == nil {

			// See if there was an I/O error to the card, and retry if so
			if !strings.Contains(string(rspbuf), "{io}") {
				rspJSON = rspbuf
				break
			}
			if i > maxRetries {
				rspJSON = []byte(fmt.Sprintf("{\"err\":\"proxy: cannot communicate with notecard {io}\"}"))
				break
			}

		} else {

			if i > maxRetries {
				rspJSON = []byte(fmt.Sprintf("{\"err\":\"%s\"}", err))
				break
			}

		}

		// Sleep in case we're in the penalty box for too high of a transaction rate
		time.Sleep(2 * time.Second)

	}

	return

}
