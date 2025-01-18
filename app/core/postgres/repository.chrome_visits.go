package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

type ChromeVisitRepository struct {
	BaseRepository[ChromeVisitModel]
}

var ChromeVisitRepo = &ChromeVisitRepository{
	BaseRepository: BaseRepository[ChromeVisitModel]{
		ModelConfig: ModelConfig[ChromeVisitModel]{
			Table: "chrome_visits",
			Cols:  []string{"id", "url_id", "opened_at", "success", "error_msg", "title", "html", "reason", "created_at"},
			BuildMap: func(model *ChromeVisitModel) map[string]interface{} {
				return map[string]interface{}{
					"url_id":     model.UrlId,
					"opened_at":  model.OpenedAt,
					"success":    model.Success,
					"error_msg":  model.ErrorMsg,
					"title":      model.Title,
					"html":       model.Html,
					"reason":     model.Reason,
					"created_at": model.CreatedAt,
				}
			},
			ScanMap: func(model *ChromeVisitModel) ([]string, []interface{}) {
				return []string{"id", "url_id", "opened_at", "success", "error_msg", "title", "html", "reason", "created_at"},
					[]interface{}{
						&model.Id,
						&model.UrlId,
						&model.OpenedAt,
						&model.Success,
						&model.ErrorMsg,
						&model.Title,
						&model.Html,
						&model.Reason,
						&model.CreatedAt,
					}
			},
		},
	},
}

func (r *ChromeVisitRepository) GetById(ctx context.Context, id string) (*ChromeVisitModel, error) {
	return r.Select(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where("id = ?", id)
	})
}

func (r *ChromeVisitRepository) GetByUrlId(ctx context.Context, urlId string) ([]ChromeVisitModel, error) {
	return r.SelectMultiple(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where("url_id = ?", urlId)
	})
}
