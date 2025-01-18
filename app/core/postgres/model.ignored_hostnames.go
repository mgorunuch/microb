package postgres

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type IgnoredHostnameModel struct {
	Id        int
	Hostname  string
	Reason    string
	CreatedAt time.Time
}

func (m *IgnoredHostnameModel) Create(ctx context.Context) error {
	return IgnoredHostnameRepo.Create(ctx, m)
}

func (m *IgnoredHostnameModel) Update(ctx context.Context) error {
	return IgnoredHostnameRepo.Update(ctx, m, func(builder sq.UpdateBuilder) sq.UpdateBuilder {
		return builder.Where("id = ?", m.Id)
	})
}

func (m *IgnoredHostnameModel) Delete(ctx context.Context) error {
	return IgnoredHostnameRepo.Delete(ctx, func(builder sq.DeleteBuilder) sq.DeleteBuilder {
		return builder.Where("id = ?", m.Id)
	})
}
