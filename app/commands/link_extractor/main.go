package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func isValidURL(urlStr string) bool {
	// Skip data URIs and obviously invalid URLs
	if strings.HasPrefix(urlStr, "data:") || strings.Count(urlStr, "/") > 10 {
		return false
	}

	// Try parsing the URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Must have a scheme or be a relative path
	return u.Scheme != "" || strings.HasPrefix(urlStr, "/") || strings.HasPrefix(urlStr, ".")
}

func extractLinks(text string) []string {
	var links []string

	// First unescape any \x sequences and unicode escapes
	unescaped := text
	hexEscapeRegex := regexp.MustCompile(`\\x([0-9a-fA-F]{2})|\\u([0-9a-fA-F]{4})`)
	unescaped = hexEscapeRegex.ReplaceAllStringFunc(unescaped, func(match string) string {
		if strings.HasPrefix(match, `\x`) {
			hex := match[2:] // Skip \x
			val, err := strconv.ParseUint(hex, 16, 8)
			if err != nil {
				return match
			}
			return string(rune(val))
		} else { // \u case
			hex := match[2:] // Skip \u
			val, err := strconv.ParseUint(hex, 16, 16)
			if err != nil {
				return match
			}
			return string(rune(val))
		}
	})

	// Match all string literals
	stringRegex := regexp.MustCompile(`(['"])(.*?)(['"])`)
	matches := stringRegex.FindAllStringSubmatch(unescaped, -1)
	for _, match := range matches {
		content := match[2] // Get the content between quotes

		// Extract URLs from the content using regex
		urlRegex := regexp.MustCompile(`(?i)(?:https?:\/\/|\/\/)[^\s<>"']+|(?:\.{0,2}\/)[^\s<>"']+\.(?:js|css|html|png|jpg|jpeg|gif|ico)`)
		urlMatches := urlRegex.FindAllString(content, -1)

		for _, url := range urlMatches {
			url = strings.Trim(url, `"',.`)
			if url != "" && isValidURL(url) {
				links = append(links, url)
			}
		}
	}

	// Also look for URLs directly in the text (outside of quotes)
	urlRegex := regexp.MustCompile(`(?i)(?:https?:\/\/|\/\/)[^\s<>"']+|(?:\.{0,2}\/)[^\s<>"']+\.(?:js|css|html|png|jpg|jpeg|gif|ico)`)
	urlMatches := urlRegex.FindAllString(unescaped, -1)

	for _, url := range urlMatches {
		url = strings.Trim(url, `"',.`)
		if url != "" && isValidURL(url) {
			links = append(links, url)
		}
	}

	return links
}

func extractHTMLLinks(n *html.Node, prefix string) []string {
	var links []string

	if n.Type == html.ElementNode {
		var attr string
		switch n.Data {
		case "a", "link":
			attr = "href"
		case "script", "img", "iframe", "embed", "source", "track":
			attr = "src"
		case "form":
			attr = "action"
		}

		if attr != "" {
			for _, a := range n.Attr {
				if a.Key == attr {
					url := a.Val
					if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "//") && prefix != "" {
						url = prefix + url
					}
					if isValidURL(url) {
						links = append(links, url)
					}
					break
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = append(links, extractHTMLLinks(c, prefix)...)
	}

	return links
}

func prefixIsIgnored(link string) bool {
	ignoredPrefixData, err := os.ReadFile("libs/ignored_link_prefixes.txt")
	if err != nil {
		return false
	}
	ignoredPrefixes := strings.Split(string(ignoredPrefixData), "\n")

	for _, prefix := range ignoredPrefixes {
		if strings.HasPrefix(prefix, "#") || prefix == "" {
			continue
		}
		if strings.HasPrefix(link, strings.TrimSpace(prefix)) {
			return true
		}
	}
	return false
}

func printLink(link string) {
	if prefixIsIgnored(link) {
		return
	}
	fmt.Println(link)
}
func main() {
	prefix := flag.String("prefix", "", "Default URL prefix for relative paths")
	flag.Parse()

	// Try parsing as HTML first
	input := bufio.NewReader(os.Stdin)
	content, err := input.Peek(1024) // Peek at first 1024 bytes

	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Check if content looks like HTML by searching for HTML tags
	isHTML := false
	if len(content) > 0 {
		for i := 0; i < len(content)-1; i++ {
			if content[i] == '<' && content[i+1] != '!' {
				isHTML = true
				break
			}
		}
	}

	if isHTML {
		// Parse as HTML
		doc, err := html.Parse(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing HTML: %v\n", err)
			os.Exit(1)
		}

		// Extract HTML links
		htmlLinks := extractHTMLLinks(doc, *prefix)
		for _, link := range htmlLinks {
			printLink(link)
		}

		// Extract JavaScript links
		var scripts []string
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "script" {
				// Extract inline JavaScript
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					scripts = append(scripts, n.FirstChild.Data)
				}

				// Extract src attribute for external scripts
				for _, a := range n.Attr {
					if a.Key == "src" {
						scripts = append(scripts, a.Val)
						break
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)

		for _, script := range scripts {
			links := extractLinks(script)
			for _, link := range links {
				if !strings.HasPrefix(link, "http") && !strings.HasPrefix(link, "//") && *prefix != "" {
					link = *prefix + link
				}
				printLink(link)
			}
		}
	} else {
		// Treat as pure JavaScript
		script, err := io.ReadAll(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading JavaScript: %v\n", err)
			os.Exit(1)
		}

		links := extractLinks(string(script))
		for _, link := range links {
			if !strings.HasPrefix(link, "http") && !strings.HasPrefix(link, "//") && *prefix != "" {
				link = *prefix + link
			}
			printLink(link)
		}
	}
}
