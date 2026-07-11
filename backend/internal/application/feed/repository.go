package feed

import (
	"context"
	domain "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/feed"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/profile"
)

type Repository interface {
	CreatePost(context.Context, profile.Profile, string) (domain.Post, error)
	GetPost(context.Context, profile.Profile, string) (domain.Post, error)
	ListFeed(context.Context, profile.Profile, *domain.Cursor, int) ([]domain.Post, error)
	UpdatePost(context.Context, profile.Profile, string, string) (domain.Post, error)
	DeletePost(context.Context, profile.Profile, string) error
	SetPostLike(context.Context, profile.Profile, string, bool) (domain.Post, error)
	SetSavedPost(context.Context, profile.Profile, string, bool) (domain.Post, error)
	ListSavedPosts(context.Context, profile.Profile, *domain.Cursor, int) ([]domain.Post, *domain.Cursor, error)
	CreateComment(context.Context, profile.Profile, string, string) (domain.Comment, error)
	ListComments(context.Context, profile.Profile, string, *domain.Cursor, int) ([]domain.Comment, error)
	UpdateComment(context.Context, profile.Profile, string, string) (domain.Comment, error)
	DeleteComment(context.Context, profile.Profile, string) error
}
