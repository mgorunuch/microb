package neo4j

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"time"
)

var CrtshConstraintQueries = []string{
	`DROP CONSTRAINT certificate_id_unique IF EXISTS`,
	`CREATE CONSTRAINT certificate_id_unique IF NOT EXISTS 
     FOR (c:Certificate) REQUIRE c.cert_id IS UNIQUE`,

	`DROP CONSTRAINT issuer_name_unique IF EXISTS`,
	`CREATE CONSTRAINT issuer_name_unique IF NOT EXISTS 
     FOR (i:Issuer) REQUIRE i.name IS UNIQUE`,

	`DROP CONSTRAINT dns_name_unique IF EXISTS`,
	`CREATE CONSTRAINT dns_name_unique IF NOT EXISTS 
     FOR (d:DnsName) REQUIRE d.name IS UNIQUE`,
}

const CrtshInsertQuery = `
MERGE (cmd:Command {name: $command_name})
MERGE (run:CommandRun {key: $run_key})
SET run.timestamp = $run_timestamp
MERGE (run)-[:EXECUTED]->(cmd)
WITH run, cmd, $certificates AS certificates
UNWIND certificates AS cert

MERGE (c:Certificate {cert_id: cert.id})
SET c.serial_number = cert.serial_number
SET c.not_before = datetime(cert.not_before)
SET c.not_after = datetime(cert.not_after)
SET c.entry_timestamp = datetime(cert.entry_timestamp)

MERGE (i:Issuer {name: cert.issuer_name})
SET i.issuer_ca_id = cert.issuer_ca_id
MERGE (c)-[:ISSUED_BY]->(i)

MERGE (d:DnsName {name: cert.common_name})
MERGE (c)-[:SECURES]->(d)
WITH c, d, cert, run
WHERE cert.name_value IS NOT NULL
UNWIND split(cert.name_value, '\n') AS alt_name
MERGE (alt:DnsName {name: alt_name})
MERGE (c)-[:SECURES]->(alt)

MERGE (run)-[:FOUND]->(c)
MERGE (run)-[:FOUND]->(i)
MERGE (run)-[:FOUND]->(d)
`

func CrtshSetup(ctx context.Context) error {
	return SetupConstraints(ctx, CrtshConstraintQueries)
}

type CrtshCert struct {
	ID             int64     `json:"id"`
	IssuerCAID     int       `json:"issuer_ca_id"`
	IssuerName     string    `json:"issuer_name"`
	CommonName     string    `json:"common_name"`
	NameValue      string    `json:"name_value"`
	SerialNumber   string    `json:"serial_number"`
	EntryTimestamp time.Time `json:"entry_timestamp"`
	NotBefore      time.Time `json:"not_before"`
	NotAfter       time.Time `json:"not_after"`
}

func (c CrtshCert) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":              c.ID,
		"issuer_ca_id":    c.IssuerCAID,
		"issuer_name":     c.IssuerName,
		"common_name":     c.CommonName,
		"name_value":      c.NameValue,
		"serial_number":   c.SerialNumber,
		"entry_timestamp": c.EntryTimestamp.Format(time.RFC3339),
		"not_before":      c.NotBefore.Format(time.RFC3339),
		"not_after":       c.NotAfter.Format(time.RFC3339),
	}
}

type InsertCrtshOpts struct {
	Certificates []CrtshCert
	RunKey       string
	RunTimestamp time.Time
	CommandName  string
}

func InsertCrtshRecords(ctx context.Context, opts InsertCrtshOpts) error {
	err := CrtshSetup(ctx)
	if err != nil {
		return err
	}

	session := Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	certMaps := make([]map[string]interface{}, len(opts.Certificates))
	for i, cert := range opts.Certificates {
		certMaps[i] = cert.ToMap()
	}

	params := map[string]interface{}{
		"certificates":  certMaps,
		"run_key":       opts.RunKey,
		"run_timestamp": opts.RunTimestamp.Unix(),
		"command_name":  opts.CommandName,
	}

	_, err = session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx, CrtshInsertQuery, params)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})

	return err
}
