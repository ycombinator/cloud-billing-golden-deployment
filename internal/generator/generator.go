package generator

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"
)

type Generator struct {
	config Config
}

func NewGenerator(cfg Config) *Generator {
	g := new(Generator)
	g.config = cfg

	return g
}

func (g *Generator) Generate() ([]byte, error) {
	var buf []byte
	var offset = g.config.StartOffsetSeconds
	var count int

	for g.keepGenerating(offset, count) {
		op := g.randOp()
		entry := newEntry(offset, op)
		raw, err := marshalEntry(entry)
		if err != nil {
			return nil, err
		}

		buf = append(buf, raw...)

		interval := g.config.MinIntervalSeconds + rand.Intn(g.config.MaxIntervalSeconds-g.config.MinIntervalSeconds)
		offset += interval
		count += 1
	}

	return buf, nil
}

func (g *Generator) keepGenerating(offset, count int) bool {
	if g.config.MaxOffsetSeconds == 0 {
		return count < g.config.MaxCount
	}

	if g.config.MaxCount == 0 {
		return offset < g.config.MaxOffsetSeconds
	}

	return count < g.config.MaxCount && offset < g.config.MaxOffsetSeconds
}

func (g *Generator) randOp() string {
	ops := make([]string, 1+g.config.IndexToSearchRatio)
	ops[0] = "search"
	for i := 1; i < len(ops); i++ {
		ops[i] = "index"
	}

	randIdx := rand.Intn(len(ops))
	return ops[randIdx]
}

func newEntry(offset int, op string) *models.Entry {
	e := models.Entry{
		Offset: offset,
		Operation: models.Operation{
			Op: op,
		},
	}

	var target string
	var body json.RawMessage
	switch op {
	case "index":
		target = "foo"
		body = randIndexBody()
	case "search":
		target = "foo*"
	}

	e.Operation.Target = target
	e.Operation.Body = body

	return &e
}

func marshalEntry(e *models.Entry) ([]byte, error) {
	raw, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	raw = append(raw, byte('\n'))

	return raw, nil
}

func randIndexBody() json.RawMessage {
	messages := []string{
		"the quick brown fox",
		"jumped over the",
		"lazy dog",
	}

	innerKeys := []string{"count", "sum"}

	randMsgIdx := rand.Intn(len(messages))
	randMsg := messages[randMsgIdx]

	randKeyIdx := rand.Intn(len(innerKeys))
	randKey := innerKeys[randKeyIdx]

	randNum := (17 + rand.Intn(10000)) % 523

	bodyTpl := `{"message":"%s","metric":{"%s":%d}}`
	body := fmt.Sprintf(bodyTpl, randMsg, randKey, randNum)

	return json.RawMessage(body)
}
