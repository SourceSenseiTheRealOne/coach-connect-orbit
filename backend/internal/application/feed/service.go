package feed

import (
	"context"
	"errors"

	applicationauth "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/auth"
	applicationidentity "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/identity"
	domainfeed "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/feed"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/profile"
)

type Service struct {
	auth       applicationauth.Verifier
	identity   IdentityService
	repository Repository
}

type IdentityService interface {
	Reconcile(ctx context.Context, verified applicationidentity.VerifiedIdentity) (profile.Profile, error)
}

func NewService(auth applicationauth.Verifier, identity IdentityService, repository Repository) (Service, error) {
	if auth == nil || identity == nil || repository == nil {
		return Service{}, domainfeed.ErrUnavailable
	}
	return Service{auth: auth, identity: identity, repository: repository}, nil
}

func (service Service) CreatePost(ctx context.Context, bearerToken string, body string) (domainfeed.Post, error) {
	author, err := service.author(ctx, bearerToken)
	if err != nil {
		return domainfeed.Post{}, err
	}

	validBody, err := domainfeed.ValidateBody("create_post", body, domainfeed.MaxPostBodyRunes)
	if err != nil {
		return domainfeed.Post{}, err
	}

	post, err := service.repository.CreatePost(ctx, author, validBody)
	return post, classify(err)
}

func (service Service) GetPost(ctx context.Context, bearerToken string, postID string) (domainfeed.Post, error) {
	viewer, err := service.author(ctx, bearerToken)
	if err != nil {
		return domainfeed.Post{}, err
	}
	postID, err = domainfeed.ValidateID("get_post", postID)
	if err != nil {
		return domainfeed.Post{}, err
	}
	post, err := service.repository.GetPost(ctx, viewer, postID)
	return post, classify(err)
}

func (service Service) ListFeed(ctx context.Context, bearerToken string, cursorValue string, limitValue int) ([]domainfeed.Post, *string, error) {
	viewer, err := service.author(ctx, bearerToken)
	if err != nil {
		return nil, nil, err
	}
	cursor, err := domainfeed.DecodeCursor(cursorValue)
	if err != nil {
		return nil, nil, err
	}
	limit, err := domainfeed.ValidateLimit(limitValue)
	if err != nil {
		return nil, nil, err
	}
	posts, err := service.repository.ListFeed(ctx, viewer, cursor, limit+1)
	if err != nil {
		return nil, nil, classify(err)
	}
	return pagePosts(posts, limit)
}

func (service Service) UpdatePost(ctx context.Context, bearerToken string, postID string, body string) (domainfeed.Post, error) {
	author, err := service.author(ctx, bearerToken)
	if err != nil {
		return domainfeed.Post{}, err
	}
	postID, err = domainfeed.ValidateID("update_post", postID)
	if err != nil {
		return domainfeed.Post{}, err
	}
	validBody, err := domainfeed.ValidateBody("update_post", body, domainfeed.MaxPostBodyRunes)
	if err != nil {
		return domainfeed.Post{}, err
	}
	post, err := service.repository.UpdatePost(ctx, author, postID, validBody)
	return post, classify(err)
}

func (service Service) DeletePost(ctx context.Context, bearerToken string, postID string) error {
	author, err := service.author(ctx, bearerToken)
	if err != nil {
		return err
	}
	postID, err = domainfeed.ValidateID("delete_post", postID)
	if err != nil {
		return err
	}
	return classify(service.repository.DeletePost(ctx, author, postID))
}

func (service Service) SetPostLike(ctx context.Context, bearerToken string, postID string, liked bool) (domainfeed.Post, error) {
	author, err := service.author(ctx, bearerToken)
	if err != nil {
		return domainfeed.Post{}, err
	}
	postID, err = domainfeed.ValidateID("set_post_like", postID)
	if err != nil {
		return domainfeed.Post{}, err
	}
	post, err := service.repository.SetPostLike(ctx, author, postID, liked)
	return post, classify(err)
}

func (service Service) SetSavedPost(ctx context.Context, bearerToken string, postID string, saved bool) (domainfeed.Post, error) {
	author, err := service.author(ctx, bearerToken)
	if err != nil {
		return domainfeed.Post{}, err
	}
	postID, err = domainfeed.ValidateID("set_saved_post", postID)
	if err != nil {
		return domainfeed.Post{}, err
	}
	post, err := service.repository.SetSavedPost(ctx, author, postID, saved)
	return post, classify(err)
}

func (service Service) ListSavedPosts(ctx context.Context, bearerToken string, cursorValue string, limitValue int) ([]domainfeed.Post, *string, error) {
	viewer, err := service.author(ctx, bearerToken)
	if err != nil {
		return nil, nil, err
	}
	cursor, err := domainfeed.DecodeCursor(cursorValue)
	if err != nil {
		return nil, nil, err
	}
	limit, err := domainfeed.ValidateLimit(limitValue)
	if err != nil {
		return nil, nil, err
	}
	posts, next, err := service.repository.ListSavedPosts(ctx, viewer, cursor, limit)
	if err != nil {
		return nil, nil, classify(err)
	}
	if next == nil {
		return posts, nil, nil
	}
	encoded, err := domainfeed.EncodeCursor(*next)
	if err != nil {
		return nil, nil, err
	}
	return posts, &encoded, nil
}

func (service Service) CreateComment(ctx context.Context, bearerToken string, postID string, body string) (domainfeed.Comment, error) {
	author, err := service.author(ctx, bearerToken)
	if err != nil {
		return domainfeed.Comment{}, err
	}
	postID, err = domainfeed.ValidateID("create_comment", postID)
	if err != nil {
		return domainfeed.Comment{}, err
	}
	validBody, err := domainfeed.ValidateBody("create_comment", body, domainfeed.MaxCommentBodyRunes)
	if err != nil {
		return domainfeed.Comment{}, err
	}
	comment, err := service.repository.CreateComment(ctx, author, postID, validBody)
	return comment, classify(err)
}

func (service Service) ListComments(ctx context.Context, bearerToken string, postID string, cursorValue string, limitValue int) ([]domainfeed.Comment, *string, error) {
	viewer, err := service.author(ctx, bearerToken)
	if err != nil {
		return nil, nil, err
	}
	postID, err = domainfeed.ValidateID("list_comments", postID)
	if err != nil {
		return nil, nil, err
	}
	cursor, err := domainfeed.DecodeCursor(cursorValue)
	if err != nil {
		return nil, nil, err
	}
	limit, err := domainfeed.ValidateLimit(limitValue)
	if err != nil {
		return nil, nil, err
	}
	comments, err := service.repository.ListComments(ctx, viewer, postID, cursor, limit+1)
	if err != nil {
		return nil, nil, classify(err)
	}
	return pageComments(comments, limit)
}

func (service Service) UpdateComment(ctx context.Context, bearerToken string, commentID string, body string) (domainfeed.Comment, error) {
	author, err := service.author(ctx, bearerToken)
	if err != nil {
		return domainfeed.Comment{}, err
	}
	commentID, err = domainfeed.ValidateID("update_comment", commentID)
	if err != nil {
		return domainfeed.Comment{}, err
	}
	validBody, err := domainfeed.ValidateBody("update_comment", body, domainfeed.MaxCommentBodyRunes)
	if err != nil {
		return domainfeed.Comment{}, err
	}
	comment, err := service.repository.UpdateComment(ctx, author, commentID, validBody)
	return comment, classify(err)
}

func (service Service) DeleteComment(ctx context.Context, bearerToken string, commentID string) error {
	author, err := service.author(ctx, bearerToken)
	if err != nil {
		return err
	}
	commentID, err = domainfeed.ValidateID("delete_comment", commentID)
	if err != nil {
		return err
	}
	return classify(service.repository.DeleteComment(ctx, author, commentID))
}

func (service Service) author(ctx context.Context, bearerToken string) (profile.Profile, error) {
	verified, err := service.auth.VerifyBearer(ctx, bearerToken)
	if err != nil {
		return profile.Profile{}, domainfeed.ErrUnauthorized
	}

	author, err := service.identity.Reconcile(ctx, verifiedIdentityAdapter{verified: verified})
	if err != nil {
		return profile.Profile{}, domainfeed.ErrUnavailable
	}
	return author, nil
}

func classify(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domainfeed.ErrInvalidCursor) ||
		errors.Is(err, domainfeed.ErrInvalidID) ||
		errors.Is(err, domainfeed.ErrInvalidLimit) ||
		errors.Is(err, domainfeed.ErrInvalidBody) ||
		errors.Is(err, domainfeed.ErrNotFound) ||
		errors.Is(err, domainfeed.ErrForbidden) {
		return err
	}
	return domainfeed.ErrUnavailable
}

func pagePosts(posts []domainfeed.Post, limit int) ([]domainfeed.Post, *string, error) {
	if len(posts) <= limit {
		return posts, nil, nil
	}
	page := posts[:limit]
	last := page[len(page)-1]
	cursor, err := domainfeed.EncodeCursor(domainfeed.Cursor{CreatedAt: last.CreatedAt, ID: last.ID})
	if err != nil {
		return nil, nil, err
	}
	return page, &cursor, nil
}

func pageComments(comments []domainfeed.Comment, limit int) ([]domainfeed.Comment, *string, error) {
	if len(comments) <= limit {
		return comments, nil, nil
	}
	page := comments[:limit]
	last := page[len(page)-1]
	cursor, err := domainfeed.EncodeCursor(domainfeed.Cursor{CreatedAt: last.CreatedAt, ID: last.ID})
	if err != nil {
		return nil, nil, err
	}
	return page, &cursor, nil
}

type verifiedIdentityAdapter struct {
	verified applicationauth.VerifiedIdentity
}

func (adapter verifiedIdentityAdapter) ClerkSubject() string {
	return adapter.verified.Subject
}

func (adapter verifiedIdentityAdapter) PublicDisplayName() string {
	return adapter.verified.DisplayName
}

func (adapter verifiedIdentityAdapter) PublicAvatarURL() *string {
	return adapter.verified.AvatarURL
}
