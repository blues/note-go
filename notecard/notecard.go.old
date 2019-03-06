// Copyright 2017 Inca Roads LLC.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
    "os"
    "bufio"
    "strings"
    "io"
    "fmt"
    "time"
    "encoding/json"
    "github.com/tarm/serial"
    //  "github.com/jacobsa/go-serial/serial"
)

// Module communication interfaces
const (
	NotecardInterfaceSerial = "serial"
	NotecardInterfaceI2C = "i2c"
)

// CardI2CMax controls chunk size that's socially appropriate on the I2C bus.
// It must be 1-250 bytes as per spec
const CardI2CMax = 127

// Context for the serial package we're using

// Context for the port that is open
type Context struct {

	// Class functions
	PortEnumFn func () (ports []string)
	PortDefaultsFn func () (port string, portConfig int)
	CloseFn func (context *Context) ()
	ResetFn func (context *Context) (err error)
	TransactionFn func (context *Context, reqJSON []byte) (rspJSON []byte, err error)

	// Serial instance state
	isSerial bool
	openSerialPort *serial.Port			// tarm
	//openSerialPort io.ReadWriteCloser // jacobsa

}

// Report a critical card error
func cardReportError(err error) {
    fmt.Printf("***\n");
    fmt.Printf("***\n");
    fmt.Printf("*** %s\n", err);
    fmt.Printf("***\n");
    fmt.Printf("***\n");
    time.Sleep(10 * time.Second)
}

// EnumPorts returns the list of all available ports on the specified interface
func (context *Context) EnumPorts() (ports []string) {
	return context.PortEnumFn()
}

// PortDefaults gets the defaults for the specified port
func (context *Context) PortDefaults() (port string, portConfig int) {
	return context.PortDefaultsFn()
}

// Open the card to establish communications
func Open(moduleInterface string, port string, portConfig int) (context Context, err error) {

    switch moduleInterface {
    case NotecardInterfaceSerial:
        context, err = OpenSerial(port, portConfig)
        break
    case NotecardInterfaceI2C:
        context, err = OpenI2C(port, portConfig)
        break
    default:
        err = fmt.Errorf("unknown interface: %s", moduleInterface)
        break
    }
    if err != nil {
        err = fmt.Errorf("error opening port: %s", err)
        return
    }

    return;

}

// Reset serial to a known state
func cardResetSerial(context *Context) (err error) {

    // In order to ensure that we're not getting the reply to a failed
    // transaction from a prior session, drain any pending input prior
    // to transmitting a command.  Note that we use this technique of
    // looking for a known reply to \n\n, rather than just "draining
    // anything pending on serial", because the nature of read() is
    // that it blocks (until timeout) if there's nothing available.
    var length int
    buf := make([]byte, 2048)
    for {
        _, err = context.openSerialPort.Write([]byte("\n\n"))
        if err != nil {
            err = fmt.Errorf("error transmitting to module: %s", err)
            cardReportError(err)
            return
        }
        time.Sleep(500*time.Millisecond)
        readBeganMs := int(time.Now().UnixNano() / 1000000)
        length, err = context.openSerialPort.Read(buf)
        readElapsedMs := int(time.Now().UnixNano() / 1000000) - readBeganMs
        if readElapsedMs == 0 && length == 0 && err == io.EOF {
            // On Linux, hardware port failures come back simply as immediate EOF
            err = fmt.Errorf("hardware failure")
        }
        if err != nil {
            err = fmt.Errorf("error reading from module: %s", err)
            cardReportError(err)
            return
        }
        somethingFound := false
        nonCRLFFound := false
        for i:=0; i<length && !nonCRLFFound; i++ {
            if (false) {
                fmt.Printf("chr: 0x%02x '%c'\n", buf[i], buf[i]);
            }
	        somethingFound = true
            if buf[i] != '\r' && buf[i] != '\n' {
                nonCRLFFound = true
            }
        }
        if somethingFound && !nonCRLFFound {
            break
        }
    }

    // Done
    return

}

// OpenSerial opens the card on serial
func OpenSerial(port string, portConfig int) (context Context, err error) {

    fmt.Printf("Using interface %s port %s at %d\n\n", NotecardInterfaceSerial, port, portConfig)

	// Set up class functions
	context.PortEnumFn = serialPortEnum
	context.PortDefaultsFn = serialDefault
	context.CloseFn = cardCloseSerial
	context.ResetFn = cardResetSerial
	context.TransactionFn = cardTransactionSerial
	context.isSerial = true

    // Open the serial port
    ///*    tarm
    c := &serial.Config{}
    c.Name = port
    c.Baud = portConfig
    c.ReadTimeout = time.Millisecond * 500
    context.openSerialPort, err = serial.OpenPort(c)
    //*/
    /* jacobsa
    c := serial.OpenOptions{}
    c.PortName = port
    c.BaudRate = portConfig
    c.MinimumReadSize = 0
    c.InterCharacterTimeout = 5000
    context.openSerialPort, err = serial.Open(c)
*/
    if err != nil {
        err = fmt.Errorf("error opening serial port %s at %d: %s", port, portConfig, err)
        return
    }

    // Reset serial to a known good state
    err = cardResetSerial(&context)

    // All set
    return

}

// Reset I2C to a known good state
func cardResetI2C(context *Context) (err error) {

    // For robustness (that is, in case the MCU was rebooted in the middle of receiving a reply, and thus
    // a partial buffer is waiting to be transmitted by the notecard), drain anything pending.
    for i:=0; i<5; i++ {
        rsp, err2 := cardTransactionI2C(context, []byte("\n"))
        if err2 == nil && string(rsp) == "\r\n" {
            break
        }
        err = err2
    }

    // Done
    return

}

// OpenI2C opens the card on I2C
func OpenI2C(port string, portConfig int) (context Context, err error) {

    fmt.Printf("Using interface %s\n\n", port)

	// Set up class functions
	context.PortEnumFn = i2cPortEnum
	context.PortDefaultsFn = i2cDefault
	context.CloseFn = cardCloseI2C
	context.ResetFn = cardResetI2C
	context.TransactionFn = cardTransactionI2C

    // Open the I2C port
    err = i2cOpen(uint8(portConfig), port)
    if err != nil {
		err = fmt.Errorf("i2c init error: %s", err)
        return
    }

    // Reset it to a known good state
    err = cardResetI2C(&context)

    // Done
    return

}

// Reset the port
func (context *Context) Reset() (err error) {
	return context.ResetFn(context)
}

// Close the port
func (context *Context) Close() {
	context.CloseFn(context)
}

// Close serial
func cardCloseSerial(context *Context) {
    context.openSerialPort.Close()
}

// Close I2C
func cardCloseI2C(context *Context) {
    i2cClose()
}

// SerialDefaults returns the default serial parameters
func SerialDefaults () (port string, portConfig int) {
	return serialDefault()
}

// I2CDefaults returns the default serial parameters
func I2CDefaults () (port string, portConfig int) {
	return i2cDefault()
}

// SerialPorts returns the list of available serial ports
func SerialPorts () (ports []string) {
	return serialPortEnum()
}

// I2CPorts returns the list of available I2C ports
func I2CPorts () (ports []string) {
	return i2cPortEnum()
}

// Trace the incoming serial output
func (context *Context) Trace() (err error) {

	// Tracing only works for USB and AUX ports
	if (!context.isSerial) {
		return fmt.Errorf("tracing is only available on USB and AUX ports")
	}

    // Turn on tracing
    req := Request{Req:ReqCardIO}
    req.Mode = "trace"
    req.Trace = "+usb"
    context.Transaction(req)

    // Spawn the input handler
    go inputHandler(context)

    // Loop, echoing to the console
    for {
        var length int
        buf := make([]byte, 2048)
        readBeganMs := int(time.Now().UnixNano() / 1000000)
        length, err = context.openSerialPort.Read(buf)
        readElapsedMs := int(time.Now().UnixNano() / 1000000) - readBeganMs
        if (false) {
            fmt.Printf("mon: elapsed:%d len:%d err:%s '%s'\n", readElapsedMs, length, err, string(buf[:length]));
        }
        if readElapsedMs == 0 && length == 0 && err == io.EOF {
            // On Linux, hardware port failures come back simply as immediate EOF
            err = fmt.Errorf("hardware failure")
        }
        if err != nil {
            if (err == io.EOF) {
                // Just a read timeout
                continue
            }
            break
        }
        fmt.Printf("%s", buf[:length])
    }

    err = fmt.Errorf("error reading from module: %s", err)
    cardReportError(err)
    return

}

// Watch for console input
func inputHandler(context *Context) {

    // Create a scanner to watch stdin
    scanner := bufio.NewScanner(os.Stdin)
    var message string

    for {

        scanner.Scan()
        message = scanner.Text()

        if (strings.HasPrefix(message, "^")) {

            for _, r := range message[1:] {
                switch {
                    // 'a' - 'z'
                case 97 <= r && r <= 122:
                    ba := make([]byte, 1)
                    ba[0] = byte(r - 96)
                    context.openSerialPort.Write(ba)
                    // 'A' - 'Z'
                case 65 <= r && r <= 90:
                    ba := make([]byte, 1)
                    ba[0] = byte(r - 64)
                    context.openSerialPort.Write(ba)
                }
            }
        } else {
            context.openSerialPort.Write([]byte(message))
            context.openSerialPort.Write([]byte("\n"))
        }
    }

}

// Transaction performs a card transaction
func (context *Context) Transaction(req Request) (rsp Request, err error) {

    // Marshal the request to JSON
    reqJSON, err2 := json.Marshal(req)
    if err2 != nil {
        err = fmt.Errorf("error marshaling request for module: %s", err2)
        return
    }

    // Perform the transaction
    rspJSON, err2 := context.TransactionJSON(reqJSON)
    if err2 != nil {
        err = fmt.Errorf("error marshaling request for module: %s", err2)
        return
    }

    // Unmarshal for convenience of the caller
    err = json.Unmarshal(rspJSON, &rsp)
    if err != nil {
        err = fmt.Errorf("error unmarshaling reply from module: %s", err)
        return
    }

    // Done
    return
}

// TransactionJSON performs a card transaction using raw JSON []bytes
func (context *Context) TransactionJSON(reqJSON []byte) (rspJSON []byte, err error) {

    fmt.Printf("%s\n", reqJSON)

    // Make sure that the JSON has a terminator
    reqJSON = []byte(string(reqJSON) + "\n")

    // Perform the transaction
    rspJSON, err = context.TransactionFn(context, reqJSON)
    if err != nil {
        context.ResetFn(context)
    }

    fmt.Printf("%s\n", string(rspJSON))
    return

}

// Perform a card transaction over serial under the assumption that request already has '\n' terminator
func cardTransactionSerial(context *Context, reqJSON []byte) (rspJSON []byte, err error) {

    // Transmit the request
    _, err = context.openSerialPort.Write(reqJSON)
    if err != nil {
        err = fmt.Errorf("error transmitting to module: %s", err)
        cardReportError(err)
        return
    }

    // Read the reply until we get '\n' at the end
    for {
        var length int
        buf := make([]byte, 2048)
        readBeganMs := int(time.Now().UnixNano() / 1000000)
        length, err = context.openSerialPort.Read(buf)
        readElapsedMs := int(time.Now().UnixNano() / 1000000) - readBeganMs
        if (false) {
            err2 := err
            if err2 == nil {
                err2 = fmt.Errorf("none")
            }
            fmt.Printf("req: elapsed:%d len:%d err:%s '%s'\n", readElapsedMs, length, err2, string(buf[:length]));
        }
        if readElapsedMs == 0 && length == 0 && err == io.EOF {
            // On Linux, hardware port failures come back simply as immediate EOF
            err = fmt.Errorf("hardware failure")
        }
        if err != nil {
            if err == io.EOF {
                // Just a read timeout
                continue
            }
            err = fmt.Errorf("error reading from module: %s", err)
            cardReportError(err)
            return
        }
        rspJSON = append(rspJSON, buf[:length]...)
        if strings.HasSuffix(string(rspJSON), "\n") {
            break
        }
    }

    // Done
    return

}

// Perform a card transaction over I2C under the assumption that request already has '\n' terminator
func cardTransactionI2C(context *Context, reqJSON []byte) (rspJSON []byte, err error) {

    var writelen, lenlen, readlen int
    lenbuf := make([]byte, 1)

    // Send the transaction on the bus
    chunkoffset := 0
    jsonbufLen := len(reqJSON)
    for jsonbufLen > 0 {
        chunklen := CardI2CMax
        if jsonbufLen < chunklen {
            chunklen = jsonbufLen
        }
        writelen, err = i2cWriteByte(byte(chunklen))
        if err != nil {
            err = fmt.Errorf("chunklen write error: %s", err)
            return
        }
        if writelen != 1 {
            err = fmt.Errorf("write of 1-byte chunklen actually wrote %d bytes", writelen)
            return
        }
        writelen, err = i2cWriteBytes(reqJSON[chunkoffset:chunkoffset+chunklen])
        if err != nil {
            err = fmt.Errorf("chunk write error: %s", err)
            return
        }
        if writelen != chunklen {
            err = fmt.Errorf("write of %d-byte chunk actually wrote %d bytes", chunklen, writelen)
            return
        }
        chunkoffset += chunklen;
        jsonbufLen -= chunklen;
    }

    // Loop, building a reply buffer out of received chunks.  We'll build the reply in the same
    // buffer we used to transmit, and will grow it as necessary.
    jsonbufLen = 0;
	receivedNewline := false;
    for {

        // Issue an "get the next chunk len to read" request
        writelen, err = i2cWriteByte(0x80)
        if err != nil {
            err = fmt.Errorf("error writing get-next-chunklen command: %s", err)
            return
        }
        if writelen != 1 {
            err = fmt.Errorf("write of get-next-chunklen command actually wrote %d bytes", writelen)
            return
        }

        // Get the length byte, which always precedes anything else
        lenlen, err = i2cReadBytes(lenbuf)
        if err != nil {
            err = fmt.Errorf("error reading response length: %s", err)
            return
        }
        if lenlen != 1 {
            err = fmt.Errorf("read of 1-byte chunklen command actually wrote %d bytes", lenlen)
            return
        }

        // If the length byte is 0 before we've received a \n, it means that the module is
        // still processing the command and the reply isn't ready.  This is the NORMAL CASE
        // because it takes some number of milliseconds or seconds for the card to actually
        // process requests, and there may be an occasional delay between chunks just because
		// of normal task preemption.
        chunklen := lenbuf[0]
        if chunklen == 0 {
	        if receivedNewline {
	            break
	        }
            time.Sleep(100 * time.Millisecond)
            continue
        }

        // Issue a "read the next chunk" request
        if chunklen > CardI2CMax {
            chunklen = CardI2CMax
        }
        writelen, err = i2cWriteByte(0x80+chunklen)
        if err != nil {
            err = fmt.Errorf("error writing read-next-chunk command: %s", err)
            return
        }
        if writelen != 1 {
            err = fmt.Errorf("write of read-next-chunk command actually wrote %d bytes", writelen)
            return
        }

        // Read the next chunk
        readbuf := make([]byte, chunklen)
        readlen, err = i2cReadBytes(readbuf)
        if err != nil {
            err = fmt.Errorf("error reading chunk: %s", err)
            return
        }
        if readlen != int(chunklen) {
            err = fmt.Errorf("read of %d-byte chunk actually read %d bytes", chunklen, readlen)
            return
        }

		// If the last byte of the chunk is \n, chances are that we're done.  However, just so
		// that we pull everything pending from the module, we only exit when we've received
		// a newline AND there's nothing left available from the module.
		if readbuf[readlen-1] == '\n' {
			receivedNewline = true
		}

		// Append to the JSON being accumulated
        rspJSON = append(rspJSON, readbuf...)
        jsonbufLen += readlen

    }

    // Done
    return
}
