package replayer

import (
	"testing"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/require"
)

func TestReplayer(t *testing.T) {
	esClient, err := es.NewDefaultClient()
	require.NoError(t, err)

	r, err := NewReplayer(*esClient, rawLog)
	require.NoError(t, err)
	require.NotNil(t, r)

	done := make(chan interface{})
	go r.Start(done)

	<-done
	require.Len(t, r.Errors, 0)
}
