package google_custom_search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type GoogleCustomSearchResponse struct {
	Items []struct {
		Title       string `json:"title"`
		Link        string `json:"link"`
		Snippet     string `json:"snippet"`
		DisplayLink string `json:"displayLink"`
	} `json:"items"`
}

func Run(googleCustomSearchApiKey string, googleCustomSearchEngineId string, query string) (*GoogleCustomSearchResponse, error) {
	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{
		"key": {googleCustomSearchApiKey},
		"cx":  {googleCustomSearchEngineId},
		"q":   {query},
	}

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var searchResponse GoogleCustomSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResponse, nil
}
