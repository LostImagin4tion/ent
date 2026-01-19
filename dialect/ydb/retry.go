// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package ydb

import (
	"context"
	"database/sql"

	"entgo.io/ent/dialect"
	entSql "entgo.io/ent/dialect/sql"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
)

// RetryExecutor implements sqlgraph.RetryExecutor for YDB
type RetryExecutor struct {
	db *sql.DB
}

// NewRetryExecutor creates a new RetryExecutor with the given database connection
func NewRetryExecutor(db *sql.DB) *RetryExecutor {
	return &RetryExecutor{db: db}
}

// Do executes a read-only operation with retry support.
// It uses ydb-go-sdk's retry.Do which handles YDB-specific retryable errors.
// Options should be created using retry.WithIdempotent(), retry.WithLabel(), etc.
func (r *RetryExecutor) Do(
	ctx context.Context,
	fn func(ctx context.Context, drv dialect.Driver) error,
	opts ...any,
) error {
	return retry.Do(
		ctx,
		r.db,
		func(ctx context.Context, conn *sql.Conn) error {
			return fn(ctx, NewRetryDriver(conn))
		},
		retry.WithDoRetryOptions(toRetryOptions(opts)...),
	)
}

// DoTx executes the operation within a transaction with retry support.
// It uses ydb-go-sdk's retry.DoTx which handles YDB-specific retryable errors.
// Options should be created using retry.WithIdempotent(), retry.WithLabel(), etc.
func (r *RetryExecutor) DoTx(
	ctx context.Context,
	fn func(ctx context.Context, drv dialect.Driver) error,
	opts ...any,
) error {
	return retry.DoTx(
		ctx,
		r.db,
		func(ctx context.Context, tx *sql.Tx) error {
			return fn(ctx, NewTxRetryDriver(tx))
		},
		retry.WithDoTxRetryOptions(toRetryOptions(opts)...),
	)
}

// toRetryOptions converts a slice of any options to retry.Option slice
func toRetryOptions(opts []any) []retry.Option {
	retryOpts := make([]retry.Option, 0, len(opts))
	for _, opt := range opts {
		if ro, ok := opt.(retry.Option); ok {
			retryOpts = append(retryOpts, ro)
		}
	}
	return retryOpts
}

// RetryDriver is designed for use only in sqlgraph,
// specifically - in retry.DoTx callbacks
type RetryDriver struct {
	entSql.Conn
}

var _ dialect.Driver = (*RetryDriver)(nil)

// NewTxRetryDriver creates a new RetryDriver from a transaction.
func NewTxRetryDriver(tx *sql.Tx) *RetryDriver {
	return &RetryDriver{
		Conn: entSql.Conn{ExecQuerier: tx},
	}
}

// NewRetryDriver creates a new RetryDriver from a database connection.
func NewRetryDriver(conn *sql.Conn) *RetryDriver {
	return &RetryDriver{
		Conn: entSql.Conn{ExecQuerier: conn},
	}
}

// sqlgraph creates nested transactions in several methods.
// But YDB doesnt support nested transactions.
// Therefore, this methods returns no-op tx
func (d *RetryDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	return dialect.NopTx(d), nil
}

// Close is a no-op for RetryDriver since retry.DoTx manages the transaction lifecycle.
func (d *RetryDriver) Close() error {
	return nil
}

// Dialect returns the YDB dialect name.
func (d *RetryDriver) Dialect() string {
	return dialect.YDB
}
