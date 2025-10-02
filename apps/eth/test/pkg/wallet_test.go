package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	v, err := parse("0x2710")
	require.NoError(t, err)
	require.Equal(t, uint64(10000), v)
}
