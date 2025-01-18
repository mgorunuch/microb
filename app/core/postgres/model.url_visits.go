package postgres

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type URLVisitModel struct {
	UrlId     string
	VisitId   string
	CreatedAt time.Time
}

func (m *URLVisitModel) Create(ctx context.Context) error {
	return URLVisitRepo.Create(ctx, m)
}

func (m *URLVisitModel) Update(ctx context.Context) error {
	return URLVisitRepo.Update(ctx, m, func(builder sq.UpdateBuilder) sq.UpdateBuilder {
		return builder.Where("url_id = ? and visit_id = ?", m.UrlId, m.VisitId)
	})
}

func (m *URLVisitModel) Delete(ctx context.Context) error {
	return URLVisitRepo.Delete(ctx, func(builder sq.DeleteBuilder) sq.DeleteBuilder {
		return builder.Where("url_id = ? and visit_id = ?", m.UrlId, m.VisitId)
	})
}
