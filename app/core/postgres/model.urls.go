package postgres

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type UrlModel struct {
	Id        string
	Raw       string
	Flags     []string
	Hostname  string
	Path      string
	Scheme    string
	Query     string
	Fragment  string
	CreatedAt time.Time
}

func (m *UrlModel) Create(ctx context.Context) error {
	return UrlRepo.Create(ctx, m)
}

func (m *UrlModel) Update(ctx context.Context) error {
	return UrlRepo.Update(ctx, m, func(builder sq.UpdateBuilder) sq.UpdateBuilder {
		return builder.Where("id = ?", m.Id)
	})
}

func (m *UrlModel) Delete(ctx context.Context) error {
	return UrlRepo.Delete(ctx, func(builder sq.DeleteBuilder) sq.DeleteBuilder {
		return builder.Where("id = ?", m.Id)
	})
}
