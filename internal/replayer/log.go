package replayer

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type replayLog struct {
	raw []byte

	entries map[int][]operation
}

func newReplayLog(raw []byte) (*replayLog, error) {
	rl := new(replayLog)

	rl.raw = raw
	if err := rl.parseRaw(); err != nil {
		return nil, fmt.Errorf("could not parse replay log: %w", err)
	}

	return rl, nil
}

func (rl *replayLog) parseRaw() error {
	type line struct {
		Offset int `json:"offset"`
		operation
	}

	rl.entries = map[int][]operation{}

	lines := bytes.Split(rl.raw, []byte("\n"))
	for _, rawLine := range lines {
		rawLine = bytes.TrimSpace(rawLine)
		if len(rawLine) == 0 {
			continue
		}

		var l line
		if err := json.Unmarshal(rawLine, &l); err != nil {
			return err
		}

		if _, exists := rl.entries[l.Offset]; !exists {
			rl.entries[l.Offset] = []operation{}
		}

		rl.entries[l.Offset] = append(rl.entries[l.Offset], l.operation)
	}

	return nil
}

func (rl *replayLog) popAtOffset(offset int) []operation {
	if entries, exist := rl.entries[offset]; exist {
		delete(rl.entries, offset)
		return entries
	} else {
		return []operation{}
	}
}

func (rl *replayLog) size() int {
	return len(rl.entries)
}
