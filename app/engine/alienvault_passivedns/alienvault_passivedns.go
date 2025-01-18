package alienvault_passivedns

import (
	"context"
	"encoding/json"
	"fmt"
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
func Get(_ context.Context, hostname string) (res PassiveDnsResp, err error) {
	url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", hostname)

	response, err := http.Get(url)
	if err != nil {
		return res, fmt.Errorf("failed to make API request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return res, fmt.Errorf("API request failed with status: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return res, fmt.Errorf("failed to read API response: %w", err)
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return res, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	return res, nil
}
