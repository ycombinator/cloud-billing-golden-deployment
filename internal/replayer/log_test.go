package replayer

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed test_log_1.txt
var rawLog []byte

func TestReplayLog(t *testing.T) {
	rl, err := newReplayLog(rawLog)
	require.NoError(t, err)
	require.Equal(t, rl.size(), 3)
}
