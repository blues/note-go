// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/blues/note-go/note"
)

// Leaseing service parameters
const leaseTransactionService = "https://notepod.io:8123"
const leaseTraceService = "proxy.notepod.io:123"

// Lease transaction
type LeaseTransaction struct {
	Request    string `json:"req,omitempty"`
	Lessor     string `json:"lessor,omitempty"`
	Scope      string `json:"scope,omitempty"`
	Expires    int64  `json:"expires,omitempty"`
	Error      string `json:"err,omitempty"`
	DeviceUID  string `json:"device,omitempty"`
	NoResponse bool   `json:"no_response,omitempty"`
	ReqJSON    string `json:"request_json,omitempty"`
	RspJSON    string `json:"response_json,omitempty"`
}

// Request types
const (
	ReqReserve     = "reserve"
	ReqTransaction = "transaction"
)

// Perform an HTTP transaction to the lease service
func leaseService(req LeaseTransaction, promoteError bool) (rsp LeaseTransaction, err error) {

	reqj, err := json.Marshal(req)
	if err != nil {
		return rsp, err
	}

	// Send the transaction
	hreq, err := http.NewRequest("POST", leaseTransactionService, bytes.NewBuffer(reqj))
	if err != nil {
		return rsp, fmt.Errorf("%s %s", err, note.ErrCardIo)
	}
	hcli := &http.Client{Timeout: time.Second * 90}
	hrsp, err := hcli.Do(hreq)
	if err != nil {
		return rsp, fmt.Errorf("%s %s", err, note.ErrCardIo)
	}
	defer hrsp.Body.Close()

	// Read the response
	var rspjb bytes.Buffer
	_, err = io.Copy(&rspjb, hrsp.Body)
	if err != nil {
		return rsp, fmt.Errorf("%s %s", err, note.ErrCardIo)
	}
	rspj := rspjb.Bytes()

	err = note.JSONUnmarshal(rspj, &rsp)
	if err != nil {
		return rsp, fmt.Errorf("%s %s", err, note.ErrCardIo)
	}

	if promoteError && rsp.Error != "" {
		return rsp, fmt.Errorf("%s", rsp.Error)
	}

	return rsp, nil

}

// Open or reopen the remote card by taking out a lease, or by renewing the lease.
func leaseReopen(context *Context, portConfig int) (err error) {

	context.portIsOpen = false

	// Don't reopen if tracing
	if InitialTraceMode {
		context.portIsOpen = true
		return
	}

	// Find out our unique ID
	context.leaseLessor = callerID()

	// Perform the lease transaction
	req := LeaseTransaction{}
	req.Request = ReqReserve
	req.Lessor = context.leaseLessor
	req.Scope = context.leaseScope
	req.Expires = context.leaseExpires
	rsp, err := leaseService(req, true)
	if err != nil {
		return err
	}

	// Trace so that we can find out when
	if context.leaseExpires == 0 {
		fmt.Printf("%s reserved until %s\n", rsp.DeviceUID, time.Unix(rsp.Expires, 0).Local().Format("03:04:05 PM MST"))
	}

	// Save the deviceUID to the allocated device
	context.leaseScope = rsp.Scope
	context.leaseExpires = rsp.Expires
	context.leaseDeviceUID = rsp.DeviceUID

	// Open
	context.portIsOpen = true

	return
}

// Close a remote notecard
func leaseClose(context *Context) {
	context.portIsOpen = false
}

// Perform a remote transaction
func leaseTransaction(context *Context, portConfig int, noResponse bool, reqJSON []byte, delay bool) (rspJSON []byte, err error) {

	// Perform the lease transaction
	req := LeaseTransaction{}
	req.Request = ReqTransaction
	req.Lessor = context.leaseLessor
	req.DeviceUID = context.leaseDeviceUID
	req.ReqJSON = string(reqJSON)
	req.NoResponse = noResponse
	rsp, err := leaseService(req, true)
	if err != nil {
		return rspJSON, err
	}

	// Done
	return []byte(rsp.RspJSON), nil

}

// Lease trace open
func leaseTraceOpen(context *Context) (err error) {

	// Scope must be a specific device UID for trace
	if !strings.HasPrefix(context.port, "dev:") {
		return fmt.Errorf("trace is only allowed when a deviceUID is specified")
	}

	// Open the service connection
	tcpServer, err := net.ResolveTCPAddr("tcp", leaseTraceService)
	if err != nil {
		return
	}
	context.leaseTraceConn, err = net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		return
	}

	// Write an initial non-json line containing scope, to signal to the service that this is a trace connection
	leaseTraceWrite(context, []byte(context.port+"\n"))

	// Done
	return

}

// Lease trace read function
func leaseTraceRead(context *Context) (data []byte, err error) {

	buf := make([]byte, 2048)
	length, err := context.leaseTraceConn.Read(buf)
	if err != nil {
		if err == io.EOF {
			// Just a read timeout
			return data, nil
		}
		return data, err
	}

	return buf[:length], nil

}

// Lease trace write function
func leaseTraceWrite(context *Context, data []byte) {
	context.leaseTraceConn.Write(data)
}
