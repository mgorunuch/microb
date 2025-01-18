package postgres

import (
	"context"
	"net/url"
	"time"
)

type URL struct {
	ID        string
	Raw       string
	Flags     []string
	Hostname  string
	Path      string
	Scheme    string
	Query     string
	Fragment  string
	CreatedAt time.Time
}

func CreateURL(ctx context.Context, rawURL string, flags []string) (*URL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	var u URL
	err = Pool.QueryRow(ctx,
		`insert into urls (raw, flags, hostname, path, scheme, query, fragment)
		values ($1, $2, $3, $4, $5, $6, $7)
		on conflict (hostname, path, scheme, query, fragment) do update
		set flags = array(
			select distinct unnest(urls.flags || excluded.flags)
		)
		returning id, raw, flags, hostname, path, scheme, query, fragment, created_at`,
		rawURL,
		flags,
		parsedURL.Hostname(),
		parsedURL.Path,
		parsedURL.Scheme,
		parsedURL.RawQuery,
		parsedURL.Fragment,
	).Scan(&u.ID, &u.Raw, &u.Flags, &u.Hostname, &u.Path, &u.Scheme, &u.Query, &u.Fragment, &u.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &u, nil
}
