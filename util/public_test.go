package util

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicAddr(t *testing.T) {
	host, _ := os.Hostname()

	addr := ":8080"
	exp := fmt.Sprintf("http://%s:8080", host)

	res, err := PublicAddr(addr)
	require.NoError(t, err)
	require.Equal(t, exp, res)
}
