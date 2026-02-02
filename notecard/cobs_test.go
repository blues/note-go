package notecard

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCob(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	min := 100
	max := 1000
	len := rng.Intn(max-min+1) + min
	buf := make([]byte, len)
	xor := byte(rng.Int())

	_, err := rng.Read(buf)
	require.NoError(t, err)

	encoded, err := CobsEncode(buf, xor)
	require.NoError(t, err)

	decoded, err := CobsDecode(encoded, xor)
	require.NoError(t, err)

	require.Equal(t, buf, decoded)
}

func TestCobsEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		xor   byte
	}{
		{"empty", []byte{}, 0},
		{"single zero", []byte{0}, 0},
		{"single nonzero", []byte{1}, 0},
		{"two zeros", []byte{0, 0}, 0},
		{"trailing zero", []byte{1, 2, 0}, 0},
		{"leading zero", []byte{0, 1, 2}, 0},
		{"middle zero", []byte{1, 0, 2}, 0},
		{"no zeros", []byte{1, 2, 3}, 0},
		{"all zeros 3", []byte{0, 0, 0}, 0},
		{"with xor", []byte{1, 2, 0, 3}, '\n'},
		{"254 bytes no zero", make254NonZero(), 0},
		{"255 bytes no zero", make255NonZero(), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := CobsEncode(tt.input, tt.xor)
			require.NoError(t, err, "encode failed")

			decoded, err := CobsDecode(encoded, tt.xor)
			require.NoError(t, err, "decode failed")

			require.Equal(t, tt.input, decoded, "roundtrip failed: encoded=%v", encoded)
		})
	}
}

func make254NonZero() []byte {
	b := make([]byte, 254)
	for i := range b {
		b[i] = byte(i%255) + 1
	}
	return b
}

func make255NonZero() []byte {
	b := make([]byte, 255)
	for i := range b {
		b[i] = byte(i%255) + 1
	}
	return b
}

// TestCobsKnownValues tests encoding against known expected output values.
// These are the canonical COBS encodings per the specification.
// This catches regressions where encode/decode are broken in compatible but wrong ways.
func TestCobsKnownValues(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		xor      byte
		expected []byte
	}{
		// Standard COBS test vectors (xor=0)
		{
			name:     "single zero",
			input:    []byte{0x00},
			xor:      0,
			expected: []byte{0x01, 0x01},
		},
		{
			name:     "single nonzero",
			input:    []byte{0x01},
			xor:      0,
			expected: []byte{0x02, 0x01},
		},
		{
			name:     "two zeros",
			input:    []byte{0x00, 0x00},
			xor:      0,
			expected: []byte{0x01, 0x01, 0x01},
		},
		{
			name:     "three nonzero bytes",
			input:    []byte{0x01, 0x02, 0x03},
			xor:      0,
			expected: []byte{0x04, 0x01, 0x02, 0x03},
		},
		{
			name:     "zero in middle",
			input:    []byte{0x01, 0x00, 0x02},
			xor:      0,
			expected: []byte{0x02, 0x01, 0x02, 0x02},
		},
		{
			name:     "leading zero",
			input:    []byte{0x00, 0x01, 0x02},
			xor:      0,
			expected: []byte{0x01, 0x03, 0x01, 0x02},
		},
		{
			name:     "trailing zero",
			input:    []byte{0x01, 0x02, 0x00},
			xor:      0,
			expected: []byte{0x03, 0x01, 0x02, 0x01},
		},
		{
			name:     "Hello",
			input:    []byte{'H', 'e', 'l', 'l', 'o'},
			xor:      0,
			expected: []byte{0x06, 'H', 'e', 'l', 'l', 'o'},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" encode", func(t *testing.T) {
			encoded, err := CobsEncode(tt.input, tt.xor)
			require.NoError(t, err)
			require.Equal(t, tt.expected, encoded, "encoded output mismatch")
		})

		t.Run(tt.name+" decode", func(t *testing.T) {
			decoded, err := CobsDecode(tt.expected, tt.xor)
			require.NoError(t, err)
			require.Equal(t, tt.input, decoded, "decoded output mismatch")
		})
	}
}

// TestCobsXORRoundtrip tests that XOR mode (used to eliminate newlines) roundtrips correctly.
// XOR mode is Blues-specific; there's no external standard, so we only test roundtrip.
func TestCobsXORRoundtrip(t *testing.T) {
	xor := byte('\n') // 0x0A - what note-c/notecard uses

	tests := []struct {
		name  string
		input []byte
	}{
		{"single zero", []byte{0x00}},
		{"single nonzero", []byte{0x01}},
		{"contains newline", []byte{0x01, '\n', 0x02}},
		{"multiple newlines", []byte{'\n', 0x01, '\n', '\n', 0x02, '\n'}},
		{"all newlines", []byte{'\n', '\n', '\n'}},
		{"binary with newlines", func() []byte {
			b := make([]byte, 256)
			for i := range b {
				b[i] = byte(i)
			}
			return b
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := CobsEncode(tt.input, xor)
			require.NoError(t, err)

			// Verify no newlines in encoded output (the whole point of XOR mode)
			for i, b := range encoded {
				require.NotEqual(t, byte('\n'), b, "found newline at position %d in encoded output", i)
			}

			decoded, err := CobsDecode(encoded, xor)
			require.NoError(t, err)
			require.Equal(t, tt.input, decoded, "roundtrip failed")
		})
	}
}

// TestCobs254ByteBoundary tests the critical 254-byte boundary where COBS
// must insert an extra code byte. This is a common source of bugs.
func TestCobs254ByteBoundary(t *testing.T) {
	// 254 non-zero bytes: [0xFF, 254 data bytes, 0x01]
	// The 0xFF means "254 data bytes follow, no implicit zero after"
	// The trailing 0x01 terminates the stream (0 more data bytes)
	data254 := make([]byte, 254)
	for i := range data254 {
		data254[i] = byte(i) + 1 // 1, 2, 3, ..., 254
	}

	encoded254, err := CobsEncode(data254, 0)
	require.NoError(t, err)
	require.Len(t, encoded254, 256, "254 non-zero bytes encode to 256 bytes")
	require.Equal(t, byte(0xFF), encoded254[0], "first code byte should be 0xFF")
	require.Equal(t, byte(0x01), encoded254[255], "trailing code byte should be 0x01")

	decoded254, err := CobsDecode(encoded254, 0)
	require.NoError(t, err)
	require.Equal(t, data254, decoded254)

	// 255 non-zero bytes: [0xFF, 254 data bytes, 0x02, 1 data byte]
	data255 := make([]byte, 255)
	for i := range data255 {
		data255[i] = byte(i) + 1
	}
	data255[254] = 1 // Last byte wraps to 1

	encoded255, err := CobsEncode(data255, 0)
	require.NoError(t, err)
	require.Len(t, encoded255, 257, "255 non-zero bytes encode to 257 bytes")
	require.Equal(t, byte(0xFF), encoded255[0], "first code byte should be 0xFF")
	require.Equal(t, byte(0x02), encoded255[255], "second code byte should be 0x02")

	decoded255, err := CobsDecode(encoded255, 0)
	require.NoError(t, err)
	require.Equal(t, data255, decoded255)

	// 253 non-zero bytes: [0xFE, 253 data bytes] - no extra code byte needed
	data253 := make([]byte, 253)
	for i := range data253 {
		data253[i] = byte(i) + 1
	}

	encoded253, err := CobsEncode(data253, 0)
	require.NoError(t, err)
	require.Len(t, encoded253, 254, "253 non-zero bytes encode to 254 bytes")
	require.Equal(t, byte(0xFE), encoded253[0], "code byte should be 0xFE for 253 data bytes")

	decoded253, err := CobsDecode(encoded253, 0)
	require.NoError(t, err)
	require.Equal(t, data253, decoded253)
}
