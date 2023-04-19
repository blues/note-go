// Copyright 2023 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

// Decode with optional XOR
func CobsDecode(inputOutput []byte, xor byte) ([]byte, error) {
	length := len(inputOutput)
	inOffset := 0
	outOffset := inOffset
	startOffset, endOffset := outOffset, inOffset+length
	var code, copy uint8 = 0xFF, 0
	for inOffset < endOffset {
		if copy != 0 {
			inputOutput[outOffset] = inputOutput[inOffset] ^ xor
			outOffset, inOffset, copy = outOffset+1, inOffset+1, copy-1
		} else {
			if code != 0xFF {
				inputOutput[outOffset] = 0
				outOffset = outOffset + 1
			}
			code = inputOutput[inOffset] ^ xor
			copy, inOffset = code, inOffset+1
			if code == 0 {
				break
			}
		}
	}
	return inputOutput[startOffset:outOffset], nil
}

// Get the maximum size of the cobs-encoded buffer
func CobsEncodedLength(length int) int {
	return length + (1 + (length / 254))
}

// Encode with optional XOR
func CobsEncode(input []byte, xor byte) ([]byte, error) {
	length := len(input)
	inOffset := 0
	output := make([]byte, CobsEncodedLength(len(input)))
	outOffset := 0
	outStartOffset := outOffset
	var ch, code uint8
	code = 1
	outCodeOffset := outOffset
	outOffset = outOffset + 1
	for length > 0 {
		ch = input[inOffset]
		inOffset = inOffset + 1
		length = length - 1
		if ch != 0 {
			output[outOffset] = ch ^ xor
			outOffset = outOffset + 1
			code = code + 1
		}
		if ch == 0 || code == 0xFF {
			output[outCodeOffset] = code ^ xor
			code = 1
			outCodeOffset = outOffset
			outOffset = outOffset + 1
		}
	}
	output[outCodeOffset] = code ^ xor
	return output[outStartOffset:outOffset], nil
}
