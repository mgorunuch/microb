package certspotter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Issuance struct {
	ID           string    `json:"id"`
	TbsSHA256    string    `json:"tbs_sha256"`
	CertSHA256   string    `json:"cert_sha256"`
	DNSNames     []string  `json:"dns_names"`
	PubkeySHA256 string    `json:"pubkey_sha256"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	Revoked      bool      `json:"revoked"`
}

func Get(domain string) ([]Issuance, error) {
	url := fmt.Sprintf("https://api.certspotter.com/v1/issuances?domain=%s&include_subdomains=true&expand=dns_names", domain)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get issuances: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var issuances []Issuance
	if err := json.Unmarshal(body, &issuances); err != nil {
		return nil, err
	}

	return issuances, nil
}
