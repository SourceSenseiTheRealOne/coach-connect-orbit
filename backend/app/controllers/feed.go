package controllers

import (
	"errors"
	appauth "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/auth"
	feedapp "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/feed"
	domainfeed "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/feed"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/transport/httpdto"
	"github.com/revel/revel"
	"net/http"
	"strconv"
)

var feedService *feedapp.Service
var feedVerifier appauth.Verifier = appauth.FailClosedVerifier{}

func ConfigureFeed(s *feedapp.Service, v appauth.Verifier) {
	feedService = s
	if v != nil {
		feedVerifier = v
	}
}

type Feed struct{ *revel.Controller }

func (c Feed) subject() (string, revel.Result) {
	if feedService == nil {
		return "", c.error(http.StatusServiceUnavailable, "service_unavailable")
	}
	token, e := appauth.ExtractBearerToken(c.Request.Header.Get("Authorization"))
	if e != nil {
		return "", c.error(http.StatusUnauthorized, "unauthenticated")
	}
	return token, nil
}
func (c Feed) error(status int, code string) revel.Result {
	c.Response.Status = status
	return c.RenderJSON(httpdto.ErrorResponse{Code: code})
}
func (c Feed) body() (string, error) {
	var request httpdto.BodyRequest
	if c.Params == nil {
		return "", errors.New("missing params")
	}
	if err := c.Params.BindJSON(&request); err != nil {
		return "", err
	}
	return request.Body, nil
}
func (c Feed) CreatePost() revel.Result {
	s, x := c.subject()
	if x != nil {
		return x
	}
	b, e := c.body()
	if e != nil {
		return c.error(400, "invalid_request")
	}
	v, e := feedService.CreatePost(c.Request.Context(), s, b)
	if e != nil {
		return c.appError(e)
	}
	c.Response.Status = 201
	return c.RenderJSON(httpdto.NewPost(v))
}
func (c Feed) GetPost(id string) revel.Result {
	s, x := c.subject()
	if x != nil {
		return x
	}
	v, e := feedService.GetPost(c.Request.Context(), s, id)
	if e != nil {
		return c.appError(e)
	}
	return c.RenderJSON(httpdto.NewPost(v))
}
func (c Feed) ListPosts() revel.Result { return c.list(false) }
func (c Feed) ListSaved() revel.Result { return c.list(true) }
func (c Feed) list(saved bool) revel.Result {
	s, x := c.subject()
	if x != nil {
		return x
	}
	limit := 20
	if raw := c.Params.Query.Get("limit"); raw != "" {
		var e error
		limit, e = strconv.Atoi(raw)
		if e != nil {
			return c.error(400, "invalid_request")
		}
	}
	var items []domainfeed.Post
	var next *string
	var e error
	if saved {
		items, next, e = feedService.ListSavedPosts(c.Request.Context(), s, c.Params.Query.Get("cursor"), limit)
	} else {
		items, next, e = feedService.ListFeed(c.Request.Context(), s, c.Params.Query.Get("cursor"), limit)
	}
	if e != nil {
		return c.appError(e)
	}
	out := httpdto.FeedPage{Items: make([]httpdto.Post, 0, len(items))}
	if next != nil {
		out.NextCursor = *next
	}
	for _, v := range items {
		out.Items = append(out.Items, httpdto.NewPost(v))
	}
	return c.RenderJSON(out)
}
func (c Feed) UpdatePost(id string) revel.Result {
	s, x := c.subject()
	if x != nil {
		return x
	}
	b, e := c.body()
	if e != nil {
		return c.error(400, "invalid_request")
	}
	v, e := feedService.UpdatePost(c.Request.Context(), s, id, b)
	if e != nil {
		return c.appError(e)
	}
	return c.RenderJSON(httpdto.NewPost(v))
}
func (c Feed) DeletePost(id string) revel.Result {
	s, x := c.subject()
	if x != nil {
		return x
	}
	if e := feedService.DeletePost(c.Request.Context(), s, id); e != nil {
		return c.appError(e)
	}
	return c.RenderText("")
}
func (c Feed) Like(id string) revel.Result   { return c.relation(id, true, true) }
func (c Feed) Unlike(id string) revel.Result { return c.relation(id, false, true) }
func (c Feed) Save(id string) revel.Result   { return c.relation(id, true, false) }
func (c Feed) Unsave(id string) revel.Result { return c.relation(id, false, false) }
func (c Feed) relation(id string, set, like bool) revel.Result {
	s, x := c.subject()
	if x != nil {
		return x
	}
	var post domainfeed.Post
	var e error
	if like {
		post, e = feedService.SetPostLike(c.Request.Context(), s, id, set)
	} else {
		post, e = feedService.SetSavedPost(c.Request.Context(), s, id, set)
	}
	if e != nil {
		return c.appError(e)
	}
	return c.RenderJSON(httpdto.NewPost(post))
}
func (c Feed) CreateComment(postID string) revel.Result {
	s, x := c.subject()
	if x != nil {
		return x
	}
	b, e := c.body()
	if e != nil {
		return c.error(400, "invalid_request")
	}
	v, e := feedService.CreateComment(c.Request.Context(), s, postID, b)
	if e != nil {
		return c.appError(e)
	}
	c.Response.Status = 201
	return c.RenderJSON(httpdto.NewComment(v))
}
func (c Feed) ListComments(postID string) revel.Result {
	s, x := c.subject()
	if x != nil {
		return x
	}
	limit := 20
	if raw := c.Params.Query.Get("limit"); raw != "" {
		var err error
		limit, err = strconv.Atoi(raw)
		if err != nil {
			return c.error(http.StatusBadRequest, "invalid_request")
		}
	}
	items, next, e := feedService.ListComments(c.Request.Context(), s, postID, c.Params.Query.Get("cursor"), limit)
	if e != nil {
		return c.appError(e)
	}
	out := httpdto.CommentPage{Items: make([]httpdto.Comment, 0, len(items))}
	if next != nil {
		out.NextCursor = *next
	}
	for _, v := range items {
		out.Items = append(out.Items, httpdto.NewComment(v))
	}
	return c.RenderJSON(out)
}
func (c Feed) UpdateComment(id string) revel.Result {
	s, x := c.subject()
	if x != nil {
		return x
	}
	b, e := c.body()
	if e != nil {
		return c.error(400, "invalid_request")
	}
	v, e := feedService.UpdateComment(c.Request.Context(), s, id, b)
	if e != nil {
		return c.appError(e)
	}
	return c.RenderJSON(httpdto.NewComment(v))
}
func (c Feed) DeleteComment(id string) revel.Result {
	s, x := c.subject()
	if x != nil {
		return x
	}
	if e := feedService.DeleteComment(c.Request.Context(), s, id); e != nil {
		return c.appError(e)
	}
	return c.RenderText("")
}
func (c Feed) appError(e error) revel.Result {
	switch {
	case errors.Is(e, feedapp.ErrUnauthenticated):
		return c.error(401, "unauthenticated")
	case errors.Is(e, feedapp.ErrForbidden):
		return c.error(403, "forbidden")
	case errors.Is(e, feedapp.ErrNotFound):
		return c.error(404, "not_found")
	case errors.Is(e, domainfeed.ErrInvalidBody), errors.Is(e, domainfeed.ErrInvalidCursor), errors.Is(e, domainfeed.ErrInvalidID), errors.Is(e, domainfeed.ErrInvalidLimit):
		return c.error(400, "invalid_request")
	default:
		return c.error(503, "service_unavailable")
	}
}
