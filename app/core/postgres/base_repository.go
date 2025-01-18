package postgres

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

type Scannable interface {
	Scan(dest ...any) error
}

type ModelConfig[T any] struct {
	Table    string
	Cols     []string
	BuildMap func(model *T) map[string]interface{}
	ScanMap  func(model *T) ([]string, []interface{})
}

type BaseRepository[V any] struct {
	ModelConfig ModelConfig[V]
}

func (r *BaseRepository[V]) Create(ctx context.Context, model *V) error {
	cols, vals := r.ModelConfig.ScanMap(model)

	query, args, err := psql.
		Insert(r.ModelConfig.Table).
		SetMap(r.ModelConfig.BuildMap(model)).
		Suffix("returning " + strings.Join(cols, ", ")).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %v", err)
	}

	err = Pool.QueryRow(ctx, query, args...).Scan(vals...)
	if err != nil {
		return fmt.Errorf("failed to create record: %v", err)
	}

	return nil
}

func (r *BaseRepository[V]) Select(ctx context.Context, buildFn func(builder sq.SelectBuilder) sq.SelectBuilder) (*V, error) {
	var model V
	cols, vals := r.ModelConfig.ScanMap(&model)

	builder := psql.Select(cols...).From(r.ModelConfig.Table)
	builder = buildFn(builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %v", err)
	}

	err = Pool.QueryRow(ctx, query, args...).Scan(vals...)
	if err != nil {
		return nil, fmt.Errorf("failed to select record: %v", err)
	}

	return &model, nil
}

func (r *BaseRepository[V]) SelectMultiple(ctx context.Context, buildFn func(builder sq.SelectBuilder) sq.SelectBuilder) ([]V, error) {
	var models []V
	var model V
	cols, vals := r.ModelConfig.ScanMap(&model)

	builder := psql.Select(cols...).From(r.ModelConfig.Table)
	builder = buildFn(builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %v", err)
	}

	rows, err := Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to select records: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var model V
		_, vals = r.ModelConfig.ScanMap(&model)
		err = rows.Scan(vals...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %v", err)
		}
		models = append(models, model)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return models, nil
}

func (r *BaseRepository[V]) Update(ctx context.Context, model *V, buildFn func(builder sq.UpdateBuilder) sq.UpdateBuilder) error {
	cols, vals := r.ModelConfig.ScanMap(model)

	builder := psql.Update(r.ModelConfig.Table).
		SetMap(r.ModelConfig.BuildMap(model)).
		Suffix("returning " + strings.Join(cols, ", "))
	builder = buildFn(builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %v", err)
	}

	err = Pool.QueryRow(ctx, query, args...).Scan(vals...)
	if err != nil {
		return fmt.Errorf("failed to update record: %v", err)
	}

	return nil
}

func (r *BaseRepository[V]) Delete(ctx context.Context, buildFn func(builder sq.DeleteBuilder) sq.DeleteBuilder) error {
	builder := psql.Delete(r.ModelConfig.Table)
	builder = buildFn(builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %v", err)
	}

	_, err = Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete record: %v", err)
	}

	return nil
}
