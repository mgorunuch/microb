package neo4j

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"time"
)

var GoogleSearchConstraintQueries = []string{
	`DROP CONSTRAINT search_result_link_unique IF EXISTS`,
	`CREATE CONSTRAINT search_result_link_unique IF NOT EXISTS 
     FOR (r:SearchResult) REQUIRE r.link IS UNIQUE`,

	`DROP CONSTRAINT website_domain_unique IF EXISTS`,
	`CREATE CONSTRAINT website_domain_unique IF NOT EXISTS 
     FOR (w:Website) REQUIRE w.domain IS UNIQUE`,
}

const GoogleSearchInsertQuery = `
MERGE (cmd:Command {name: $command_name})
MERGE (run:CommandRun {key: $run_key})
SET run.timestamp = $run_timestamp
MERGE (run)-[:EXECUTED]->(cmd)
WITH run, cmd, $results AS results
UNWIND results AS result

MERGE (w:Website {domain: result.display_link})
SET w.first_seen = CASE 
    WHEN w.first_seen IS NULL OR w.first_seen > result.timestamp 
    THEN result.timestamp 
    ELSE w.first_seen 
END
SET w.last_seen = CASE 
    WHEN w.last_seen IS NULL OR w.last_seen < result.timestamp 
    THEN result.timestamp 
    ELSE w.last_seen 
END

MERGE (r:SearchResult {link: result.link})
SET r.title = result.title
SET r.snippet = result.snippet
SET r.first_seen = CASE 
    WHEN r.first_seen IS NULL OR r.first_seen > result.timestamp 
    THEN result.timestamp 
    ELSE r.first_seen 
END
SET r.last_seen = CASE 
    WHEN r.last_seen IS NULL OR r.last_seen < result.timestamp 
    THEN result.timestamp 
    ELSE r.last_seen 
END

MERGE (r)-[:HOSTED_BY]->(w)
MERGE (run)-[:FOUND]->(r)
MERGE (run)-[:FOUND]->(w)
`

func GoogleSearchSetup(ctx context.Context) error {
	return SetupConstraints(ctx, GoogleSearchConstraintQueries)
}

type GoogleSearchResult struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Snippet     string    `json:"snippet"`
	DisplayLink string    `json:"display_link"`
	Timestamp   time.Time `json:"timestamp"`
}

func (r GoogleSearchResult) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"title":        r.Title,
		"link":         r.Link,
		"snippet":      r.Snippet,
		"display_link": r.DisplayLink,
		"timestamp":    r.Timestamp.Unix(),
	}
}

type InsertGoogleSearchOpts struct {
	Results      []GoogleSearchResult
	RunKey       string
	RunTimestamp time.Time
	CommandName  string
}

func InsertGoogleSearchRecords(ctx context.Context, opts InsertGoogleSearchOpts) error {
	err := GoogleSearchSetup(ctx)
	if err != nil {
		return err
	}

	session := Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	resultMaps := make([]map[string]interface{}, len(opts.Results))
	for i, result := range opts.Results {
		resultMaps[i] = result.ToMap()
	}

	params := map[string]interface{}{
		"results":       resultMaps,
		"run_key":       opts.RunKey,
		"run_timestamp": opts.RunTimestamp.Unix(),
		"command_name":  opts.CommandName,
	}

	_, err = session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx, GoogleSearchInsertQuery, params)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})

	return err
}
