package main

import (
	"context"
	"sync"
	"time"

	"github.com/hasura/go-graphql-client"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()

	if err := do(ctx); err != nil {
		logger.Error("do", zap.Error(err))
	}
}

func do(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	sourceClient, err := getGraphQLClient(ctx, cfg, "token-source.json")
	if err != nil {
		return err
	}

	sourceViewer, err := fetchViewer(ctx, sourceClient)
	if err != nil {
		return err
	}
	logger.Info("source user", zap.String("name", sourceViewer.Viewer.Name), zap.Int("id", sourceViewer.Viewer.ID))

	targetClient, err := getGraphQLClient(ctx, cfg, "token-target.json")
	if err != nil {
		return err
	}

	targetViewer, err := fetchViewer(ctx, targetClient)
	if err != nil {
		return err
	}
	logger.Info("target user", zap.String("name", targetViewer.Viewer.Name), zap.Int("id", targetViewer.Viewer.ID))

	return doLoop(ctx, cfg, sourceClient, targetClient, sourceViewer.Viewer.ID, targetViewer.Viewer.ID)
}

func doLoop(ctx context.Context, cfg *config, sourceClient, targetClient *graphql.Client, sourceUserID, targetUserID int) error {
	sourceEntries, err := fetchAllEntries(ctx, sourceClient, sourceUserID)
	if err != nil {
		return err
	}
	logger.Info("source entries", zap.Int("len", len(sourceEntries)))

	targetEntries, err := fetchAllEntries(ctx, targetClient, targetUserID)
	if err != nil {
		return err
	}
	logger.Info("target entries", zap.Int("len", len(targetEntries)))

	if err = update(ctx, sourceEntries, targetEntries, targetClient); err != nil {
		return err
	}

	if cfg.IntervalMinutes == 0 {
		return nil
	}

	duration := time.Duration(cfg.IntervalMinutes) * time.Minute
	logger.Info("sleep", zap.String("duration", duration.String()))
	time.Sleep(duration)
	return doLoop(ctx, cfg, sourceClient, targetClient, sourceUserID, targetUserID)
}

// Source と Target の差分を探して更新する
func update(ctx context.Context, source []libraryEntry, target []libraryEntry, targetClient *graphql.Client) error {
	// Source と Target で共通 Entry を探す
	for _, se := range source {
		for _, te := range target {
			if se.Media.ID == te.Media.ID {
				var update = false
				if se.Status != te.Status {
					update = true
					logger.Info("found diff in status", zap.String("title", se.Media.Title.Native), zap.String("old_status", string(se.Status)), zap.String("new_status", string(te.Status)))
				}
				if se.Progress != te.Progress {
					update = true
					logger.Info("found diff in progress", zap.String("title", se.Media.Title.Native), zap.Int("old_progress", se.Progress), zap.Int("new_progress", te.Progress))
				}
				if se.Score != te.Score {
					update = true
					logger.Info("found diff in score", zap.String("title", se.Media.Title.Native), zap.Float32("old_score", se.Score), zap.Float32("new_score", te.Score))
				}

				if !update {
					continue
				}

				if err := updateMediaStatus(ctx, targetClient, te.ID, se.Status, se.Progress, se.Score); err != nil {
					return err
				}

				logger.Info("updated", zap.String("title", se.Media.Title.Native))
				time.Sleep(300 * time.Millisecond)
			}
		}
	}

	return nil
}

func fetchAllEntries(ctx context.Context, client *graphql.Client, userID int) ([]libraryEntry, error) {
	var eg errgroup.Group
	var mutex sync.Mutex
	var entries []libraryEntry

	for _, s := range mediaListStatuses {
		status := s
		eg.Go(func() error {
			var chunk = 0
			for {
				library, err := fetchLibrary(ctx, client, userID, chunk, status)
				if err != nil {
					return err
				}

				mutex.Lock()
				for _, list := range library.MediaListCollection.Lists {
					entries = append(entries, list.Entries...)
				}
				mutex.Unlock()

				if !library.MediaListCollection.HasNextChunk {
					return nil
				}

				chunk++
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return entries, nil
}
