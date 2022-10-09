package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	sourceClient, err := getGraphQLClient(ctx, cfg, "token-source.json")
	if err != nil {
		log.Fatal(err)
	}

	sourceViewer, err := fetchViewer(ctx, sourceClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Source User = %s (%d)\n", sourceViewer.Viewer.Name, sourceViewer.Viewer.ID)

	sourceEntries, err := fetchAllEntries(ctx, sourceClient, sourceViewer.Viewer.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Source User has %d entries\n", len(sourceEntries))

	targetClient, err := getGraphQLClient(ctx, cfg, "token-target.json")
	if err != nil {
		log.Fatal(err)
	}

	targetViewer, err := fetchViewer(ctx, targetClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Target User = %s (%d)\n", targetViewer.Viewer.Name, targetViewer.Viewer.ID)

	targetEntries, err := fetchAllEntries(ctx, targetClient, targetViewer.Viewer.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Target User has %d entries\n", len(targetEntries))

	if err = update(ctx, sourceEntries, targetEntries, targetClient); err != nil {
		log.Fatal(err)
	}
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
					log.Printf("Found diff in status: %s (%s -> %s)\n", se.Media.Title.Native, se.Status, te.Status)
				}
				if se.Progress != te.Progress {
					update = true
					log.Printf("Found diff in progress: %s (%d -> %d)\n", se.Media.Title.Native, se.Progress, te.Progress)
				}
				if se.Score != te.Score {
					update = true
					log.Printf("Found diff in score: %s (%f -> %f)\n", se.Media.Title.Native, se.Score, te.Score)
				}

				if !update {
					continue
				}

				if err := updateMediaStatus(ctx, targetClient, te.ID, se.Status, se.Progress, se.Score); err != nil {
					return err
				}

				log.Printf("Updated: %s\n", se.Media.Title.Native)
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
