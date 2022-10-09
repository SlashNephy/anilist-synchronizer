package main

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

func getGraphQLClient(ctx context.Context, config *config, tokenFile string) (*graphql.Client, error) {
	client, err := getOAuth2Client(ctx, config, tokenFile)
	if err != nil {
		return nil, err
	}

	return graphql.NewClient("https://graphql.anilist.co", client), nil
}

type viewerQuery struct {
	Viewer struct {
		ID   int    `graphql:"id"`
		Name string `graphql:"name"`
	} `graphql:"Viewer"`
}

func fetchViewer(ctx context.Context, client *graphql.Client) (*viewerQuery, error) {
	var q viewerQuery
	if err := client.Query(ctx, &q, nil); err != nil {
		return nil, err
	}

	return &q, nil
}

type libraryQuery struct {
	MediaListCollection struct {
		Lists []struct {
			Entries []libraryEntry `graphql:"entries"`
		} `graphql:"lists"`
		HasNextChunk bool `graphql:"hasNextChunk"`
	} `graphql:"MediaListCollection(userId: $userId, sort: $sort, perChunk: $perChunk, chunk: $chunk, type: ANIME, forceSingleCompletedList: true, status: $status)"`
}

type libraryEntry struct {
	ID       int             `graphql:"id"`
	Progress int             `graphql:"progress"`
	Status   MediaListStatus `graphql:"status"`
	Score    float32         `graphql:"score"`
	Media    struct {
		ID    int `graphql:"id"`
		Title struct {
			Romaji string `graphql:"romaji"`
			Native string `graphql:"native"`
		} `graphql:"title"`
	} `graphql:"media"`
}

type MediaListSort string
type MediaListStatus string

var mediaListStatuses = []MediaListStatus{"COMPLETED", "CURRENT", "DROPPED", "PAUSED", "PLANNING", "REPEATING"}

func fetchLibrary(ctx context.Context, client *graphql.Client, userID, chunk int, status MediaListStatus) (*libraryQuery, error) {
	var q libraryQuery
	v := map[string]any{
		"userId":   userID,
		"sort":     []MediaListSort{"STARTED_ON"},
		"perChunk": 500,
		"chunk":    chunk,
		"status":   status,
	}
	if err := client.Query(ctx, &q, v); err != nil {
		return nil, err
	}

	return &q, nil
}

type createMediaStatusQuery struct {
	SaveMediaListEntry struct {
		ID int `graphql:"id"`
	} `graphql:"SaveMediaListEntry(mediaId: $mediaId, status: $status, progress: $progress, score: $score)"`
}

func createMediaStatus(ctx context.Context, client *graphql.Client, mediaID int, status MediaListStatus, progress int, score float32) error {
	var q createMediaStatusQuery
	v := map[string]any{
		"mediaId":  mediaID,
		"status":   status,
		"progress": progress,
		"score":    score,
	}
	if err := client.Mutate(ctx, &q, v); err != nil {
		return err
	}

	return nil
}

type updateMediaStatusQuery struct {
	UpdateMediaListEntries []struct {
		ID int `graphql:"id"`
	} `graphql:"UpdateMediaListEntries(ids: [$entryID], status: $status, progress: $progress, score: $score)"`
}

func updateMediaStatus(ctx context.Context, client *graphql.Client, entryID int, status MediaListStatus, progress int, score float32) error {
	var q updateMediaStatusQuery
	v := map[string]any{
		"entryID":  entryID,
		"status":   status,
		"progress": progress,
		"score":    score,
	}
	if err := client.Mutate(ctx, &q, v); err != nil {
		return err
	}

	return nil
}

type deleteMediaStatusQuery struct {
	DeleteMediaListEntry struct {
		Deleted bool `graphql:"deleted"`
	} `graphql:"DeleteMediaListEntry(id: $entryId)"`
}

func deleteMediaStatus(ctx context.Context, client *graphql.Client, entryID int) error {
	var q deleteMediaStatusQuery
	v := map[string]any{
		"entryId": entryID,
	}
	if err := client.Mutate(ctx, &q, v); err != nil {
		return err
	}

	return nil
}
