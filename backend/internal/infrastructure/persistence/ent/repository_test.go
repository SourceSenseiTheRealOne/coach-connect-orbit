package ent

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	entdb "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/ent"
	app "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/feed"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/profile"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestRepositorySavedPaginationAndMissingPostComments(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping database: %v", err)
	}
	queryCount := 0
	countQueries := false
	client := entdb.NewClient(
		entdb.Driver(entsql.OpenDB(dialect.Postgres, db)),
		entdb.Debug(),
		entdb.Log(func(...any) {
			if countQueries {
				queryCount++
			}
		}),
	)
	t.Cleanup(func() { _ = client.Close() })

	now := time.Now().UTC().Truncate(time.Microsecond)
	profileID := uuid.NewString()
	clerkSubject := "integration_" + uuid.NewString()
	_, err = client.Profile.Create().
		SetID(profileID).
		SetClerkSubject(clerkSubject).
		SetDisplayName("Integration Coach").
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	t.Cleanup(func() { _ = client.Profile.DeleteOneID(profileID).Exec(context.Background()) })

	viewer, err := profile.NewProfile(profileID, clerkSubject, "Integration Coach", nil, now, now)
	if err != nil {
		t.Fatalf("create domain profile: %v", err)
	}
	newPostID := uuid.NewString()
	oldPostID := uuid.NewString()
	for _, fixture := range []struct {
		id        string
		createdAt time.Time
	}{{newPostID, now.Add(-time.Hour)}, {oldPostID, now.Add(-24 * time.Hour)}} {
		_, err = client.Post.Create().
			SetID(fixture.id).
			SetAuthorID(profileID).
			SetBody("integration post").
			SetCreatedAt(fixture.createdAt).
			SetUpdatedAt(fixture.createdAt).
			Save(ctx)
		if err != nil {
			t.Fatalf("create post: %v", err)
		}
	}

	firstSaveID := uuid.NewString()
	latestSaveID := uuid.NewString()
	_, err = client.SavedPost.Create().SetID(firstSaveID).SetPostID(newPostID).SetProfileID(profileID).SetCreatedAt(now).Save(ctx)
	if err != nil {
		t.Fatalf("create first save: %v", err)
	}
	latestSaveAt := now.Add(time.Minute)
	_, err = client.SavedPost.Create().SetID(latestSaveID).SetPostID(oldPostID).SetProfileID(profileID).SetCreatedAt(latestSaveAt).Save(ctx)
	if err != nil {
		t.Fatalf("create latest save: %v", err)
	}
	_, err = client.PostLike.Create().SetID(uuid.NewString()).SetPostID(newPostID).SetProfileID(profileID).SetCreatedAt(now).Save(ctx)
	if err != nil {
		t.Fatalf("create like: %v", err)
	}
	_, err = client.Comment.Create().SetID(uuid.NewString()).SetPostID(oldPostID).SetAuthorID(profileID).SetBody("integration comment").SetCreatedAt(now).SetUpdatedAt(now).Save(ctx)
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	repository := NewRepository(client)
	countQueries = true
	feed, err := repository.ListFeed(ctx, viewer, nil, 2)
	countQueries = false
	if err != nil {
		t.Fatalf("list feed: %v", err)
	}
	if len(feed) != 2 {
		t.Fatalf("feed length = %d, want 2", len(feed))
	}
	if !feed[0].ViewerHasLiked || !feed[0].ViewerHasSaved || feed[0].LikeCount != 1 {
		t.Fatalf("new post hydration = %#v", feed[0])
	}
	if !feed[1].ViewerHasSaved || feed[1].CommentCount != 1 {
		t.Fatalf("old post hydration = %#v", feed[1])
	}
	if queryCount > 8 {
		t.Fatalf("feed hydration used %d SQL queries, want at most 8", queryCount)
	}
	queryCount = 0
	countQueries = true
	allSaved, _, err := repository.ListSavedPosts(ctx, viewer, nil, 2)
	countQueries = false
	if err != nil || len(allSaved) != 2 {
		t.Fatalf("list all saved: items=%d error=%v", len(allSaved), err)
	}
	if queryCount > 9 {
		t.Fatalf("saved feed hydration used %d SQL queries, want at most 9", queryCount)
	}

	firstPage, cursor, err := repository.ListSavedPosts(ctx, viewer, nil, 1)
	if err != nil {
		t.Fatalf("list first saved page: %v", err)
	}
	if len(firstPage) != 1 || firstPage[0].ID != oldPostID {
		t.Fatalf("first saved page = %#v, want latest-saved older post", firstPage)
	}
	if cursor == nil || cursor.ID != latestSaveID || !cursor.CreatedAt.Equal(latestSaveAt) {
		t.Fatalf("saved cursor = %#v, want latest save record", cursor)
	}

	secondPage, next, err := repository.ListSavedPosts(ctx, viewer, cursor, 1)
	if err != nil {
		t.Fatalf("list second saved page: %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].ID != newPostID || next != nil {
		t.Fatalf("second saved page = %#v, next = %#v", secondPage, next)
	}

	_, err = repository.ListComments(ctx, viewer, uuid.NewString(), nil, 20)
	if !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("missing post comments error = %v, want not found", err)
	}
}
