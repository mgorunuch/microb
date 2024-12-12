package neo4j

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"time"
)

var CommonCrawlConstraintQueries = []string{
	`DROP CONSTRAINT webpage_url_unique IF EXISTS`,
	`CREATE CONSTRAINT webpage_url_unique IF NOT EXISTS 
     FOR (w:Webpage) REQUIRE w.url IS UNIQUE`,

	`DROP CONSTRAINT webpage_urlkey_unique IF EXISTS`,
	`CREATE CONSTRAINT webpage_urlkey_unique IF NOT EXISTS 
     FOR (w:Webpage) REQUIRE w.urlkey IS UNIQUE`,
}

const CommonCrawlInsertQuery = `
MERGE (cmd:Command {name: $command_name})
MERGE (run:CommandRun {key: $run_key})
SET run.timestamp = $run_timestamp
MERGE (run)-[:EXECUTED]->(cmd)
WITH run, cmd, $webpages AS webpages
UNWIND webpages AS webpage

MERGE (w:Webpage {urlkey: webpage.urlkey})
SET w.url = webpage.url
SET w.mime = webpage.mime
SET w.mime_detected = webpage.mime_detected
SET w.status = webpage.status
SET w.digest = webpage.digest
SET w.length = webpage.length
SET w.offset = webpage.offset
SET w.filename = webpage.filename
SET w.languages = webpage.languages
SET w.encoding = webpage.encoding
SET w.timestamp = webpage.timestamp
SET w.first_seen = CASE 
    WHEN w.first_seen IS NULL OR w.first_seen > webpage.timestamp 
    THEN webpage.timestamp 
    ELSE w.first_seen 
END
SET w.last_seen = CASE 
    WHEN w.last_seen IS NULL OR w.last_seen < webpage.timestamp 
    THEN webpage.timestamp 
    ELSE w.last_seen 
END

MERGE (run)-[:FOUND]->(w)
`

func CommonCrawlSetup(ctx context.Context) error {
	return SetupConstraints(ctx, CommonCrawlConstraintQueries)
}

type CommonCrawlWebpage struct {
	Urlkey       string    `json:"urlkey"`
	Timestamp    time.Time `json:"timestamp"`
	URL          string    `json:"url"`
	Mime         string    `json:"mime"`
	MimeDetected string    `json:"mime_detected"`
	Status       string    `json:"status"`
	Digest       string    `json:"digest"`
	Length       string    `json:"length"`
	Offset       string    `json:"offset"`
	Filename     string    `json:"filename"`
	Languages    string    `json:"languages"`
	Encoding     string    `json:"encoding"`
}

func (w CommonCrawlWebpage) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"urlkey":        w.Urlkey,
		"url":           w.URL,
		"mime":          w.Mime,
		"mime_detected": w.MimeDetected,
		"status":        w.Status,
		"digest":        w.Digest,
		"length":        w.Length,
		"offset":        w.Offset,
		"filename":      w.Filename,
		"languages":     w.Languages,
		"encoding":      w.Encoding,
		"timestamp":     w.Timestamp.Unix(),
	}
}

type InsertCommonCrawlOpts struct {
	Webpages     []CommonCrawlWebpage
	RunKey       string
	RunTimestamp time.Time
	CommandName  string
}

func InsertCommonCrawlRecords(ctx context.Context, opts InsertCommonCrawlOpts) error {
	err := CommonCrawlSetup(ctx)
	if err != nil {
		return err
	}

	session := Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	webpageMaps := make([]map[string]interface{}, len(opts.Webpages))
	for i, webpage := range opts.Webpages {
		webpageMaps[i] = webpage.ToMap()
	}

	params := map[string]interface{}{
		"webpages":      webpageMaps,
		"run_key":       opts.RunKey,
		"run_timestamp": opts.RunTimestamp.Unix(),
		"command_name":  opts.CommandName,
	}

	_, err = session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx, CommonCrawlInsertQuery, params)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})

	return err
}
