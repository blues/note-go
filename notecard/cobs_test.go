package notecard

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCob(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	min := 100
	max := 1000
	len := rand.Intn(max-min+1) + min
	buf := make([]byte, len)
	xor := byte(rand.Int())

	_, err := rand.Read(buf)
	require.NoError(t, err)

	encoded, err := CobsEncode(buf, xor)
	require.NoError(t, err)

	decoded, err := CobsDecode(encoded, xor)
	require.NoError(t, err)

	require.Equal(t, buf, decoded)
}
