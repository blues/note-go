// Copyright 2023 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

// Decode with optional XOR
func CobsDecode(input []byte, xor byte) ([]byte, error) {
	output := make([]byte, len(input))
	length := len(output)
	inOffset := 0
	outOffset := inOffset
	startOffset, endOffset := outOffset, inOffset+length
	var code, copy uint8 = 0xFF, 0
	for ; inOffset < endOffset; copy-- {
		if copy != 0 {
			output[outOffset] = input[inOffset] ^ xor
			outOffset, inOffset = outOffset+1, inOffset+1
		} else {
			if code != 0xFF {
				output[outOffset] = 0
				outOffset = outOffset + 1
			}
			code = input[inOffset] ^ xor
			copy, inOffset = code, inOffset+1
			if code == 0 {
				break
			}
		}
	}
	return output[startOffset:outOffset], nil
}

// Get the maximum size of the cobs-encoded buffer
func CobsEncodedLength(length int) int {
	return length + (1 + (length / 254))
}

// Encode with optional XOR
func CobsEncode(input []byte, xor byte) ([]byte, error) {
	length := len(input)
	inOffset := 0
	// Allocate with +1 capacity so append(result, '\n') won't reallocate
	maxLen := CobsEncodedLength(len(input))
	output := make([]byte, maxLen, maxLen+1)
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

// CobsEncodeAppend encodes data and appends a delimiter in one operation.
// This avoids the reallocation that would occur with append(CobsEncode(...), delim).
func CobsEncodeAppend(input []byte, xor byte, delimiter byte) ([]byte, error) {
	encoded, err := CobsEncode(input, xor)
	if err != nil {
		return nil, err
	}
	// Since CobsEncode allocates with +1 capacity, this append won't reallocate
	return append(encoded, delimiter), nil
}
