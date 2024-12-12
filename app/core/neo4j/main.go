package neo4j

import (
	"context"
	"github.com/mgorunuch/microb/app/core"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"time"
)

func NEO4J_DB_URI() string {
	return core.Env.GetDefault("NEO4J_DB_URI", "neo4j://localhost")
}

func NEO4J_DB_USER() string {
	return core.Env.GetDefault("NEO4J_DB_USER", "neo4j")
}

func NEO4J_DB_PASSWORD() string {
	return core.Env.Get("NEO4J_DB_PASSWORD", true)
}

var Driver neo4j.DriverWithContext

func Init(ctx context.Context) func() error {
	dbUri := NEO4J_DB_URI()
	dbUser := NEO4J_DB_USER()
	dbPassword := NEO4J_DB_PASSWORD()
	Driver = core.Fatal1Err(neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, "")))
	core.FatalErr(Driver.VerifyConnectivity(ctx))
	return func() error { return Driver.Close(ctx) }
}

func SetupConstraints(ctx context.Context, constraintQueries []string) error {
	session := Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer core.CtxCloser(ctx, session.Close)()

	for _, query := range constraintQueries {
		_, err := session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx, query, nil)
			if err != nil {
				return nil, err
			}
			return result.Consume(ctx)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type DnsRecord struct {
	Hostname   string    `json:"hostname"`
	Address    string    `json:"address"`
	RecordType string    `json:"record_type"`
	AssetType  string    `json:"asset_type"`
	Timestamp  time.Time `json:"timestamp"`
}

func (d DnsRecord) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"hostname":    d.Hostname,
		"address":     d.Address,
		"record_type": d.RecordType,
		"asset_type":  d.AssetType,
		"timestamp":   d.Timestamp.Unix(),
	}
}
