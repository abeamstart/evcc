package public

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddr(t *testing.T) {
	host, _ := os.Hostname()

	addr := ":8080"
	exp := fmt.Sprintf("http://%s:8080", host)

	res, err := SetAddr(addr)
	require.NoError(t, err)
	require.Equal(t, exp, res)
}
