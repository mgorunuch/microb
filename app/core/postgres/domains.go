package postgres

import (
	"context"
	"time"
)

type Domain struct {
	ID        int
	Domain    string
	Reasons   []string
	CreatedAt time.Time
}

// CreateDomain inserts a new domain record or adds reasons if domain exists
func CreateDomain(ctx context.Context, domain string, reasons []string) (*Domain, error) {
	var d Domain
	err := Pool.QueryRow(ctx,
		`insert into domains (domain, reasons) 
		values ($1, $2)
		on conflict (domain) do update 
		set reasons = array(
			select distinct unnest(domains.reasons || excluded.reasons)
		)
		returning id, domain, reasons, created_at`,
		domain, reasons,
	).Scan(&d.ID, &d.Domain, &d.Reasons, &d.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &d, nil
}

// GetDomain retrieves a domain record by its domain name
func GetDomain(ctx context.Context, domain string) (*Domain, error) {
	var d Domain
	err := Pool.QueryRow(ctx,
		`SELECT id, domain, reasons, created_at 
		FROM domains 
		WHERE domain = $1`,
		domain,
	).Scan(&d.ID, &d.Domain, &d.Reasons, &d.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDomains retrieves all domain records with optional pagination
func ListDomains(ctx context.Context, limit, offset int) ([]Domain, error) {
	rows, err := Pool.Query(ctx,
		`SELECT id, domain, reasons, created_at 
		FROM domains 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var d Domain
		if err := rows.Scan(&d.ID, &d.Domain, &d.Reasons, &d.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

// DeleteDomain removes a domain record by its domain name
func DeleteDomain(ctx context.Context, domain string) error {
	_, err := Pool.Exec(ctx,
		`DELETE FROM domains 
		WHERE domain = $1`,
		domain,
	)
	return err
}

// UpdateDomainReason updates the reasons for a domain
func UpdateDomainReason(ctx context.Context, domain string, reasons []string) (*Domain, error) {
	var d Domain
	err := Pool.QueryRow(ctx,
		`UPDATE domains 
		SET reasons = $2 
		WHERE domain = $1 
		RETURNING id, domain, reasons, created_at`,
		domain, reasons,
	).Scan(&d.ID, &d.Domain, &d.Reasons, &d.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &d, nil
}
