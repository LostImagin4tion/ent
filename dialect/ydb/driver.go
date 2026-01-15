// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package ydb

import (
	"context"
	"database/sql"

	"entgo.io/ent/dialect"
	entSql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	ydb "github.com/ydb-platform/ydb-go-sdk/v3"
)

// YDBDriver is a [dialect.Driver] implementation for YDB.
type YDBDriver struct {
	*entSql.Driver

	nativeDriver  *ydb.Driver
	retryExecutor *RetryExecutor
}

var _ sqlgraph.RetryExecutorGetter = (*YDBDriver)(nil)

func Open(ctx context.Context, dsn string) (*YDBDriver, error) {
	nativeDriver, err := ydb.Open(ctx, dsn)
	if err != nil {
		return nil, err
	}

	conn, err := ydb.Connector(
		nativeDriver,
		ydb.WithAutoDeclare(),
		ydb.WithTablePathPrefix(nativeDriver.Name()),
	)
	if err != nil {
		return nil, err
	}

	dbSQLDriver := sql.OpenDB(conn)

	return &YDBDriver{
		Driver:        entSql.OpenDB(dialect.YDB, dbSQLDriver),
		nativeDriver:  nativeDriver,
		retryExecutor: NewRetryExecutor(dbSQLDriver),
	}, nil
}

func (y *YDBDriver) NativeDriver() *ydb.Driver {
	return y.nativeDriver
}

// RetryExecutor returns the RetryExecutor for this driver.
// This allows sqlgraph to automatically wrap operations with YDB retry logic.
func (y *YDBDriver) RetryExecutor() sqlgraph.RetryExecutor {
	return y.retryExecutor
}
