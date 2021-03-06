package usage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/logging"
	"go.uber.org/zap"

	es "github.com/elastic/go-elasticsearch/v7"
)

var (
	connectionSingleton *Connection
)

type Connection struct {
	esClient *es.Client
}

type Query struct {
	ClusterIDs []string
	From       string
	To         string
}

func (q *Query) toElasticsearchFilters(clusterIDFieldName, timestampFieldName string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"terms": map[string][]string{
				clusterIDFieldName: q.ClusterIDs,
			},
		},
		{
			"range": map[string]interface{}{
				timestampFieldName: map[string]string{
					"gte": q.From,
					"lt":  q.To,
				},
			},
		},
	}
}

func NewConnection(address, username, password string) (*Connection, error) {
	if connectionSingleton == nil {
		c, err := es.NewClient(es.Config{
			Addresses: []string{address},
			Username:  username,
			Password:  password,
		})
		if err != nil {
			return nil, err
		}

		connectionSingleton = new(Connection)
		connectionSingleton.esClient = c
	}

	return connectionSingleton, nil
}

func (c *Connection) GetInstanceCapacityGBHours(q Query) (float64, error) {
	// TODO: refactor to PG package as this will need a PostgresQL connection and query
	return 0, nil
}

func (c *Connection) GetDataOutGB(q Query) (float64, error) {
	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": q.toElasticsearchFilters("cluster_id.keyword", "@timestamp"),
			},
		},
		"aggs": map[string]interface{}{
			"total": map[string]interface{}{
				"sum": map[string]string{
					"field": "out.value",
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return 0, fmt.Errorf("error encoding query: %w", err)
	}

	logging.Logger.Debug("data out query", zap.String("body", buf.String()))

	// Perform the search request.
	res, err := c.esClient.Search(
		c.esClient.Search.WithContext(context.Background()),
		c.esClient.Search.WithIndex("aggregations-proxy-metering-*"),
		c.esClient.Search.WithBody(&buf),
		c.esClient.Search.WithSize(0),
	)
	if err != nil {
		return 0, fmt.Errorf("error getting response: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return 0, fmt.Errorf("error parsing the response body: %w", err)
		} else {
			errStr := fmt.Errorf("query error: status: [%s], type: [%s], reason: [%s]",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
			return 0, errStr
		}
	}

	var r struct {
		Aggregations struct {
			Total struct {
				Value float64 `json:"value"`
			} `json:"total"`
		} `json:"aggregations"`
	}

	logging.Logger.Debug("data out requests response", zap.String("body", res.String()))

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return 0, fmt.Errorf("error parsing the response body: %w", err)
	}

	return r.Aggregations.Total.Value, nil
}

func (c *Connection) GetDataInterNodeGB(q Query) (float64, error) {
	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": q.toElasticsearchFilters("deployment_id.keyword", "@timestamp"),
			},
		},
		"aggs": map[string]interface{}{
			"total": map[string]interface{}{
				"sum": map[string]string{
					"field": "out.value",
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return 0, fmt.Errorf("error encoding query: %w", err)
	}

	logging.Logger.Debug("data internode query", zap.String("body", buf.String()))

	// Perform the search request.
	res, err := c.esClient.Search(
		c.esClient.Search.WithContext(context.Background()),
		c.esClient.Search.WithIndex("aggregations-data-transfer-*"),
		c.esClient.Search.WithBody(&buf),
		c.esClient.Search.WithSize(1),
	)
	if err != nil {
		return 0, fmt.Errorf("error getting response: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return 0, fmt.Errorf("error parsing the response body: %w", err)
		} else {
			errStr := fmt.Errorf("query error: status: [%s], type: [%s], reason: [%s]",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
			return 0, errStr
		}
	}

	var r struct {
		Aggregations struct {
			Total struct {
				Value float64 `json:"value"`
			} `json:"total"`
		} `json:"aggregations"`
	}

	logging.Logger.Debug("data internode response", zap.String("body", res.String()))

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return 0, fmt.Errorf("error parsing the response body: %w", err)
	}

	return r.Aggregations.Total.Value, nil
}

func (c *Connection) GetSnapshotStorageSizeGB(q Query) (float64, error) {
	// TODO
	return 0, nil
}
func (c *Connection) GetSnapshotAPIRequestsCount(q Query) (float64, error) {
	// Build the request body.
	usageTypeFilter := map[string]interface{}{
		"term": map[string]string{
			"ece.usage.type": "storage_api",
		},
	}
	filters := q.toElasticsearchFilters("ece.source.cluster", "@timestamp")
	filters = append(filters, usageTypeFilter)

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filters,
			},
		},
		"aggs": map[string]interface{}{
			"total": map[string]interface{}{
				"sum": map[string]string{
					"field": "ece.usage.count",
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return 0, fmt.Errorf("error encoding query: %w", err)
	}

	logging.Logger.Debug("snapshot api requests query", zap.String("body", buf.String()))

	// Perform the search request.
	res, err := c.esClient.Search(
		c.esClient.Search.WithContext(context.Background()),
		c.esClient.Search.WithIndex("usage-v*"),
		c.esClient.Search.WithBody(&buf),
		c.esClient.Search.WithSize(1),
	)
	if err != nil {
		return 0, fmt.Errorf("error getting response: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return 0, fmt.Errorf("error parsing the response body: %w", err)
		} else {
			errStr := fmt.Errorf("query error: status: [%s], type: [%s], reason: [%s]",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
			return 0, errStr
		}
	}

	var r struct {
		Aggregations struct {
			Total struct {
				Value float64 `json:"value"`
			} `json:"total"`
		} `json:"aggregations"`
	}

	logging.Logger.Debug("snapshot api requests response", zap.String("body", res.String()))

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return 0, fmt.Errorf("error parsing the response body: %w", err)
	}

	return r.Aggregations.Total.Value, nil
}
