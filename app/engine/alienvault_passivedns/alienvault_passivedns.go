package alienvault_passivedns

import (
	"encoding/json"
	"fmt"
	"github.com/mgorunuch/microb/app/core"
	"io"
	"net/http"
)

type PassiveDns struct {
	Hostname   string `json:"hostname"`
	Address    string `json:"address"`
	RecordType string `json:"record_type"`
	AssetType  string `json:"asset_type"`
}

type PassiveDnsResp struct {
	PassiveDns []PassiveDns `json:"passive_dns"`
}

// Get retrieves passive DNS information for a hostname using the cache provider
// If the data is not in cache, it will fetch from the AlienVault API
func Get(cacheProvider core.CacheProvider[PassiveDnsResp], hostname string) (*PassiveDnsResp, error) {
	// Check if we have this hostname in cache
	hasCached, err := cacheProvider.HasCached(hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to check cache: %w", err)
	}

	// If cached, return cached data
	if hasCached {
		cachedData, err := cacheProvider.GetFromCache(hostname)
		if err != nil {
			return nil, fmt.Errorf("failed to get from cache: %w", err)
		}

		return &cachedData, nil
	}

	// If not cached, fetch from API
	resp, err := fetchFromAPI(hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}

	// Cache the new data
	if err := cacheProvider.AddToCache(hostname, *resp); err != nil {
		return nil, fmt.Errorf("failed to add to cache: %w", err)
	}

	return resp, nil
}

// fetchFromAPI retrieves passive DNS information directly from the AlienVault API
func fetchFromAPI(hostname string) (*PassiveDnsResp, error) {
	url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", hostname)

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	var result PassiveDnsResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	return &result, nil
}
