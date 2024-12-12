package neo4j

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"time"
)

var CertspotterConstraintQueries = []string{
	`DROP CONSTRAINT cert_id_unique IF EXISTS`,
	`CREATE CONSTRAINT cert_id_unique IF NOT EXISTS 
     FOR (c:Certificate) REQUIRE c.cert_id IS UNIQUE`,

	`DROP CONSTRAINT cert_sha256_unique IF EXISTS`,
	`CREATE CONSTRAINT cert_sha256_unique IF NOT EXISTS 
     FOR (c:Certificate) REQUIRE c.cert_sha256 IS UNIQUE`,

	`DROP CONSTRAINT pubkey_sha256_unique IF EXISTS`,
	`CREATE CONSTRAINT pubkey_sha256_unique IF NOT EXISTS 
     FOR (p:PublicKey) REQUIRE p.pubkey_sha256 IS UNIQUE`,

	`DROP CONSTRAINT dns_name_unique IF EXISTS`,
	`CREATE CONSTRAINT dns_name_unique IF NOT EXISTS 
     FOR (d:DnsName) REQUIRE d.name IS UNIQUE`,
}

const CertspotterInsertQuery = `
MERGE (cmd:Command {name: $command_name})
MERGE (run:CommandRun {key: $run_key})
SET run.timestamp = $run_timestamp
MERGE (run)-[:EXECUTED]->(cmd)
WITH run, cmd, $certificates AS certificates
UNWIND certificates AS cert

MERGE (c:Certificate {cert_id: cert.id})
SET c.tbs_sha256 = cert.tbs_sha256
SET c.cert_sha256 = cert.cert_sha256
SET c.not_before = datetime(cert.not_before)
SET c.not_after = datetime(cert.not_after)
SET c.revoked = cert.revoked

MERGE (pk:PublicKey {pubkey_sha256: cert.pubkey_sha256})

MERGE (c)-[:USES]->(pk)

WITH c, run, cert
UNWIND cert.dns_names AS dns_name
MERGE (d:DnsName {name: dns_name})
MERGE (c)-[:SECURES]->(d)

MERGE (run)-[:FOUND]->(c)
MERGE (run)-[:FOUND]->(pk)
WITH d, run
MERGE (run)-[:FOUND]->(d)
`

func CertspotterSetup(ctx context.Context) error {
	return SetupConstraints(ctx, CertspotterConstraintQueries)
}

type CertspotterCert struct {
	ID           string    `json:"id"`
	TbsSHA256    string    `json:"tbs_sha256"`
	CertSHA256   string    `json:"cert_sha256"`
	DNSNames     []string  `json:"dns_names"`
	PubkeySHA256 string    `json:"pubkey_sha256"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	Revoked      bool      `json:"revoked"`
}

func (c CertspotterCert) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":            c.ID,
		"tbs_sha256":    c.TbsSHA256,
		"cert_sha256":   c.CertSHA256,
		"dns_names":     c.DNSNames,
		"pubkey_sha256": c.PubkeySHA256,
		"not_before":    c.NotBefore.Format(time.RFC3339),
		"not_after":     c.NotAfter.Format(time.RFC3339),
		"revoked":       c.Revoked,
	}
}

type InsertCertspotterOpts struct {
	Certificates []CertspotterCert
	RunKey       string
	RunTimestamp time.Time
	CommandName  string
}

func InsertCertspotterRecords(ctx context.Context, opts InsertCertspotterOpts) error {
	err := CertspotterSetup(ctx)
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
		result, err := transaction.Run(ctx, CertspotterInsertQuery, params)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})

	return err
}
