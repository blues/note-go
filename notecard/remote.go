// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blues/note-go/note"
	"github.com/gofrs/flock"
)

// RemoteCard is the full description of notecards managed by the farm
type RemoteCard struct {
	Instance    string `json:"instance,omitempty"`
	Address     int    `json:"address,omitempty"`
	DirectURL   string `json:"direct,omitempty"`
	ProxyURL    string `json:"proxy,omitempty"`
	Version     string `json:"version,omitempty"`
	CardTest    string `json:"card,omitempty"`
	DeviceUID   string `json:"device,omitempty"`
	ProductUID  string `json:"product,omitempty"`
	SN          string `json:"sn,omitempty"`
	Reservation string `json:"reservation,omitempty"`
	Modified    uint32 `json:"modified,omitempty"`
	Refreshed   uint32 `json:"refreshed,omitempty"`
}

// RemoteCards are the objects that get published
type RemoteCards struct {
	Cards []RemoteCard `json:"notecards,omitempty"`
}

// Close a remote notecard
func remoteClose(context *Context) {
	// Reset the remote card to release the reservation

	// 'https://DirectURL&reset=true'
	var req *http.Request
	resetURL := context.farmCard.DirectURL + `&reset=true`

	req, err := http.NewRequest("GET", resetURL, nil)
	if err != nil {
		if context.Debug {
			fmt.Printf("notefarm: remoteClose NewReq Fail %v", err)
		}
		return
	}
	httpclient := &http.Client{Timeout: time.Second * 90}
	resp, err := httpclient.Do(req)
	if err == nil {
		defer resp.Body.Close()
	} else {
		if context.Debug {
			fmt.Printf("notefarm: remoteClose http.Do Fail %v", err)
		}
	}

	context.portIsOpen = false

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
	if err2 == nil {
		defer resp.Body.Close()
	} else {
		err = fmt.Errorf("notefarm: can't get device list: %s", err2)
		return
	}

	var buf bytes.Buffer
	_, err3 := io.Copy(&buf, resp.Body)
	if err3 != nil {
		err = fmt.Errorf("notefarm: can't read device list: %s", err3)
		return
	}
	rspbuf := buf.Bytes()

	remoteCards := RemoteCards{}
	err3 = note.JSONUnmarshal(rspbuf, &remoteCards)
	if err3 != nil {
		// If a web URL is mistakenly used, clarify the error
		if strings.Contains(string(rspbuf), "DOCTYPE") {
			err = fmt.Errorf("notefarm: not a device list")
			return
		}
		err = fmt.Errorf("notefarm: can't unmarshal device list: %s", err3)
		return
	}

	if len(remoteCards.Cards) == 0 {
		err = fmt.Errorf("notefarm: empty farm: document: %s", url.QueryEscape(string(rspbuf)))
		return
	}

	cards = remoteCards.Cards
	return
}

// Open or reopen the remote card. Locked to prevent multiple processes on this
// machine from stepping on eachother's toes. Doesn't prevent us from stepping
// on the toes of processes running on different machines.
func remoteReopen(context *Context, portConfig int) (err error) {

	context.portIsOpen = false

	// Get Mutex file lock to prevent a race with other processes on this machine.
	fileLock := flock.New(filepath.Join(os.TempDir(), "notefarm.lock"))
	err = fileLock.Lock()
	if err != nil {
		err = fmt.Errorf("notefarm reservation error: can not lock [%v]", fileLock.Path())
		return
	}

	err = uRemoteReopen(context, portConfig)

	err2 := fileLock.Unlock()
	if err2 != nil {
		err = fmt.Errorf("notefarm reservation error: can not unlock [%v]: %s; inner error: %w",
			fileLock.Path(), err2, err)
	}

	if err == nil {
		context.portIsOpen = true
	}

	return
}

// Open or reopen the remote card. Unlocked.
func uRemoteReopen(context *Context, portConfig int) (err error) {
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
			cid, expires := extractCallerID(c.Reservation)
			if cid == ourCallerID {
				// We don't need to reserve the card if it expires after what we need
				if context.farmCheckoutExpires <= expires {
					context.farmCard = c
					now := time.Now().Unix()
					secs := int(expires-now) % 60
					mins := int(expires-now) / 60
					if context.Debug {
						fmt.Printf("notefarm: %s reserved for %dm %2ds\n", c.DeviceUID, mins, secs)
					}
					return
				}
				ourCard = c
				if context.Debug {
					fmt.Printf("notefarm: trying to extend our reservation of %s by %d seconds\n", ourCard.DeviceUID, context.farmCheckoutExpires-expires)
				}
				break
			}
		}

		// If we didn't find it, get the LRU card that hasn't expired
		first := true
		oursExpires := int64(0)
		now := time.Now().Unix()
		if ourCard.Reservation == "" {
			for _, c := range cards {
				_, expires := extractCallerID(c.Reservation)
				if expires > now {
					// Someone else has this card reserved. We must not claim it.
					continue
				}
				// We found a card that's not reserved.
				if first || expires < oursExpires {
					if context.Debug {
						fmt.Printf("notefarm: %v looks unreserved because its expire time %v <= %v [now].\n", c.DeviceUID, expires, now)
					}
					// Let's plan on this being our card until we find a less recently used one.
					first = false
					ourCard = c
					oursExpires = expires
				}
			}
			if first {
				err = fmt.Errorf("notefarm: all cards are currently reserved")
				return
			}
			if context.Debug {
				fmt.Printf("notefarm: trying to reserve a new card %v.\n", ourCard.DeviceUID)
			}
		}

		// On an interim basis claim the card
		context.farmCard = ourCard

		// Reserve the card
		req := Request{Req: "card.reserve"}
		reservation := callerIDWithExpiration(now + int64(context.farmCheckoutMins*60))
		req.Status = reservation
		reqJSON, err1 := note.ObjectToJSON(req)
		if err1 != nil {
			err = err1
			return
		}
		if context.Debug {
			fmt.Printf("notefarm: sending reservation request\n")
		}
		_, err = remoteTransaction(context, portConfig, true, reqJSON)
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
					if c.Reservation == reservation {
						if context.Debug {
							fmt.Printf("notefarm: %s reserved for %d minutes\n", c.DeviceUID, context.farmCheckoutMins)
						}
						return
					}
				}
			}

			if context.Debug {
				fmt.Printf("notefarm: waiting for reservation confirmation of %s\n", ourCard.DeviceUID)
			}

		}

		// We no longer have the card
		context.farmCard = RemoteCard{}

	}
}

// Perform a remote transaction
func remoteTransaction(context *Context, portConfig int, noResponse bool, reqJSON []byte) (rspJSON []byte, err error) {
	// If our reservation has expired, fail the transaction
	if time.Now().Unix() > context.farmCheckoutExpires {
		err = fmt.Errorf("notefarm reservation of %d min has expired", context.farmCheckoutMins)
		return
	}

	// Perform the transaction several times to cover the case of exceeding the
	// transaction rate of the notefarm's proxy infrastructure
	var rspbuf []byte
	var resp *http.Response
	maxRetries := 5
	for i := 0; ; i++ {

		// Retry requests because Balena server needs to throttle us when we are hammering it
		success := false
		for i := 0; i < 10; i++ {

			// Do the HTTP request
			var req *http.Request
			req, err = http.NewRequest("POST", context.farmCard.DirectURL, bytes.NewBuffer(reqJSON))
			if err != nil {
				rspJSON = note.ErrorJSON("create request", err)
				break
			}
			httpclient := &http.Client{Timeout: time.Second * 90}
			resp, err = httpclient.Do(req)
			if err == nil {
				success = true
				break
			}

			// The standard web method for LB rate limit rejection is to reset the TCP circuit
			// Note that we need to detect EOF in this hacky way because
			// it is embedded at the end of a very long "Post:" message.
			if !strings.HasSuffix(fmt.Sprintf("%s", err), "EOF") {
				err = fmt.Errorf("http transmit after %d retries %s: %s", i+1, note.ErrCardIo, err)
				rspJSON = note.ErrorJSON("", err)
				if context.Debug {
					fmt.Printf("notefarm: send request err: %s\n", err)
				}
				break
			}

			if context.Debug {
				fmt.Printf("notefarm: send request SERVICE RATE LIMITING\n")
			}
			// Handle service rate-limiting by delaying for a moment, then retrying.  we
			// preset the response in case we exceed the maximum retries.
			time.Sleep(2 * time.Second)
			err = fmt.Errorf("rate limited after %d retries "+note.ErrCardIo, i+1)
			rspJSON = note.ErrorJSON("", err)

		}
		if !success {
			break
		}

		// Success, so now we read the response
		var buf bytes.Buffer
		_, err = io.Copy(&buf, resp.Body)
		resp.Body.Close()
		if err != nil {
			err = fmt.Errorf("reading response %s: %s", note.ErrCardIo, err)
			rspJSON = note.ErrorJSON("", err)
			break
		}
		rspbuf = buf.Bytes()

		// Validate that it's compliant JSON
		var jobj map[string]interface{}
		err = note.JSONUnmarshal(rspbuf, &jobj)
		if err == nil {

			// See if there was an I/O error to the card, and retry if so
			if !strings.Contains(string(rspbuf), note.ErrCardIo) {
				rspJSON = rspbuf
				break
			}
			if i > maxRetries {
				rsperr := fmt.Errorf("%s", string(rspbuf))
				rspJSON = note.ErrorJSON("proxy: cannot communicate with notecard", rsperr)
				break
			}
			if context.Debug {
				fmt.Printf("notefarm: send request read response I/O error\n")
			}

		} else {
			// Sometimes the response will not unmarshal because it's html from Balena
			// Balena error message embedded in html as shown below:
			// <p>UUID xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx</p><p>ERROR MESSAGE WE WANT</p>
			errstring := string(rspbuf)
			if strings.Contains(errstring, "<html>") && strings.Contains(errstring, "UUID") {
				if context.Debug {
					fmt.Printf("notefarm: send request read response IS HTML\n")
				}
				uuid := strings.Split(errstring, "UUID")
				// take everything after UUID and split by paragraph
				errmsg := strings.Split(uuid[1], "</p><p>")
				err = fmt.Errorf("notefarm controller error: %s", url.QueryEscape(errmsg[1]))
			} else {
				if context.Debug {
					fmt.Printf("notefarm: send request read response is not HTML but NOT JSON: %s\n", errstring)
				}
			}

			// Retry
			err = fmt.Errorf("hit max retries %s: %s", note.ErrCardIo, err)
			if i > maxRetries {
				rspJSON = note.ErrorJSON("", err)
				break
			}

		}

		// Sleep in case we're in the penalty box for too high of a transaction rate
		time.Sleep(2 * time.Second)

	}

	return
}
