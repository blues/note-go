// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blues/note-go/note"
	"github.com/blues/note-go/notecard"
	"github.com/blues/note-go/notehub"
)

// NotecardFirmwareSignature is used to identify whether or not this firmware is a
// candidate for downloading onto notecards.  Note that this is not a security feature; if someone
// embeds this binary sequence and embeds it, they will be able to do precisely what they can do
// by using the USB to directly download firmware onto the device. This mechanism is intended for
// convenience and is just intended to keep people from inadvertently hurting themselves.
var NotecardFirmwareSignature = []byte{0x82, 0x1c, 0x6e, 0xb7, 0x18, 0xec, 0x4e, 0x6f, 0xb3, 0x9e, 0xc1, 0xe9, 0x8f, 0x22, 0xe9, 0xf6}

// Side-loads a file to the DFU area of the notecard, to avoid download
func dfuLoad(filename string, slow bool) (err error) {
	var req notecard.Request

	// Side-loading generally is performed with a USB-connected Notecard
	// that is idle, as opposed to being connected with long header wires
	// that have significant capacitance and resistance, and where there
	// may be arbitrary activity on the Notecard.  This is written to anticipate
	// I/O failure due to Notecard overrun (note.ErrCardIo) so that we can send
	// long requests much faster.
	if !slow {
		notecard.RequestSegmentMaxLen = 1024
		notecard.RequestSegmentDelayMs = 30
	}

	// Read the file up-front so we can handle this common failure
	// before we go into dfu mode
	var bin []byte
	bin, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	// Set the current mode to DFU mode
	fmt.Printf("placing notecard into DFU mode so that we can send file to its external flash storage\n")
	req = notecard.Request{Req: "hub.set"}
	req.Mode = "dfu"
	_, err = card.TransactionRequest(req)
	if err != nil {
		return
	}

	// Wait until dfu status says that we're in DFU mode
	for {
		fmt.Printf("waiting for notecard to power-up the external flash storage\n")
		_, err = card.TransactionRequest(notecard.Request{Req: "dfu.put"})
		if err != nil && !note.ErrorContains(err, note.ErrDFUNotReady) && !note.ErrorContains(err, note.ErrCardIo) {
			restoreMode()
			return
		}
		if err == nil {
			break
		}
		time.Sleep(1500 * time.Millisecond)
	}

	// Do the write
	fmt.Printf("sending DFU binary to notecard\n")
	err = loadBin(filename, bin)

	// Restore the DFU state
	fmt.Printf("restoring notecard so that it is no longer in DFU mode\n")
	restoreMode()

	// Done
	fmt.Printf("sideload completed\n")
	return
}

// Side-load a single bin
func loadBin(filename string, bin []byte) (err error) {
	var req, rsp notecard.Request
	totalLen := len(bin)

	// Generate the simulated firmware info
	var dbu notehub.HubRequestFile
	dbu.Created = time.Now().Unix()
	dbu.Source = filename
	dbu.MD5 = fmt.Sprintf("%x", md5.Sum(bin))
	dbu.CRC32 = crc32.ChecksumIEEE(bin)
	dbu.Length = totalLen
	dbu.Name = filename
	dbu.FileType = notehub.HubFileTypeUserFirmware
	if bytes.Contains(bin, NotecardFirmwareSignature) {
		dbu.FileType = notehub.HubFileTypeCardFirmware
	}
	var body map[string]interface{}
	body, err = note.ObjectToBody(dbu)
	if err != nil {
		return
	}

	// Issue the first request, which is to initiate the DFU put
	req = notecard.Request{Req: "dfu.put"}
	req.Body = &body
	rsp, err = card.TransactionRequest(req)
	if err != nil {
		return
	}
	chunkLen := int(rsp.Length)

	// Send the chunk to sideload
	offset := 0
	lenRemaining := totalLen
	for lenRemaining > 0 {

		// Determine how much to send
		thisLen := lenRemaining
		if thisLen > chunkLen {
			thisLen = chunkLen
		}

		// Send the chunk
		fmt.Printf("side-loading %d bytes (%d remaining)\n", thisLen, lenRemaining-thisLen)
		req = notecard.Request{Req: "dfu.put"}
		payload := bin[offset : offset+thisLen]
		req.Payload = &payload
		req.Offset = int32(offset)
		req.Length = int32(thisLen)
		rsp, err = card.TransactionRequest(req)
		if err != nil {
			if note.ErrorContains(err, note.ErrCardIo) {
				fmt.Printf("retrying after side-loading error: %s\n", err)
				continue
			}
			fmt.Printf("aborting after side-loading error: %s\n", err)
			return
		}

		// Move on to next chunk
		lenRemaining -= thisLen
		offset += thisLen

		// Wait until the migration succeeds
		for rsp.Pending {
			rsp, err = card.TransactionRequest(notecard.Request{Req: "dfu.put"})
			if err != nil {
				if note.ErrorContains(err, note.ErrDFUNotReady) && lenRemaining == 0 {
					err = nil
					break
				}
				fmt.Printf("aborting after error retrieving side-loading status: %s\n", err)
				return
			}
			time.Sleep(750 * time.Millisecond)
		}

	}

	// Done
	return

}

// Put hub.set mode back to what it had been
func restoreMode() {
	req := notecard.Request{Req: "hub.set"}
	req.Mode = "dfu-completed"
	card.TransactionRequest(req)
}

// Collects multiple .bin files into a single multi-bin file for composite sideloads/downloads
func dfuPackage(outfile string, hostProcessorType string, args []string) (err error) {

	// Preset error
	badFmtErr := fmt.Errorf("MCU type must be followed addr:bin list such as '0x0:bootloader.bin 0x10000:user.bin'")

	// Parse args
	if len(args) == 0 {
		return badFmtErr
	}

	addresses := []int{}
	regions := []int{}
	filenames := []string{}
	files := [][]byte{}
	for _, pair := range args {

		pairSplit := strings.Split(pair, ":")
		if len(pairSplit) < 2 || pairSplit[0] == "" || pairSplit[1] == "" {
			return badFmtErr
		}

		fn := pairSplit[1]
		filenames = append(filenames, filepath.Base(fn))

		if strings.HasPrefix(fn, "~/") {
			usr, _ := user.Current()
			fn = filepath.Join(usr.HomeDir, fn[2:])
		}
		bin, err := ioutil.ReadFile(fn)
		if err != nil {
			return fmt.Errorf("%s: %s", fn, err)
		}
		files = append(files, bin)

		var num int
		numstr := pairSplit[0]
		numsplit := strings.Split(numstr, ",")
		if len(numsplit) == 1 {
			num, err = parseNumber(numstr)
			if err != nil {
				return err
			}
			addresses = append(addresses, num)
			regions = append(regions, len(bin))
		} else {
			num, err = parseNumber(numsplit[0])
			if err != nil {
				return err
			}
			addresses = append(addresses, num)
			num, err = parseNumber(numsplit[1])
			if err != nil {
				return err
			}
			regions = append(regions, num)
		}

	}

	// Build the prefix string
	now := time.Now().UTC()
	prefix := "/// BINPACK ///\n"
	prefix += "WHEN: " + now.Format("2006-01-02 15:04:05 UTC") + "\n"
	prefix += "HOST: " + hostProcessorType + "\n"
	for i := range addresses {
		cleanFn := strings.ReplaceAll(filenames[i], ",", "")
		prefix += fmt.Sprintf("LOAD: %s,%d,%d,%d,%x\n", cleanFn, addresses[i], regions[i], len(files[i]), md5.Sum(files[i]))
	}
	prefix += "/// BINPACK ///\n"

	// Create the output file
	ext := ".binpack"
	if strings.HasPrefix(outfile, "~/") {
		usr, _ := user.Current()
		outfile = filepath.Join(usr.HomeDir, outfile[2:])
	}
	if outfile == "" {
		outfile = now.Format("2006-01-02-150405") + ext
	} else if strings.HasSuffix(outfile, "/") {
		outfile += now.Format("2006-01-02-150405") + ext
	} else if !strings.HasSuffix(outfile, ext) {
		tmp := strings.Split(outfile, ".")
		if len(tmp) > 1 {
			outfile = strings.Join(tmp[:len(tmp)-1], ".")
		}
		outfile += ext
	}

	os.Remove(outfile)
	fd, err := os.Create(outfile)
	if err != nil {
		return err
	}

	// Write the prefix, with its terminators
	fd.Write([]byte(prefix))
	fd.Write([]byte{0})

	// Concatenate the binaries
	for i := range files {
		fd.Write(files[i])
	}

	// Don't need file anymore
	fd.Close()

	// Get stats
	fi, err := os.Stat(outfile)
	if err != nil {
		return err
	}

	// Done
	fmt.Printf("%s now incorporates %d files and is %d bytes:\n\n%s\n", outfile, len(addresses), fi.Size(), prefix)
	return nil

}

// Parse a number, allowing for hex or decimal
func parseNumber(numstr string) (num int, err error) {
	var num64 int64
	if strings.HasPrefix(numstr, "0x") || strings.HasPrefix(numstr, "0X") {
		numstr = strings.TrimPrefix(strings.TrimPrefix(numstr, "0x"), "0X")
		num64, err = strconv.ParseInt(numstr, 16, 64)
		if err != nil {
			return 0, err
		}
		return int(num64), nil
	}
	num64, err = strconv.ParseInt(numstr, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(num64), nil
}
