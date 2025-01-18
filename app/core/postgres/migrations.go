package postgres

import (
	"context"
	"fmt"

	"github.com/mgorunuch/microb/app/core"
)

var migrations = [][2]string{
	{
		"Create schema_migrations table",
		`create table if not exists schema_migrations (
			version integer primary key,
			applied_at timestamp with time zone default current_timestamp
		)`,
	},
	{
		"Create domains table",
		`create table if not exists domains (
			id serial primary key,
			domain text not null,
			reason text,
			created_at timestamp with time zone default current_timestamp,
			constraint domains_domain_unique unique (domain)
		)`,
	},
	{
		"Update domains table to use text array for reasons",
		`alter table domains
			drop column if exists reason,
			add column if not exists reasons text[] default array[]::text[]`,
	},
	{
		"Create chrome_visits table",
		`create table if not exists chrome_visits (
			id uuid primary key default gen_random_uuid(),
			url text not null,
			title text,
			html text,
			opened_at timestamp with time zone default current_timestamp,
			success boolean not null,
			error_msg text,
			reason text not null
		)`,
	},
	{
		"Create urls table",
		`create table if not exists urls (
			id uuid primary key default gen_random_uuid(),
			raw text not null,
			flags text[] default array[]::text[],
			hostname text not null,
			path text not null,
			scheme text not null,
			query text,
			fragment text,
			created_at timestamp with time zone default current_timestamp,
			constraint urls_unique unique (hostname, path, scheme, query, fragment)
		)`,
	},
	{
		"Add url_id to chrome_visits and migrate data",
		`
		-- Add url_id column
		alter table chrome_visits
		add column if not exists url_id uuid references urls(id);

		-- Create urls for existing chrome_visits
		insert into urls (raw, hostname, path, scheme, query, fragment)
		select distinct
			cv.url as raw,
			split_part(split_part(url, '://', 2), '/', 1) as hostname,
			coalesce('/' || nullif(split_part(split_part(url, '://', 2), '/', 2), ''), '/') as path,
			split_part(url, '://', 1) as scheme,
			split_part(split_part(url, '?', 2), '#', 1) as query,
			case when position('#' in url) > 0 then split_part(url, '#', 2) end as fragment
		from chrome_visits cv
		where cv.url_id is null
		on conflict (hostname, path, scheme, query, fragment) do update
		set raw = excluded.raw
		returning id, raw;

		-- Update chrome_visits with url_id
		update chrome_visits cv
		set url_id = u.id
		from urls u
		where cv.url = u.raw and cv.url_id is null;

		-- Make sure all rows have url_id populated
		do $$
		begin
			if exists (select 1 from chrome_visits where url_id is null) then
				raise exception 'Some chrome_visits rows still have null url_id';
			end if;
		end $$;

		-- Make url_id not null
		alter table chrome_visits
		alter column url_id set not null;

		-- Drop old url column
		alter table chrome_visits
		drop column url;
		`,
	},
	{
		"Create ignored_hostnames table",
		`create table if not exists ignored_hostnames (
			id serial primary key,
			hostname text not null,
			reason text,
			created_at timestamp with time zone default current_timestamp,
			constraint ignored_hostnames_hostname_unique unique (hostname)
		)`,
	},
	{
		"Convert to many-to-many relationship between urls and chrome_visits",
		`
		-- Create the junction table
		create table if not exists url_visits (
			url_id uuid references urls(id),
			visit_id uuid references chrome_visits(id),
			created_at timestamp with time zone default current_timestamp,
			primary key (url_id, visit_id)
		);

		-- Migrate existing relationships
		insert into url_visits (url_id, visit_id)
		select url_id, id from chrome_visits;

		-- Keep the url_id column in chrome_visits for the primary URL
		`,
	},
}

func Migrate(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := Pool.QueryRow(ctx, migrations[0][1]).Scan(); err != nil {
		// Ignore error as it's expected when creating table
		core.Logger.Debug("Created schema_migrations table")
	}

	// Get the current migration version
	var currentVersion int
	err := Pool.QueryRow(ctx, "select coalesce(max(version), 0) from schema_migrations").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("error getting current migration version: %w", err)
	}

	// Apply any new migrations
	for version, migration := range migrations[1:] {
		migrationVersion := version + 1
		if migrationVersion <= currentVersion {
			continue
		}

		core.Logger.Info("Applying migration", "version", migrationVersion, "name", migration[0])

		// Start transaction for each migration
		tx, err := Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("error starting transaction: %w", err)
		}

		// Execute migration
		if _, err := tx.Exec(ctx, migration[1]); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("error executing migration %d (%s): %w", migrationVersion, migration[0], err)
		}

		// Record migration
		if _, err := tx.Exec(ctx, "insert into schema_migrations (version) values ($1)", migrationVersion); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("error recording migration %d (%s): %w", migrationVersion, migration[0], err)
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("error committing migration %d (%s): %w", migrationVersion, migration[0], err)
		}

		core.Logger.Info("Successfully applied migration", "version", migrationVersion, "name", migration[0])
	}
	return nil
}
