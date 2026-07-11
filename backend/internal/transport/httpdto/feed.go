package httpdto

import (
	domain "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/feed"
	"time"
)

type BodyRequest struct {
	Body string `json:"body"`
}
type AuthorSummary struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
}
type Post struct {
	ID             string        `json:"id"`
	Author         AuthorSummary `json:"author"`
	Body           string        `json:"body"`
	Media          *ImageMedia   `json:"media"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
	LikeCount      int           `json:"likeCount"`
	CommentCount   int           `json:"commentCount"`
	ViewerHasLiked bool          `json:"viewerHasLiked"`
	ViewerHasSaved bool          `json:"viewerHasSaved"`
	ViewerOwns     bool          `json:"viewerOwns"`
}
type ImageMedia struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	AltText   string    `json:"altText"`
	MimeType  string    `json:"mimeType"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	CreatedAt time.Time `json:"createdAt"`
}
type FeedPage struct {
	Items      []Post `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
}
type Comment struct {
	ID         string        `json:"id"`
	PostID     string        `json:"postId"`
	Author     AuthorSummary `json:"author"`
	Body       string        `json:"body"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
	ViewerOwns bool          `json:"viewerOwns"`
}
type CommentPage struct {
	Items      []Comment `json:"items"`
	NextCursor string    `json:"nextCursor,omitempty"`
}
type ErrorResponse struct {
	Code string `json:"code"`
}

func NewPost(v domain.Post) Post {
	var media *ImageMedia
	if v.Media != nil {
		media = &ImageMedia{ID: v.Media.ID, URL: v.Media.URL, AltText: v.Media.AltText, MimeType: v.Media.MimeType, Width: v.Media.Width, Height: v.Media.Height, CreatedAt: v.Media.CreatedAt}
	}
	return Post{ID: v.ID, Author: AuthorSummary{ID: v.Author.ID, DisplayName: v.Author.DisplayName, AvatarURL: v.Author.AvatarURL}, Body: v.Body, Media: media, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, LikeCount: v.LikeCount, CommentCount: v.CommentCount, ViewerHasLiked: v.ViewerHasLiked, ViewerHasSaved: v.ViewerHasSaved, ViewerOwns: v.ViewerOwns}
}
func NewComment(v domain.Comment) Comment {
	return Comment{ID: v.ID, PostID: v.PostID, Author: AuthorSummary{ID: v.Author.ID, DisplayName: v.Author.DisplayName, AvatarURL: v.Author.AvatarURL}, Body: v.Body, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, ViewerOwns: v.ViewerOwns}
}
