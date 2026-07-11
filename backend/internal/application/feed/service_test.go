package feed

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationauth "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/auth"
	applicationidentity "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/identity"
	domainfeed "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/feed"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/profile"
)

func TestServiceRejectsUnauthenticatedCreate(t *testing.T) {
	t.Parallel()

	fixture := newTestFeedService(t)
	fixture.auth.err = applicationauth.ErrInvalidBearerToken

	_, err := fixture.service.CreatePost(context.Background(), "bad-token", "match note")
	if !errors.Is(err, domainfeed.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestServiceCreatePostValidatesAndDelegatesServerAuthor(t *testing.T) {
	t.Parallel()

	fixture := newTestFeedService(t)

	post, err := fixture.service.CreatePost(context.Background(), "token", "  pressing trigger  ")
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if post.Author.ID != fixture.author.ID() {
		t.Fatalf("expected reconciled author id %q, got %q", fixture.author.ID(), post.Author.ID)
	}
	if fixture.repository.createdBody != "pressing trigger" {
		t.Fatalf("expected trimmed body, got %q", fixture.repository.createdBody)
	}
}

func TestServiceListFeedUsesOpaqueCursorAndLimitPlusOne(t *testing.T) {
	t.Parallel()

	fixture := newTestFeedService(t)
	createdAt := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	fixture.repository.posts = []domainfeed.Post{
		{ID: "00000000-0000-4000-8000-000000000003", CreatedAt: createdAt},
		{ID: "00000000-0000-4000-8000-000000000002", CreatedAt: createdAt},
		{ID: "00000000-0000-4000-8000-000000000001", CreatedAt: createdAt},
	}

	posts, cursor, err := fixture.service.ListFeed(context.Background(), "token", "", 2)
	if err != nil {
		t.Fatalf("list feed: %v", err)
	}

	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}
	if cursor == nil {
		t.Fatal("expected next cursor")
	}
	decoded, err := domainfeed.DecodeCursor(*cursor)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if decoded == nil || decoded.ID != "00000000-0000-4000-8000-000000000002" {
		t.Fatalf("expected cursor at last visible post, got %#v", decoded)
	}
	if fixture.repository.lastLimit != 3 {
		t.Fatalf("expected limit+1 repository fetch, got %d", fixture.repository.lastLimit)
	}
}

func TestServiceRejectsMalformedResourceID(t *testing.T) {
	t.Parallel()

	fixture := newTestFeedService(t)

	_, err := fixture.service.GetPost(context.Background(), "token", "not-a-uuid")
	if !errors.Is(err, domainfeed.ErrInvalidID) {
		t.Fatalf("expected invalid id, got %v", err)
	}
}

func TestServiceRejectsMalformedCursor(t *testing.T) {
	t.Parallel()

	fixture := newTestFeedService(t)

	_, _, err := fixture.service.ListFeed(context.Background(), "token", "not-base64", 20)
	if !errors.Is(err, domainfeed.ErrInvalidCursor) {
		t.Fatalf("expected invalid cursor, got %v", err)
	}
}

func TestServiceListSavedPostsUsesRepositorySavedAtCursor(t *testing.T) {
	t.Parallel()

	fixture := newTestFeedService(t)
	savedAt := time.Date(2026, 7, 11, 11, 0, 0, 0, time.UTC)
	fixture.repository.posts = []domainfeed.Post{{ID: "older-post"}}
	fixture.repository.nextCursor = &domainfeed.Cursor{CreatedAt: savedAt, ID: "00000000-0000-4000-8000-000000000123"}

	posts, cursor, err := fixture.service.ListSavedPosts(context.Background(), "token", "", 1)
	if err != nil {
		t.Fatalf("list saved posts: %v", err)
	}
	if len(posts) != 1 || posts[0].ID != "older-post" {
		t.Fatalf("unexpected posts: %#v", posts)
	}
	if fixture.repository.lastLimit != 1 {
		t.Fatalf("repository limit = %d, want 1", fixture.repository.lastLimit)
	}
	if cursor == nil {
		t.Fatal("expected saved-at cursor")
	}
	decoded, err := domainfeed.DecodeCursor(*cursor)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if decoded == nil || decoded.ID != "00000000-0000-4000-8000-000000000123" || !decoded.CreatedAt.Equal(savedAt) {
		t.Fatalf("unexpected saved-at cursor: %#v", decoded)
	}
}

type feedServiceFixture struct {
	service    Service
	auth       *fakeVerifier
	identity   *fakeIdentity
	repository *fakeRepository
	author     profile.Profile
}

func newTestFeedService(t *testing.T) feedServiceFixture {
	t.Helper()

	author := mustProfile(t, "profile_123", "user_123", "Ada Coach")
	auth := &fakeVerifier{identity: applicationauth.VerifiedIdentity{Subject: "user_123", DisplayName: "Ada Coach"}}
	identity := &fakeIdentity{author: author}
	repository := &fakeRepository{}
	service, err := NewService(auth, identity, repository)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	return feedServiceFixture{service: service, auth: auth, identity: identity, repository: repository, author: author}
}

type fakeVerifier struct {
	identity applicationauth.VerifiedIdentity
	err      error
}

func (verifier *fakeVerifier) VerifyBearer(_ context.Context, _ string) (applicationauth.VerifiedIdentity, error) {
	if verifier.err != nil {
		return applicationauth.VerifiedIdentity{}, verifier.err
	}
	return verifier.identity, nil
}

type fakeIdentity struct {
	author profile.Profile
}

func (identity *fakeIdentity) Reconcile(_ context.Context, _ applicationidentity.VerifiedIdentity) (profile.Profile, error) {
	return identity.author, nil
}

type fakeRepository struct {
	posts       []domainfeed.Post
	nextCursor  *domainfeed.Cursor
	createdBody string
	lastLimit   int
}

func (repository *fakeRepository) CreatePost(_ context.Context, author profile.Profile, body string) (domainfeed.Post, error) {
	repository.createdBody = body
	return domainfeed.Post{ID: "post_123", Author: domainfeed.NewAuthorSummary(author), Body: body}, nil
}

func (repository *fakeRepository) GetPost(_ context.Context, _ profile.Profile, _ string) (domainfeed.Post, error) {
	return domainfeed.Post{}, nil
}

func (repository *fakeRepository) ListFeed(_ context.Context, _ profile.Profile, _ *domainfeed.Cursor, limit int) ([]domainfeed.Post, error) {
	repository.lastLimit = limit
	return repository.posts, nil
}

func (repository *fakeRepository) UpdatePost(_ context.Context, _ profile.Profile, postID string, body string) (domainfeed.Post, error) {
	return domainfeed.Post{ID: postID, Body: body}, nil
}

func (repository *fakeRepository) DeletePost(_ context.Context, _ profile.Profile, _ string) error {
	return nil
}

func (repository *fakeRepository) SetPostLike(_ context.Context, _ profile.Profile, postID string, liked bool) (domainfeed.Post, error) {
	return domainfeed.Post{ID: postID, ViewerHasLiked: liked}, nil
}

func (repository *fakeRepository) SetSavedPost(_ context.Context, _ profile.Profile, postID string, saved bool) (domainfeed.Post, error) {
	return domainfeed.Post{ID: postID, ViewerHasSaved: saved}, nil
}

func (repository *fakeRepository) ListSavedPosts(_ context.Context, _ profile.Profile, _ *domainfeed.Cursor, limit int) ([]domainfeed.Post, *domainfeed.Cursor, error) {
	repository.lastLimit = limit
	return repository.posts, repository.nextCursor, nil
}

func (repository *fakeRepository) CreateComment(_ context.Context, author profile.Profile, postID string, body string) (domainfeed.Comment, error) {
	return domainfeed.Comment{ID: "comment_123", PostID: postID, Author: domainfeed.NewAuthorSummary(author), Body: body}, nil
}

func (repository *fakeRepository) ListComments(_ context.Context, _ profile.Profile, _ string, _ *domainfeed.Cursor, limit int) ([]domainfeed.Comment, error) {
	repository.lastLimit = limit
	return nil, nil
}

func (repository *fakeRepository) UpdateComment(_ context.Context, _ profile.Profile, commentID string, body string) (domainfeed.Comment, error) {
	return domainfeed.Comment{ID: commentID, Body: body}, nil
}

func (repository *fakeRepository) DeleteComment(_ context.Context, _ profile.Profile, _ string) error {
	return nil
}

func mustProfile(t *testing.T, id string, subject string, name string) profile.Profile {
	t.Helper()
	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	author, err := profile.NewProfile(id, subject, name, nil, now, now)
	if err != nil {
		t.Fatalf("create profile fixture: %v", err)
	}
	return author
}
