package postgres

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type ChromeVisitModel struct {
	Id        string
	UrlId     string
	OpenedAt  time.Time
	Success   bool
	ErrorMsg  string
	Title     string
	Html      string
	Reason    string
	CreatedAt time.Time
}

func (m *ChromeVisitModel) Create(ctx context.Context) error {
	return ChromeVisitRepo.Create(ctx, m)
}

func (m *ChromeVisitModel) Update(ctx context.Context) error {
	return ChromeVisitRepo.Update(ctx, m, func(builder sq.UpdateBuilder) sq.UpdateBuilder {
		return builder.Where("id = ?", m.Id)
	})
}

func (m *ChromeVisitModel) Delete(ctx context.Context) error {
	return ChromeVisitRepo.Delete(ctx, func(builder sq.DeleteBuilder) sq.DeleteBuilder {
		return builder.Where("id = ?", m.Id)
	})
}
