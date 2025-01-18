package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

type IgnoredHostnameRepository struct {
	BaseRepository[IgnoredHostnameModel]
}

var IgnoredHostnameRepo = &IgnoredHostnameRepository{
	BaseRepository: BaseRepository[IgnoredHostnameModel]{
		ModelConfig: ModelConfig[IgnoredHostnameModel]{
			Table: "ignored_hostnames",
			Cols:  []string{"id", "hostname", "reason", "created_at"},
			BuildMap: func(model *IgnoredHostnameModel) map[string]interface{} {
				return map[string]interface{}{
					"hostname":   model.Hostname,
					"reason":     model.Reason,
					"created_at": model.CreatedAt,
				}
			},
			ScanMap: func(model *IgnoredHostnameModel) ([]string, []interface{}) {
				return []string{"id", "hostname", "reason", "created_at"},
					[]interface{}{
						&model.Id,
						&model.Hostname,
						&model.Reason,
						&model.CreatedAt,
					}
			},
		},
	},
}

func (r *IgnoredHostnameRepository) GetById(ctx context.Context, id int) (*IgnoredHostnameModel, error) {
	return r.Select(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where("id = ?", id)
	})
}

func (r *IgnoredHostnameRepository) GetByHostname(ctx context.Context, hostname string) (*IgnoredHostnameModel, error) {
	return r.Select(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where("hostname = ?", hostname)
	})
}

func (r *IgnoredHostnameRepository) ListAll(ctx context.Context) ([]IgnoredHostnameModel, error) {
	return r.SelectMultiple(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.OrderBy("hostname asc")
	})
}
