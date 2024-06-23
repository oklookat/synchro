package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/oklookat/synchro/shared"
)

func genRepositoryID() shared.RepositoryID {
	return shared.RepositoryID(ulid.Make().String())
}

func genEntityID() shared.RepositoryID {
	return genRepositoryID()
}

func genNewerQuery(tableName string, newerThan time.Time, syncParamName string) string {
	dated := strconv.FormatInt(shared.TimestampNano(newerThan), 10)
	return fmt.Sprintf("SELECT * FROM %s WHERE %s_modified_at > %s", tableName, syncParamName, dated)
}

func genOlderQuery(tableName string, olderThan time.Time, syncParamName string) string {
	dated := strconv.FormatInt(shared.TimestampNano(olderThan), 10)
	return fmt.Sprintf("SELECT * FROM %s WHERE %s_modified_at < %s", tableName, syncParamName, dated)
}

func execSnapshotGetCountQuery(tableName string, snapshotId shared.RepositoryID) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE snapshot_id = ?", tableName)
	result, err := dbGetOneSimple[int](context.Background(), query, snapshotId)
	if err != nil {
		return 0, err
	}
	return *result, err
}

// Get one item.
func dbGetOne[T any](ctx context.Context, query string, args ...interface{}) (*T, error) {
	row := _db.QueryRowxContext(ctx, query, args...)
	out := new(T)
	err := row.StructScan(out)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		slog.Error(err.Error(), "dbGetOne", query)
	}
	return out, err
}

// Get one item into simple value like int.
func dbGetOneSimple[T comparable](ctx context.Context, query string, args ...interface{}) (*T, error) {
	row := _db.QueryRowxContext(ctx, query, args...)
	out := new(T)
	err := row.Scan(out)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		slog.Error(err.Error(), "dbGetOneSimple", query)
	}
	return out, err
}

// Get many items.
//
// T MUST IMPLEMENT R. (STRUCT TO INTERFACE).
func dbGetManyConvert[T any, R any](ctx context.Context, hook func(*T) error, query string, args ...interface{}) ([]R, error) {
	rows, err := _db.QueryxContext(ctx, query, args...)

	defer func() {
		if err != nil {
			slog.Error(err.Error(), "dbGetManyConvert", query)
		}
	}()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
			return nil, nil
		}
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
			return result, err
		}

		if hook != nil {
			if err = hook(newType); err != nil {
				return nil, err
			}
		}

		converted, ok := any(newType).(R)
		if !ok {
			err = errors.New("dbGetMany: T not implements R")
			return nil, err
		}

		result = append(result, converted)
	}

	return result, err
}

func dbGetMany[T any](ctx context.Context, query string, hook func(*T) error, args ...interface{}) ([]*T, error) {
	rows, err := _db.QueryxContext(ctx, query, args...)

	defer func() {
		if err != nil {
			slog.Error(err.Error(), "dbGetMany", query)
		}
	}()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
			return nil, nil
		}
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
			return result, err
		}

		if hook != nil {
			if err = hook(newType); err != nil {
				return nil, err
			}
		}

		result = append(result, newType)
	}

	return result, err
}

// Exec.
func dbExec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := _db.ExecContext(ctx, query, args...)

	if err != nil {
		slog.Error(err.Error(), "dbExec", query)
	}
	return result, err
}
