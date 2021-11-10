package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {
	g := NewGenerator(Config{
		StartOffsetSeconds: 0,
		MaxCount:           20,
		MaxOffsetSeconds:   0,
		MinIntervalSeconds: 1,
		MaxIntervalSeconds: 3,
		IndexToSearchRatio: 4,
	})

	buf, err := g.Generate()
	require.NoError(t, err)
	require.Greater(t, len(buf), 0)
}
