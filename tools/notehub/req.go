// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/blues/note-go/notehub"
	"github.com/blues/note-go/noteutil"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Add an arg to an URL query string
func addQuery(in string, key string, value string) (out string) {
	out = in
	if value != "" {
		if out == "" {
			out += "?"
		} else {
			out += "&"
		}
		out += key
		out += "=\""
		out += value
		out += "\""
	}
	return
}

// Perform an HTTP requet, but do so using structs rather than bytes
func reqHub(hub string, request notehub.HubRequest, requestFile string, filetype string, filetags string, filenotes string, overwrite bool, secure bool, dropNonJSON bool, outq chan string) (response notehub.HubRequest, err error) {

	reqJSON, err2 := json.Marshal(request)
	if err2 != nil {
		err = err2
		return
	}

	rspJSON, err2 := reqHubJSON(hub, reqJSON, requestFile, filetype, filetags, filenotes, overwrite, secure, dropNonJSON, outq)
	if err2 != nil {
		err = err2
		return
	}

	err = json.Unmarshal(rspJSON, &response)
	return

}

// Perform an HTTP request
func reqHubJSON(hub string, request []byte, requestFile string, filetype string, filetags string, filenotes string, overwrite bool, secure bool, dropNonJSON bool, outq chan string) (response []byte, err error) {

	scheme := "http"
	if secure {
		scheme = "https"
	}

	fn := ""
	path := strings.Split(requestFile, "/")
	if len(path) > 0 {
		fn = path[len(path)-1]
	}

	httpurl := fmt.Sprintf("%s://%s%s", scheme, hub, notehub.DefaultAPITopicReq)
	query := addQuery("", "app", noteutil.Config.App)
	query = addQuery(query, "device", noteutil.Config.Device)
	query = addQuery(query, "upload", fn)
	if overwrite {
		query = addQuery(query, "overwrite", "true")
	}
	if filetype != "" {
		query = addQuery(query, "type", filetype)
	}
	if filetags != "" {
		query = addQuery(query, "tags", filetags)
	}
	if filenotes != "" {
		query = addQuery(query, "notes", url.PathEscape(filenotes))
	}
	httpurl += query

	var fileContents []byte
	var fileLength int
	buffer := bytes.NewBuffer(request)
	if requestFile != "" {
		fileContents, err = ioutil.ReadFile(requestFile)
		if err != nil {
			return
		}
		fileLength = len(fileContents)
		buffer = bytes.NewBuffer(fileContents)
	}

	if false {
		fmt.Printf("secure: %t\n%s\n", secure, httpurl)
	}
	httpReq, err := http.NewRequest("POST", httpurl, buffer)
	if err != nil {
		return
	}
	httpReq.Header.Set("User-Agent", "notehub-client")
	if requestFile != "" {
		httpReq.Header.Set("Content-Length", fmt.Sprintf("%d", fileLength))
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	httpClient := &http.Client{}

	if secure {

		if noteutil.Config.Cert == "" || noteutil.Config.Key == "" {
			err = fmt.Errorf("HTTPS client cert (-cert) and key (-key) are required for secure API access")
			return
		}

		clientCert, err2 := tls.LoadX509KeyPair(noteutil.Config.Cert, noteutil.Config.Key)
		if err2 != nil {
			err = err2
			return
		}

		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{clientCert},
			InsecureSkipVerify: true,
		}

		if noteutil.Config.Root != "" {

			caCert, err2 := ioutil.ReadFile(noteutil.Config.Root)
			if err2 != nil {
				err = err2
				return
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			tlsConfig = &tls.Config{
				Certificates: []tls.Certificate{clientCert},
				RootCAs:      caCertPool,
			}

		}

		tlsConfig.BuildNameToCertificate()

		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
		}

		httpClient = &http.Client{
			Transport: transport,
		}

	}

	httpRsp, err2 := httpClient.Do(httpReq)
	if err2 != nil {
		err = err2
		return
	}

	// Note that we must do this with no timeout specified on
	// the httpClient, else monitor mode would time out.
	b := make([]byte, 2048)
	linebuf := []byte{}
	for {
		n, err2 := httpRsp.Body.Read(b)
		if n > 0 {

			// Append to result buffer if no outq is specified
			if outq == nil {

				response = append(response, b[:n]...)

			} else {

				// Enqueue lines for monitoring
				linebuf = append(linebuf, b[:n]...)
				for {

					// Parse out a full line and queue it, saving the leftover
					i := bytes.IndexRune(linebuf, '\n')
					if i == -1 {
						break
					}
					line := linebuf[0 : i+1]
					linebuf = linebuf[i+1:]
					if !dropNonJSON {
						outq <- string(line)
					} else {
						if strings.HasPrefix(string(line), "{") {
							outq <- string(line)
						}
					}

					// Remember the very last line as the response, in case it
					// was an error and we're about to get an io.EOF
					response = line
				}

			}

		}
		if err2 != nil {
			if err2 != io.EOF {
				err = err2
				return
			}
			break
		}
	}

	return

}
