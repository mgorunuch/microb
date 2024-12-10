package binary_edge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type BinaryEdgeResponse struct {
	Query    string   `json:"query"`
	Page     int      `json:"page"`
	PageSize int      `json:"pagesize"`
	Total    int      `json:"total"`
	Events   []string `json:"events"`
}

func Run(binaryEdgeApiKey string, domain string) (*BinaryEdgeResponse, error) {
	baseURL := "https://api.binaryedge.io/v2/query/domains/subdomain"
	requestURL := fmt.Sprintf("%s/%s", baseURL, url.PathEscape(domain))
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-Key", binaryEdgeApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var searchResponse BinaryEdgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResponse, nil
}
