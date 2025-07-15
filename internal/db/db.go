package db

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func New(
	url string,
	maxOpenConns int,
	maxIdleConns int,
	maxConnLifetime time.Duration,
	maxConnIdleTime time.Duration,
) (*bun.DB, error) {
	dbConfig, err := pgx.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	// With pgx, you can disable implicit prepared statements, because Bun does
	// not benefit from using them
	// https://bun.uptrace.dev/postgres/#pgx
	dbConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	sqldb := stdlib.OpenDB(*dbConfig)
	sqldb.SetMaxOpenConns(maxOpenConns)
	sqldb.SetMaxIdleConns(maxIdleConns)
	sqldb.SetConnMaxLifetime(maxConnLifetime)
	sqldb.SetConnMaxIdleTime(maxConnIdleTime)

	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(QueryHook{})

	return db, nil
}
