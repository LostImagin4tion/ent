// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/examples/ydb/ent"
	"entgo.io/ent/examples/ydb/ent/episode"
	"entgo.io/ent/examples/ydb/ent/season"
	"entgo.io/ent/examples/ydb/ent/series"

	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
)

func main() {
	// Open connection to YDB
	client, err := ent.Open("ydb", "grpc://localhost:2136/local")
	if err != nil {
		log.Fatalf("failed opening connection to ydb: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Run the auto migration tool to create tables with debug logging
	err = client.Schema.Create(
		ctx,
		schema.WithDropColumn(true),
		schema.WithDropIndex(true),
		schema.WithApplyHook(func(next schema.Applier) schema.Applier {
			return schema.ApplyFunc(func(ctx context.Context, conn dialect.ExecQuerier, plan *migrate.Plan) error {
				log.Println("=== DDL Commands ===")
				for i, c := range plan.Changes {
					log.Printf("DDL[%d] Comment: %s", i, c.Comment)
					log.Printf("DDL[%d] Cmd: %s", i, c.Cmd)
					log.Printf("DDL[%d] Args: %v", i, c.Args)
				}
				return next.Apply(ctx, conn, plan)
			})
		}),
	)
	if err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	// Clear existing data before running example
	log.Println("Clearing existing data...")
	if _, err := client.Episode.Delete().Exec(ctx); err != nil {
		log.Printf("Warning: failed to clear episodes: %v", err)
	}
	if _, err := client.Season.Delete().Exec(ctx); err != nil {
		log.Printf("Warning: failed to clear seasons: %v", err)
	}
	if _, err := client.Series.Delete().Exec(ctx); err != nil {
		log.Printf("Warning: failed to clear series: %v", err)
	}
	log.Println("Data cleared")

	// Run the example
	if err := Example(ctx, client); err != nil {
		log.Fatal(err)
	}
}

// Example demonstrates CRUD operations with YDB and ent.
func Example(ctx context.Context, client *ent.Client) error {
	// Create a new series with retry options
	breakingBad, err := client.Series.Create().
		SetTitle("Breaking Bad").
		SetInfo("A high school chemistry teacher turned meth producer").
		SetReleaseDate(time.Date(2008, 1, 20, 0, 0, 0, 0, time.UTC)).
		WithRetryOptions(retry.WithIdempotent(true)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed creating series: %w", err)
	}
	log.Printf("Created series: %v", breakingBad)

	// Create seasons for the series
	season1, err := client.Season.Create().
		SetTitle("Season 1").
		SetFirstAired(time.Date(2008, 1, 20, 0, 0, 0, 0, time.UTC)).
		SetLastAired(time.Date(2008, 3, 9, 0, 0, 0, 0, time.UTC)).
		SetSeries(breakingBad).
		WithRetryOptions(retry.WithIdempotent(true)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed creating season: %w", err)
	}
	log.Printf("Created season: %v", season1)

	// Create episodes
	ep1, err := client.Episode.Create().
		SetTitle("Pilot").
		SetAirDate(time.Date(2008, 1, 20, 0, 0, 0, 0, time.UTC)).
		SetSeason(season1).
		WithRetryOptions(retry.WithIdempotent(true)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed creating episode: %w", err)
	}
	log.Printf("Created episode: %v", ep1)

	ep2, err := client.Episode.Create().
		SetTitle("Cat's in the Bag...").
		SetAirDate(time.Date(2008, 1, 27, 0, 0, 0, 0, time.UTC)).
		SetSeason(season1).
		WithRetryOptions(retry.WithIdempotent(true)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed creating episode: %w", err)
	}
	log.Printf("Created episode: %v", ep2)

	// Query series with retry options
	allSeries, err := client.Series.Query().
		WithRetryOptions(retry.WithIdempotent(true)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed querying series: %w", err)
	}
	log.Printf("All series: %v", allSeries)

	// Query with filtering
	bbSeries, err := client.Series.Query().
		Where(series.TitleContains("Breaking")).
		WithRetryOptions(retry.WithIdempotent(true)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed querying series by title: %w", err)
	}
	log.Printf("Found series: %v", bbSeries)

	// Query seasons for a series using edge traversal
	seasons, err := bbSeries.QuerySeasons().
		WithRetryOptions(retry.WithIdempotent(true)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed querying seasons: %w", err)
	}
	log.Printf("Seasons of %s: %v", bbSeries.Title, seasons)

	// Query episodes for a season
	episodes, err := client.Episode.Query().
		Where(episode.HasSeasonWith(season.TitleEQ("Season 1"))).
		WithRetryOptions(retry.WithIdempotent(true)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed querying episodes: %w", err)
	}
	log.Printf("Episodes in Season 1: %v", episodes)

	// Update series info
	_, err = client.Series.Update().
		Where(series.IDEQ(breakingBad.ID)).
		SetInfo("A high school chemistry teacher turned meth producer - one of the greatest TV series ever made").
		WithRetryOptions(retry.WithIdempotent(true)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed updating series: %w", err)
	}
	log.Printf("Updated series info")

	// Delete episode
	_, err = client.Episode.Delete().
		Where(episode.TitleEQ("Cat's in the Bag...")).
		WithRetryOptions(retry.WithIdempotent(true)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed deleting episode: %w", err)
	}
	log.Printf("Deleted episode")

	// Verify deletion
	remaining, err := client.Episode.Query().
		Where(episode.HasSeasonWith(season.TitleEQ("Season 1"))).
		WithRetryOptions(retry.WithIdempotent(true)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed querying remaining episodes: %w", err)
	}
	log.Printf("Remaining episodes: %v", remaining)

	return nil
}
