// Copyright 2023 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

// Adapted from https://github.com/mrVanboy/go-simple-cobs (MIT)

package notecard

import "fmt"

// Decode with optional XOR
func CobsDecode(input []byte, xor byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("cobs: no input")
	}
	var output = make([]byte, 0)
	var lastZI = 0
	for {
		if lastZI == len(input) {
			break
		}
		if (input[lastZI] ^ xor) == 0x00 {
			return nil, fmt.Errorf("cobs: zero found in input array")
		}
		if int((input[lastZI] ^ xor)) > (len(input) - int(lastZI)) {
			return nil, fmt.Errorf("cobs: out of bounds")
		}
		var nextZI = lastZI + int(input[lastZI]^xor)
		for i := lastZI + 1; i < nextZI; i++ {
			if (input[i] ^ xor) == 0x00 {
				return nil, fmt.Errorf("cobs: zero not allowed in input")
			}
			output = append(output, (input[i] ^ xor))
		}
		if nextZI < len(input) && (input[lastZI]^xor) != 0xFF {
			output = append(output, 0x00)
		}
		lastZI = nextZI
	}
	return output, nil
}

// Encode with optional XOR
func CobsEncode(input []byte, xor byte) ([]byte, error) {
	var out = []byte{0x01 ^ xor}
	var lastZI = 0
	var delta byte = 1
	for i := range input {
		if input[i] == 0x00 {
			out[lastZI] = delta ^ xor
			out = append(out, 0x01^xor)
			lastZI = len(out) - 1
			delta = 1
		} else {
			if delta == 255 {
				out[lastZI] = delta ^ xor
				out = append(out, 0x01^xor)
				lastZI = len(out) - 1
				delta = 1
			}
			out = append(out, input[i]^xor)
			delta++
		}
	}
	out[lastZI] = delta ^ xor
	return out, nil
}
