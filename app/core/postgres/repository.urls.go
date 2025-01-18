package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

type UrlRepository struct {
	BaseRepository[UrlModel]
}

var UrlRepo = &UrlRepository{
	BaseRepository: BaseRepository[UrlModel]{
		ModelConfig: ModelConfig[UrlModel]{
			Table: "urls",
			Cols:  []string{"id", "raw", "flags", "hostname", "path", "scheme", "query", "fragment", "created_at"},
			BuildMap: func(model *UrlModel) map[string]interface{} {
				return map[string]interface{}{
					"raw":        model.Raw,
					"flags":      model.Flags,
					"hostname":   model.Hostname,
					"path":       model.Path,
					"scheme":     model.Scheme,
					"query":      model.Query,
					"fragment":   model.Fragment,
					"created_at": model.CreatedAt,
				}
			},
			ScanMap: func(model *UrlModel) ([]string, []interface{}) {
				return []string{"id", "raw", "flags", "hostname", "path", "scheme", "query", "fragment", "created_at"},
					[]interface{}{
						&model.Id,
						&model.Raw,
						&model.Flags,
						&model.Hostname,
						&model.Path,
						&model.Scheme,
						&model.Query,
						&model.Fragment,
						&model.CreatedAt,
					}
			},
		},
	},
}

func (r *UrlRepository) UpsertByRaw(ctx context.Context, raw string) (*UrlModel, error) {
	model, err := r.GetByRaw(ctx, raw)
	if err == nil {
		return nil, err
	}

	model = &UrlModel{Raw: raw}
	err = model.CalcFromRaw(ctx)
	if err != nil {
		return nil, err
	}

	return model, model.Create(ctx)
}

func (r *UrlRepository) GetById(ctx context.Context, id string) (*UrlModel, error) {
	return r.Select(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where("id = ?", id)
	})
}

func (r *UrlRepository) GetByRaw(ctx context.Context, raw string) (*UrlModel, error) {
	return r.Select(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where("raw = ?", raw)
	})
}

func (r *UrlRepository) GetByHostname(ctx context.Context, hostname string) ([]UrlModel, error) {
	return r.SelectMultiple(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where("hostname = ?", hostname)
	})
}

func (r *UrlRepository) ListAll(ctx context.Context) ([]UrlModel, error) {
	return r.SelectMultiple(ctx, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.OrderBy("created_at desc")
	})
}
