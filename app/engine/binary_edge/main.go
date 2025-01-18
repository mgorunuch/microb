package binary_edge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mgorunuch/microb/app/core"
)

func BINARYEDGE_API_KEY() string {
	return core.Env.Get("BINARYEDGE_API_KEY", true)
}

type BinaryEdgeResponse struct {
	Query    string   `json:"query"`
	Page     int      `json:"page"`
	PageSize int      `json:"pagesize"`
	Total    int      `json:"total"`
	Events   []string `json:"events"`
}

func Run(_ context.Context, domain string) (res BinaryEdgeResponse, err error) {
	baseURL := "https://api.binaryedge.io/v2/query/domains/subdomain"
	requestURL := fmt.Sprintf("%s/%s", baseURL, url.PathEscape(domain))
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-Key", BINARYEDGE_API_KEY())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return res, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return res, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode response: %w", err)
	}

	return res, nil
}
