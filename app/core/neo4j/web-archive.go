package neo4j

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"net/url"
	"path"
	"time"
)

var WebArchiveConstraintQueries = []string{
	`DROP CONSTRAINT url_unique IF EXISTS`,
	`CREATE CONSTRAINT url_unique IF NOT EXISTS 
     FOR (u:URL) REQUIRE u.url IS UNIQUE`,

	`DROP CONSTRAINT website_domain_unique IF EXISTS`,
	`CREATE CONSTRAINT website_domain_unique IF NOT EXISTS 
     FOR (w:Website) REQUIRE w.domain IS UNIQUE`,

	`DROP CONSTRAINT url_path_unique IF EXISTS`,
	`CREATE CONSTRAINT url_path_unique IF NOT EXISTS 
     FOR (p:URLPath) REQUIRE (p.website, p.path) IS UNIQUE`,
}

const WebArchiveInsertQuery = `
MERGE (cmd:Command {name: $command_name})
MERGE (run:CommandRun {key: $run_key})
SET run.timestamp = $run_timestamp
MERGE (run)-[:EXECUTED]->(cmd)
WITH run, cmd, $urls AS urls
UNWIND urls AS url_data

MERGE (w:Website {domain: url_data.domain})
SET w.first_seen = CASE 
    WHEN w.first_seen IS NULL OR w.first_seen > url_data.timestamp 
    THEN url_data.timestamp 
    ELSE w.first_seen 
END
SET w.last_seen = CASE 
    WHEN w.last_seen IS NULL OR w.last_seen < url_data.timestamp 
    THEN url_data.timestamp 
    ELSE w.last_seen 
END

MERGE (p:URLPath {website: url_data.domain, path: url_data.path})
SET p.first_seen = CASE 
    WHEN p.first_seen IS NULL OR p.first_seen > url_data.timestamp 
    THEN url_data.timestamp 
    ELSE p.first_seen 
END
SET p.last_seen = CASE 
    WHEN p.last_seen IS NULL OR p.last_seen < url_data.timestamp 
    THEN url_data.timestamp 
    ELSE p.last_seen 
END

MERGE (u:URL {url: url_data.url})
SET u.scheme = url_data.scheme
SET u.first_seen = CASE 
    WHEN u.first_seen IS NULL OR u.first_seen > url_data.timestamp 
    THEN url_data.timestamp 
    ELSE u.first_seen 
END
SET u.last_seen = CASE 
    WHEN u.last_seen IS NULL OR u.last_seen < url_data.timestamp 
    THEN url_data.timestamp 
    ELSE u.last_seen 
END

MERGE (w)-[:HAS_PATH]->(p)
MERGE (p)-[:HAS_URL]->(u)

MERGE (run)-[:FOUND]->(w)
MERGE (run)-[:FOUND]->(p)
MERGE (run)-[:FOUND]->(u)
`

func WebArchiveSetup(ctx context.Context) error {
	return SetupConstraints(ctx, WebArchiveConstraintQueries)
}

type WebArchiveURL struct {
	URL       string    `json:"url"`
	Domain    string    `json:"domain"`
	Path      string    `json:"path"`
	Scheme    string    `json:"scheme"`
	Timestamp time.Time `json:"timestamp"`
}

func ParseWebArchiveURL(rawURL string, timestamp time.Time) (*WebArchiveURL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// Clean the path to ensure consistent format
	cleanPath := path.Clean(parsedURL.Path)
	if cleanPath == "." {
		cleanPath = "/"
	}

	// Add query parameters to path if they exist
	if parsedURL.RawQuery != "" {
		cleanPath = cleanPath + "?" + parsedURL.RawQuery
	}

	// Add fragment to path if it exists
	if parsedURL.Fragment != "" {
		cleanPath = cleanPath + "#" + parsedURL.Fragment
	}

	return &WebArchiveURL{
		URL:       rawURL,
		Domain:    parsedURL.Host,
		Path:      cleanPath,
		Scheme:    parsedURL.Scheme,
		Timestamp: timestamp,
	}, nil
}

func (w WebArchiveURL) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"url":       w.URL,
		"domain":    w.Domain,
		"path":      w.Path,
		"scheme":    w.Scheme,
		"timestamp": w.Timestamp.Unix(),
	}
}

type InsertWebArchiveOpts struct {
	URLs         []WebArchiveURL
	RunKey       string
	RunTimestamp time.Time
	CommandName  string
}

func InsertWebArchiveRecords(ctx context.Context, opts InsertWebArchiveOpts) error {
	err := WebArchiveSetup(ctx)
	if err != nil {
		return err
	}

	session := Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	urlMaps := make([]map[string]interface{}, len(opts.URLs))
	for i, url := range opts.URLs {
		urlMaps[i] = url.ToMap()
	}

	params := map[string]interface{}{
		"urls":          urlMaps,
		"run_key":       opts.RunKey,
		"run_timestamp": opts.RunTimestamp.Unix(),
		"command_name":  opts.CommandName,
	}

	_, err = session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx, WebArchiveInsertQuery, params)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})

	return err
}
