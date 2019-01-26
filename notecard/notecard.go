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

// The module interface which is currently open
var portOpen string

// I2C

// CardI2CMax controls chunk size that's socially appropriate on the I2C bus.
// It must be 1-250 bytes as per spec
const CardI2CMax = 127

// Context for the serial package we're using
var openSerialPort *serial.Port         // tarm
//var openSerialPort io.ReadWriteCloser // jacobsa

// Create a card request
func Request(request string) (req CardRequest) {
    req.Request = request
    return
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

// NotecardPortEnum returns the list of all available ports on the specified interface
func PortEnum(interf string) (ports []string) {
    if interf == NotecardInterfaceSerial {
        ports = serialPortEnum()
    }
    if interf == NotecardInterfaceI2C {
        ports = i2cPortEnum()
    }
	return
}

// NotecardPortDefaults gets the defaults for the specified port
func PortDefaults(interf string) (port string, portConfig int) {
    if interf == NotecardInterfaceSerial {
        port, portConfig = serialDefault()
    }
    if interf == NotecardInterfaceI2C {
        port, portConfig = i2cDefault()
    }
    return
}

// Open the card to establish communications
func Open(moduleInterface string, port string, portConfig int) (err error) {

    // Open the interface
    switch moduleInterface {
    case NotecardInterfaceSerial:
        err = cardOpenSerial(port, portConfig)
        break
    case NotecardInterfaceI2C:
        err = cardOpenI2C(port, portConfig)
        break
    default:
        err = fmt.Errorf("unknown interface: %s", moduleInterface)
        break
    }
    if err != nil {
        portOpen = ""
        err = fmt.Errorf("error opening port: %s", err)
        return
    }

    // Success
    portOpen = moduleInterface
    return;

}

// Reset serial to a known state
func cardResetSerial() (err error) {

    // In order to ensure that we're not getting the reply to a failed
    // transaction from a prior session, drain any pending input prior
    // to transmitting a command.  Note that we use this technique of
    // looking for a known reply to \n\n, rather than just "draining
    // anything pending on serial", because the nature of read() is
    // that it blocks (until timeout) if there's nothing available.
    var length int
    buf := make([]byte, 2048)
    for {
        _, err = openSerialPort.Write([]byte("\n\n"))
        if err != nil {
            err = fmt.Errorf("error transmitting to module: %s", err)
            cardReportError(err)
            Close();
            return
        }
        time.Sleep(500*time.Millisecond)
        readBeganMs := int(time.Now().UnixNano() / 1000000)
        length, err = openSerialPort.Read(buf)
        readElapsedMs := int(time.Now().UnixNano() / 1000000) - readBeganMs
        if readElapsedMs == 0 && length == 0 && err == io.EOF {
            // On Linux, hardware port failures come back simply as immediate EOF
            err = fmt.Errorf("hardware failure")
        }
        if err != nil {
            err = fmt.Errorf("error reading from module: %s", err)
            cardReportError(err)
            Close();
            return
        }
        nonCRLFFound := false
        for i:=0; i<length && !nonCRLFFound; i++ {
            if (false) {
                fmt.Printf("chr: 0x%02x '%c'\n", buf[i], buf[i]);
            }
            if buf[i] != '\r' && buf[i] != '\n' {
                nonCRLFFound = true
            }
        }
        if !nonCRLFFound {
            break
        }
    }

    // Done
    return

}

// Open the card on serial
func cardOpenSerial(port string, portConfig int) (err error) {

    fmt.Printf("Using interface %s port %s at %d\n\n",
        NotecardInterfaceSerial, port, portConfig)

    // Open the serial port
    ///*    tarm
    c := &serial.Config{}
    c.Name = port
    c.Baud = portConfig
    c.ReadTimeout = time.Millisecond * 500
    openSerialPort, err = serial.OpenPort(c)
    //*/
    /* jacobsa
    c := serial.OpenOptions{}
    c.PortName = port
    c.BaudRate = portConfig
    c.MinimumReadSize = 0
    c.InterCharacterTimeout = 5000
    openSerialPort, err = serial.Open(c)
*/
    if err != nil {
        err = fmt.Errorf("error opening serial port %s at %d: %s", port, portConfig, err)
        return
    }

    // Reset serial to a known good state
    err = cardResetSerial()

    // All set
    return

}

// Reset I2C to a known good state
func cardResetI2C() (err error) {

    // For robustness (that is, in case the MCU was rebooted in the middle of receiving a reply, and thus
    // a partial buffer is waiting to be transmitted by the notecard), drain anything pending.
    for i:=0; i<5; i++ {
        rsp, err2 := cardTransactionI2C([]byte("\n"))
        if err2 == nil && string(rsp) == "\r\n" {
            break
        }
        err = err2
    }

    // Done
    return

}

// Open the card on I2C
func cardOpenI2C(port string, portConfig int) (err error) {

    fmt.Printf("Using interface %s\n\n", port)

    // Open the I2C port
    err = i2cOpen(uint8(portConfig), port)
    if err != nil {
        return fmt.Errorf("i2c init error: %s", err)
    }

    // Reset it to a known good state
    err = cardResetI2C()

    // Done
    return

}

// Close the port
func Close() {
    switch portOpen {
    case NotecardInterfaceSerial:
        cardCloseSerial()
        break
    case NotecardInterfaceI2C:
        cardCloseI2C()
        break
    default:
        break
    }
    portOpen = ""
}

// Close serial
func cardCloseSerial() {
    openSerialPort.Close()
}

// Close I2C
func cardCloseI2C() {
    i2cClose()
}

// Trace the incoming serial output
func Trace() (err error) {

    // Turn on tracing
    req := Request(ReqCardIO)
    req.Mode = "trace"
    req.Trace = "+usb"
    Transaction(req)

    // Spawn the input handler
    go inputHandler()

    // Loop, echoing to the console
    for {
        var length int
        buf := make([]byte, 2048)
        readBeganMs := int(time.Now().UnixNano() / 1000000)
        length, err = openSerialPort.Read(buf)
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
    Close();
    return

}

// Watch for console input
func inputHandler() {

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
                    openSerialPort.Write(ba)
                    // 'A' - 'Z'
                case 65 <= r && r <= 90:
                    ba := make([]byte, 1)
                    ba[0] = byte(r - 64)
                    openSerialPort.Write(ba)
                }
            }
        } else {
            openSerialPort.Write([]byte(message))
            openSerialPort.Write([]byte("\n"))
        }
    }

}

// Perform a card transaction
func Transaction(req CardRequest) (rsp CardRequest, err error) {

    // Marshal the request to JSON
    reqJSON, err2 := json.Marshal(req)
    if err2 != nil {
        err = fmt.Errorf("error marshaling request for module: %s", err2)
        return
    }

    // Perform the transaction
    rspJSON, err2 := TransactionJSON(reqJSON)
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

// Perform a card transaction
func TransactionJSON(reqJSON []byte) (rspJSON []byte, err error) {

    fmt.Printf("%s\n", reqJSON)

    // Make sure that the JSON has a terminator
    reqJSON = []byte(string(reqJSON) + "\n")

    // Perform the transaction
    switch portOpen {
    case NotecardInterfaceSerial:
        rspJSON, err = cardTransactionSerial(reqJSON)
        if err != nil {
            cardResetSerial()
        }
        break
    case NotecardInterfaceI2C:
        rspJSON, err = cardTransactionI2C(reqJSON)
        if err != nil {
            cardResetI2C()
        }
        break
    default:
        err = fmt.Errorf("unrecognized interface")
        return
    }
    if err != nil {
        return
    }

    fmt.Printf("%s\n", string(rspJSON))
    return

}

// Perform a card transaction over serial under the assumption that request already has '\n' terminator
func cardTransactionSerial(reqJSON []byte) (rspJSON []byte, err error) {

    // Transmit the request
    _, err = openSerialPort.Write(reqJSON)
    if err != nil {
        err = fmt.Errorf("error transmitting to module: %s", err)
        cardReportError(err)
        Close();
        return
    }

    // Read the reply until we get '\n' at the end
    for {
        var length int
        buf := make([]byte, 2048)
        readBeganMs := int(time.Now().UnixNano() / 1000000)
        length, err = openSerialPort.Read(buf)
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
            Close();
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
func cardTransactionI2C(reqJSON []byte) (rspJSON []byte, err error) {

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

        // If the length byte is 0 before we've received anything, it means that the module is
        // still processing the command and the reply isn't ready.  This is the NORMAL CASE
        // because it takes some number of milliseconds or seconds for the card to actually
        // process requests.
        chunklen := lenbuf[0]
        if chunklen == 0 && jsonbufLen == 0 {
            time.Sleep(100 * time.Millisecond)
            continue
        }

        // If the length byte is 0 after we've received something, it means that no more is coming.
        if chunklen == 0 {
            break
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
        rspJSON = append(rspJSON, readbuf...)
        jsonbufLen += readlen

    }

    // Done
    return
}
