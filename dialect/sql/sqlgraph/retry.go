// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package sqlgraph

import (
	"context"

	"entgo.io/ent/dialect"
)

// RetryExecutor is an interface for database operations with automatic retries.
type RetryExecutor interface {
	// Do executes the given function within a retry loop without a transaction.
	// The function receives a dialect.Driver that wraps the connection.
	// opts are driver-specific retry options (e.g., ydb retry.Option).
	Do(
		ctx context.Context,
		fn func(ctx context.Context, drv dialect.Driver) error,
		opts ...any,
	) error

	// DoTx executes the given function within a retry loop with a transaction.
	// The function receives a dialect.Driver that wraps the database/sql.Tx transaction.
	// opts are driver-specific retry options (e.g., ydb retry.Option).
	DoTx(
		ctx context.Context,
		fn func(ctx context.Context, drv dialect.Driver) error,
		opts ...any,
	) error
}

// RetryExecutorGetter is an optional interface that drivers can implement to provide
// a RetryExecutor for automatic retry handling.
// If a driver implements this interface,
// sqlgraph operations will use the RetryExecutor for database operations.
type RetryExecutorGetter interface {
	// RetryExecutor returns the RetryExecutor for this driver.
	// If nil is returned, no retry handling will be applied.
	RetryExecutor() RetryExecutor
}

// getRetryExecutor returns the RetryExecutor for the given driver if available.
func getRetryExecutor(drv dialect.Driver) RetryExecutor {
	if reg, ok := drv.(RetryExecutorGetter); ok {
		return reg.RetryExecutor()
	}
	return nil
}

// RetryConfig holds retry configuration for sqlgraph operations.
// This is used to pass retry options to the RetryExecutor.
type RetryConfig struct {
	// Options are driver-specific retry options (e.g., ydb retry.Option).
	// For YDB, these should be retry.Option values from ydb-go-sdk.
	Options []any
}
