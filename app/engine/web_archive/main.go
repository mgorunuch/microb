package web_archive

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
)

// Get fetches the list of URLs from the web archive for the given domain
// Logic: http://web.archive.org/cdx/search/cdx?url=*.{domain}/*&output=text&fl=original&collapse=urlkey
// and returns them as a slice of strings.
func Get(_ context.Context, domain string) ([]string, error) {
	url := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=text&fl=original&collapse=urlkey", domain)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	var urls []string
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
