package repository

import (
	"context"
	"database/sql"
	"errors"
)

// Get one item.
func dbGetOne[T any](ctx context.Context, query string, args ...interface{}) (*T, error) {
	row := _db.QueryRowxContext(ctx, query, args...)
	out := new(T)
	err := row.StructScan(out)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		_log.Err(err).Str("query", query).Str("where", "dbGetOne").Send()
	}
	return out, err
}

// Get many items.
//
// T MUST IMPLEMENT R. (STRUCT TO INTERFACE).
func dbGetManyConvert[T any, R any](ctx context.Context, hook func(*T) error, query string, args ...interface{}) ([]R, error) {
	rows, err := _db.QueryxContext(ctx, query, args...)

	theLog := _log.With().Str("where", "dbGetManyConvert").Str("query", query).Logger()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		theLog.Err(err).Send()
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	result := []R{}
	for rows.Next() {
		newType := new(T)
		err = rows.StructScan(newType)
		if err != nil {
			theLog.Err(err).Send()
			return result, err
		}

		if hook != nil {
			if err = hook(newType); err != nil {
				theLog.Err(err).Send()
				return nil, err
			}
		}

		converted, ok := any(newType).(R)
		if !ok {
			err = errors.New("dbGetMany: T not implements R")
			theLog.Err(err).Send()
			return nil, errors.New("dbGetMany: T not implements R")
		}

		result = append(result, converted)
	}

	return result, err
}

func dbGetMany[T any](ctx context.Context, query string, args ...interface{}) ([]*T, error) {
	rows, err := _db.QueryxContext(ctx, query, args...)

	theLog := _log.With().Str("where", "dbGetMany").Str("query", query).Logger()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		theLog.Err(err).Send()
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	result := []*T{}
	for rows.Next() {
		newType := new(T)
		err = rows.StructScan(newType)
		if err != nil {
			theLog.Err(err).Send()
			return result, err
		}

		result = append(result, newType)
	}

	return result, err
}

// Exec.
func dbExec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := _db.ExecContext(ctx, query, args...)
	if err != nil {
		_log.Err(err).Str("where", "dbGetMany").Str("query", query).Send()
	}
	return result, err
}
