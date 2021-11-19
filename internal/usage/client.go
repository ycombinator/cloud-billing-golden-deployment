package usage

import (
	es "github.com/elastic/go-elasticsearch/v7"
)

var (
	connectionSingleton *Connection
)

type Connection struct {
	esClient *es.Client
}

type Query struct {
	ClusterID string
	From      string
	To        string
}

func NewConnection(address, apiKey string) (*Connection, error) {
	if connectionSingleton == nil {
		c, err := es.NewClient(es.Config{
			Addresses: []string{address},
			APIKey:    apiKey,
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
	return 0, nil
}

func (c *Connection) GetDataOutGB(q Query) (float64, error) {
	return 0, nil
}

func (c *Connection) GetDataInterNodeGB(q Query) (float64, error) {
	return 0, nil
}

func (c *Connection) GetSnapshotStorageSizeGB(q Query) (float64, error) {
	return 0, nil
}
func (c *Connection) GetSnapshotAPIRequestsCount(q Query) (float64, error) {
	return 0, nil
}
