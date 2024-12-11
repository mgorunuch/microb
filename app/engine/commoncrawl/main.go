package commoncrawl

import (
	"encoding/json"
	"fmt"
	"github.com/mgorunuch/microb/app/core"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type CrawlData struct {
	Urlkey       string `json:"urlkey"`
	Timestamp    string `json:"timestamp"`
	Url          string `json:"url"`
	Mime         string `json:"mime"`
	MimeDetected string `json:"mime-detected"`
	Status       string `json:"status"`
	Digest       string `json:"digest"`
	Length       string `json:"length"`
	Offset       string `json:"offset"`
	Filename     string `json:"filename"`
	Languages    string `json:"languages"`
	Encoding     string `json:"encoding"`
}

type LibList struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Timegate string `json:"timegate"`
	CDXAPI   string `json:"cdx-api"`
	From     string `json:"from"`
	To       string `json:"to"`
}

func Get(domain string) ([]CrawlData, error) {
	var libList []LibList

	// Load data from http://index.commoncrawl.org/collinfo.json
	resp, err := http.Get("http://index.commoncrawl.org/collinfo.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load collection info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &libList); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	var mx sync.Mutex
	var allCrawlData []CrawlData

	var wg sync.WaitGroup
	core.SpawnArrayElements(core.RunParallel(&wg, 1, func(line LibList) {
		crawlData, err := crawlLibData(domain, line, allCrawlData)
		if err != nil {
			fmt.Println("Error fetching Common Crawl data:", err)
			return
		}

		fmt.Println("Successfully processed domain:", domain, "from:", line.From, "to:", line.To)

		mx.Lock()
		allCrawlData = append(allCrawlData, crawlData...)
		mx.Unlock()
	}), libList)
	wg.Wait()

	return allCrawlData, nil
}

func crawlLibData(domain string, lib LibList, allCrawlData []CrawlData) ([]CrawlData, error) {
	// Load data from the CDX API
	apiURL := fmt.Sprintf("%s?url=*.%s&output=json", lib.CDXAPI, domain)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to load data from CDX API: %w, %s", err, apiURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from CDX API: %d, %s", resp.StatusCode, apiURL)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from CDX API: %w, %s", err, apiURL)
	}

	response := strings.Join(strings.Split(string(body), "\n"), ",")
	response = strings.TrimSuffix(response, ",")
	response = fmt.Sprintf("[%s]", response)

	var crawlData []CrawlData
	if err := json.Unmarshal([]byte(response), &crawlData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response from CDX API: %w", err)
	}

	allCrawlData = append(allCrawlData, crawlData...)
	return allCrawlData, nil
}
