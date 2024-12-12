package neo4j

import (
	"context"
	"github.com/mgorunuch/microb/app/core"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"time"
)

var DnsRecordConstraintQueries = []string{
	`DROP CONSTRAINT hostname_unique IF EXISTS`,
	`CREATE CONSTRAINT hostname_unique IF NOT EXISTS 
     FOR (h:Hostname) REQUIRE h.name IS UNIQUE`,

	`DROP CONSTRAINT address_unique IF EXISTS`,
	`CREATE CONSTRAINT address_unique IF NOT EXISTS 
     FOR (a:Address) REQUIRE a.value IS UNIQUE`,

	`DROP CONSTRAINT record_type_unique IF EXISTS`,
	`CREATE CONSTRAINT record_type_unique IF NOT EXISTS 
     FOR (r:RecordType) REQUIRE r.type IS UNIQUE`,

	`DROP CONSTRAINT command_run_unique IF EXISTS`,
	`CREATE CONSTRAINT command_run_unique IF NOT EXISTS 
     FOR (c:CommandRun) REQUIRE c.key IS UNIQUE`,

	`DROP CONSTRAINT command_unique IF EXISTS`,
	`CREATE CONSTRAINT command_unique IF NOT EXISTS 
     FOR (c:Command) REQUIRE c.name IS UNIQUE`,
}

const DnsRecordInsertQuery = `
MERGE (cmd:Command {name: $command_name})
MERGE (run:CommandRun {key: $run_key})
SET run.timestamp = $run_timestamp
MERGE (run)-[:EXECUTED]->(cmd)
WITH run, cmd, $records AS records
UNWIND records AS record

MERGE (h:Hostname {name: record.hostname})
SET h.asset_type = record.asset_type
SET h.first_seen = CASE 
    WHEN h.first_seen IS NULL OR h.first_seen > record.timestamp 
    THEN record.timestamp 
    ELSE h.first_seen 
END
SET h.last_seen = CASE 
    WHEN h.last_seen IS NULL OR h.last_seen < record.timestamp 
    THEN record.timestamp 
    ELSE h.last_seen 
END

MERGE (a:Address {value: record.address})

MERGE (rt:RecordType {type: record.record_type})

MERGE (h)-[r:HAS_DNS_RECORD]->(a)
SET r.record_type = record.record_type
SET r.first_seen = CASE 
    WHEN r.first_seen IS NULL OR r.first_seen > record.timestamp 
    THEN record.timestamp 
    ELSE r.first_seen 
END
SET r.last_seen = CASE 
    WHEN r.last_seen IS NULL OR r.last_seen < record.timestamp 
    THEN record.timestamp 
    ELSE r.last_seen 
END

MERGE (run)-[:FOUND]->(h)
MERGE (run)-[:FOUND]->(a)
`

func DbsRecordSetup(ctx context.Context) error {
	return SetupConstraints(ctx, DnsRecordConstraintQueries)
}

type InsertDnsRecordOpts struct {
	Records      []DnsRecord
	RunKey       string
	RunTimestamp time.Time
	CommandName  string
}

func InsertDNSRecords(ctx context.Context, opts InsertDnsRecordOpts) error {
	core.FatalErr(DbsRecordSetup(ctx))

	session := Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer core.CtxCloser(ctx, session.Close)()

	recordMaps := make([]map[string]interface{}, len(opts.Records))
	for i, record := range opts.Records {
		recordMaps[i] = record.ToMap()
	}

	params := map[string]interface{}{
		"records":       recordMaps,
		"run_key":       opts.RunKey,
		"run_timestamp": opts.RunTimestamp.Unix(),
		"command_name":  opts.CommandName,
	}

	_, err := session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx, DnsRecordInsertQuery, params)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})

	return err
}
