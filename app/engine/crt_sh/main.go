package crt_sh

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// CertData represents the structure of the certificate data returned by crt.sh
type CertData struct {
	IssuerCaID     int    `json:"issuer_ca_id"`
	IssuerName     string `json:"issuer_name"`
	CommonName     string `json:"common_name"`
	NameValue      string `json:"name_value"`
	ID             int64  `json:"id"`
	EntryTimestamp string `json:"entry_timestamp"`
	NotBefore      string `json:"not_before"`
	NotAfter       string `json:"not_after"`
	SerialNumber   string `json:"serial_number"`
	ResultCount    int    `json:"result_count"`
}

// Get fetches the list of certificates from crt.sh for the given domain
// and returns them as a slice of CertData.
func Get(_ context.Context, domain string) ([]CertData, error) {
	url := fmt.Sprintf("https://crt.sh/?q=*.%s&output=json", domain)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	var certs []CertData
	if err := json.NewDecoder(resp.Body).Decode(&certs); err != nil {
		return nil, err
	}

	return certs, nil
}
