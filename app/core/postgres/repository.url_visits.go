package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

type URLVisitRepository struct {
	BaseRepository[URLVisitModel]
}

var URLVisitRepo = &URLVisitRepository{
	BaseRepository: BaseRepository[URLVisitModel]{
		ModelConfig: ModelConfig[URLVisitModel]{
			Table: "url_visits",
			Cols:  []string{"url_id", "visit_id", "created_at"},
			BuildMap: func(model *URLVisitModel) map[string]interface{} {
				return map[string]interface{}{
					"url_id":     model.UrlId,
					"visit_id":   model.VisitId,
					"created_at": model.CreatedAt,
				}
			},
			ScanMap: func(model *URLVisitModel) ([]string, []interface{}) {
				return []string{"url_id", "visit_id", "created_at"},
					[]interface{}{
						&model.UrlId,
						&model.VisitId,
						&model.CreatedAt,
					}
			},
		},
	},
}

func (r *URLVisitRepository) GetByUrlId(ctx context.Context, urlId string) ([]URLVisitModel, error) {
	return r.SelectMultiple(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where("url_id = ?", urlId)
	})
}

func (r *URLVisitRepository) GetByVisitId(ctx context.Context, visitId string) ([]URLVisitModel, error) {
	return r.SelectMultiple(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where("visit_id = ?", visitId)
	})
}

func (r *URLVisitRepository) UpsertRaw(ctx context.Context, urlId string, visitId string) error {
	query := `
		insert into url_visits (url_id, visit_id, created_at)
		values ($1, $2, now())
		on conflict (url_id, visit_id) do nothing
	`
	_, err := Pool.Exec(ctx, query, urlId, visitId)
	return err
}
