package dao

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func handleESAPIErrorResponse(res *esapi.Response) error {
	var e map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
		return fmt.Errorf("error parsing the response body: %w", err)
	} else {
		if v, ok := e["message"]; ok {
			return fmt.Errorf("[%s] %s", res.Status(), v)
		} else {
			return fmt.Errorf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}
}
