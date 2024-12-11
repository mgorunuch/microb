package google_custom_search

import (
	"encoding/json"
	"fmt"
	"github.com/mgorunuch/microb/app/core"
	"net/http"
	"net/url"
)

func GOOGLE_CUSTOM_SEARCH_API() string {
	return core.Env.Get("GOOGLE_CUSTOM_SEARCH_API", true)
}

func GOOGLE_CUSTOM_SEARCH_ENGINE_ID() string {
	return core.Env.Get("GOOGLE_CUSTOM_SEARCH_ENGINE_ID", true)
}

type GoogleCustomSearchResponse struct {
	Items []struct {
		Title       string `json:"title"`
		Link        string `json:"link"`
		Snippet     string `json:"snippet"`
		DisplayLink string `json:"displayLink"`
	} `json:"items"`
}

func Run(query string) (GoogleCustomSearchResponse, error) {
	googleCustomSearchApiKey := GOOGLE_CUSTOM_SEARCH_API()
	googleCustomSearchEngineId := GOOGLE_CUSTOM_SEARCH_ENGINE_ID()

	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{
		"key": {googleCustomSearchApiKey},
		"cx":  {googleCustomSearchEngineId},
		"q":   {query},
	}

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	resp, err := http.Get(requestURL)
	if err != nil {
		return GoogleCustomSearchResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GoogleCustomSearchResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var searchResponse GoogleCustomSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return GoogleCustomSearchResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return searchResponse, nil
}
