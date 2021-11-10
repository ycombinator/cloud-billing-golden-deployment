package replayer

import (
	"time"

	es "github.com/elastic/go-elasticsearch/v7"
)

type Replayer struct {
	esClient es.Client
	log      replayLog

	Errors []error
}

func NewReplayer(esClient es.Client, rawLog []byte) (*Replayer, error) {
	r := new(Replayer)

	log, err := newReplayLog(rawLog)
	if err != nil {
		return nil, err
	}

	r.log = *log
	r.esClient = esClient

	return r, nil
}

func (r *Replayer) Start(done chan interface{}) {
	var elapsed int

	if err := r.replayLogEntriesAtOffset(elapsed); err != nil {
		r.Errors = append(r.Errors, err)
	}

	if r.log.size() == 0 {
		close(done)
	}

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			elapsed += 1
			if err := r.replayLogEntriesAtOffset(elapsed); err != nil {
				r.Errors = append(r.Errors, err)
			}

			if r.log.size() == 0 {
				ticker.Stop()
				close(done)
			}
		}
	}
}

func (r *Replayer) replayLogEntriesAtOffset(offset int) error {
	entries := r.log.popAtOffset(offset)

	for _, entry := range entries {
		switch entry.Op {
		case "search":
			o := searchOperation(entry)
			return o.Do(r.esClient)
		case "index":
			o := indexOperation(entry)
			return o.Do(r.esClient)
		}
	}

	return nil
}
