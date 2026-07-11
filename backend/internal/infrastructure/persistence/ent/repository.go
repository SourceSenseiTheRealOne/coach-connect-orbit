package ent

import (
	"context"
	"errors"
	"time"

	entdb "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/ent"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/ent/comment"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/ent/post"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/ent/postlike"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/ent/postmedia"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/ent/profile"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/ent/savedpost"
	app "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/feed"
	"github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/application/identity"
	domain "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/feed"
	profiledomain "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/profile"
	"github.com/google/uuid"
)

type Repository struct{ client *entdb.Client }
type page struct {
	Items      []domain.Post
	NextCursor string
}

type postAggregate struct {
	PostID string `json:"post_id"`
	Count  int    `json:"count"`
}

type postIDRow struct {
	PostID string `json:"post_id"`
}

func NewRepository(c *entdb.Client) *Repository { return &Repository{c} }
func (r *Repository) ProfileIDBySubject(ctx context.Context, s string) (string, error) {
	p, e := r.client.Profile.Query().Where(profile.ClerkSubjectEQ(s)).Only(ctx)
	if entdb.IsNotFound(e) {
		return "", app.ErrUnauthenticated
	}
	if e != nil {
		return "", app.ErrUnavailable
	}
	return p.ID, nil
}
func (r *Repository) FindByClerkSubject(ctx context.Context, s string) (profiledomain.Profile, error) {
	p, e := r.client.Profile.Query().Where(profile.ClerkSubjectEQ(s)).Only(ctx)
	if entdb.IsNotFound(e) {
		return profiledomain.Profile{}, identity.ErrProfileNotFound
	}
	if e != nil {
		return profiledomain.Profile{}, e
	}
	return mapProfile(p)
}
func mapProfile(p *entdb.Profile) (profiledomain.Profile, error) {
	return profiledomain.NewProfile(p.ID, p.ClerkSubject, p.DisplayName, p.AvatarURL, p.CreatedAt, p.UpdatedAt)
}
func (r *Repository) Create(ctx context.Context, p profiledomain.Profile) (profiledomain.Profile, error) {
	q := r.client.Profile.Create().SetID(p.ID()).SetClerkSubject(p.ClerkSubject()).SetDisplayName(p.DisplayName()).SetCreatedAt(p.CreatedAt()).SetUpdatedAt(p.UpdatedAt())
	if v, ok := p.AvatarURL(); ok {
		q.SetAvatarURL(v)
	}
	v, e := q.Save(ctx)
	if entdb.IsConstraintError(e) {
		return profiledomain.Profile{}, identity.ErrProfileConflict
	}
	if e != nil {
		return profiledomain.Profile{}, e
	}
	return mapProfile(v)
}
func (r *Repository) UpdatePublicIdentity(ctx context.Context, p profiledomain.Profile) (profiledomain.Profile, error) {
	q := r.client.Profile.UpdateOneID(p.ID()).SetDisplayName(p.DisplayName()).SetUpdatedAt(p.UpdatedAt())
	if v, ok := p.AvatarURL(); ok {
		q.SetAvatarURL(v)
	} else {
		q.ClearAvatarURL()
	}
	v, e := q.Save(ctx)
	if e != nil {
		return profiledomain.Profile{}, e
	}
	return mapProfile(v)
}
func (r *Repository) createPost(ctx context.Context, p domain.Post) (domain.Post, error) {
	v, e := r.client.Post.Create().SetID(p.ID).SetAuthorID(p.Author.ID).SetBody(p.Body).SetCreatedAt(p.CreatedAt).SetUpdatedAt(p.UpdatedAt).Save(ctx)
	if e != nil {
		return domain.Post{}, app.ErrUnavailable
	}
	return r.item(ctx, v, p.Author.ID)
}
func (r *Repository) getPost(ctx context.Context, viewer, id string) (domain.Post, error) {
	v, e := r.client.Post.Get(ctx, id)
	if entdb.IsNotFound(e) {
		return domain.Post{}, app.ErrNotFound
	}
	if e != nil {
		return domain.Post{}, app.ErrUnavailable
	}
	return r.item(ctx, v, viewer)
}
func (r *Repository) item(ctx context.Context, p *entdb.Post, viewer string) (domain.Post, error) {
	a, e := p.QueryAuthor().Only(ctx)
	if e != nil {
		return domain.Post{}, app.ErrUnavailable
	}
	likes, e := p.QueryLikes().Count(ctx)
	if e != nil {
		return domain.Post{}, app.ErrUnavailable
	}
	comments, e := p.QueryComments().Count(ctx)
	if e != nil {
		return domain.Post{}, app.ErrUnavailable
	}
	liked, e := p.QueryLikes().Where(postlike.ProfileIDEQ(viewer)).Exist(ctx)
	if e != nil {
		return domain.Post{}, app.ErrUnavailable
	}
	saved, e := p.QuerySaves().Where(savedpost.ProfileIDEQ(viewer)).Exist(ctx)
	if e != nil {
		return domain.Post{}, app.ErrUnavailable
	}
	var media *domain.ImageMedia
	mediaRow, e := p.QueryMedia().Order(entdb.Asc(postmedia.FieldCreatedAt)).First(ctx)
	if e != nil && !entdb.IsNotFound(e) {
		return domain.Post{}, app.ErrUnavailable
	}
	if mediaRow != nil {
		media = &domain.ImageMedia{ID: mediaRow.ID, URL: mediaRow.PublicURL, AltText: mediaRow.AltText, MimeType: mediaRow.MimeType, Width: mediaRow.Width, Height: mediaRow.Height, CreatedAt: mediaRow.CreatedAt}
	}
	return domain.Post{ID: p.ID, Author: domain.AuthorSummary{ID: a.ID, DisplayName: a.DisplayName, AvatarURL: a.AvatarURL}, Body: p.Body, Media: media, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt, LikeCount: likes, CommentCount: comments, ViewerHasLiked: liked, ViewerHasSaved: saved, ViewerOwns: p.AuthorID == viewer}, nil
}

func (r *Repository) hydratePosts(ctx context.Context, rows []*entdb.Post, viewer string) ([]domain.Post, error) {
	if len(rows) == 0 {
		return []domain.Post{}, nil
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	var likeRows, commentRows []postAggregate
	if err := r.client.PostLike.Query().Where(postlike.PostIDIn(ids...)).GroupBy(postlike.FieldPostID).Aggregate(entdb.Count()).Scan(ctx, &likeRows); err != nil {
		return nil, app.ErrUnavailable
	}
	if err := r.client.Comment.Query().Where(comment.PostIDIn(ids...)).GroupBy(comment.FieldPostID).Aggregate(entdb.Count()).Scan(ctx, &commentRows); err != nil {
		return nil, app.ErrUnavailable
	}
	var likedRows, savedRows []postIDRow
	if err := r.client.PostLike.Query().Where(postlike.PostIDIn(ids...), postlike.ProfileIDEQ(viewer)).Select(postlike.FieldPostID).Scan(ctx, &likedRows); err != nil {
		return nil, app.ErrUnavailable
	}
	if err := r.client.SavedPost.Query().Where(savedpost.PostIDIn(ids...), savedpost.ProfileIDEQ(viewer)).Select(savedpost.FieldPostID).Scan(ctx, &savedRows); err != nil {
		return nil, app.ErrUnavailable
	}

	likeCounts := make(map[string]int, len(likeRows))
	commentCounts := make(map[string]int, len(commentRows))
	liked := make(map[string]bool, len(likedRows))
	saved := make(map[string]bool, len(savedRows))
	for _, row := range likeRows {
		likeCounts[row.PostID] = row.Count
	}
	for _, row := range commentRows {
		commentCounts[row.PostID] = row.Count
	}
	for _, row := range likedRows {
		liked[row.PostID] = true
	}
	for _, row := range savedRows {
		saved[row.PostID] = true
	}

	items := make([]domain.Post, 0, len(rows))
	for _, row := range rows {
		author, err := row.Edges.AuthorOrErr()
		if err != nil {
			return nil, app.ErrUnavailable
		}
		var media *domain.ImageMedia
		if len(row.Edges.Media) > 0 {
			value := row.Edges.Media[0]
			media = &domain.ImageMedia{ID: value.ID, URL: value.PublicURL, AltText: value.AltText, MimeType: value.MimeType, Width: value.Width, Height: value.Height, CreatedAt: value.CreatedAt}
		}
		items = append(items, domain.Post{
			ID: row.ID, Author: domain.AuthorSummary{ID: author.ID, DisplayName: author.DisplayName, AvatarURL: author.AvatarURL}, Body: row.Body,
			Media: media, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, LikeCount: likeCounts[row.ID], CommentCount: commentCounts[row.ID],
			ViewerHasLiked: liked[row.ID], ViewerHasSaved: saved[row.ID], ViewerOwns: row.AuthorID == viewer,
		})
	}
	return items, nil
}

func (r *Repository) listPosts(ctx context.Context, viewer string, c *domain.Cursor, limit int) (page, error) {
	q := r.client.Post.Query().WithAuthor().WithMedia(func(media *entdb.PostMediaQuery) {
		media.Order(entdb.Asc(postmedia.FieldCreatedAt))
	})
	if c != nil {
		q = q.Where(post.Or(post.CreatedAtLT(c.CreatedAt), post.And(post.CreatedAtEQ(c.CreatedAt), post.IDLT(c.ID))))
	}
	rows, e := q.Order(entdb.Desc(post.FieldCreatedAt), entdb.Desc(post.FieldID)).Limit(limit + 1).All(ctx)
	if e != nil {
		return page{}, app.ErrUnavailable
	}
	more := len(rows) > limit
	if more {
		rows = rows[:limit]
	}
	items, e := r.hydratePosts(ctx, rows, viewer)
	if e != nil {
		return page{}, e
	}
	out := page{Items: items}
	if more && len(rows) > 0 {
		last := rows[len(rows)-1]
		out.NextCursor, _ = domain.EncodeCursor(domain.Cursor{CreatedAt: last.CreatedAt, ID: last.ID})
	}
	return out, nil
}

func (r *Repository) listSavedPosts(ctx context.Context, viewer string, c *domain.Cursor, limit int) ([]domain.Post, *domain.Cursor, error) {
	q := r.client.SavedPost.Query().Where(savedpost.ProfileIDEQ(viewer)).WithPost(func(postQuery *entdb.PostQuery) {
		postQuery.WithAuthor().WithMedia(func(media *entdb.PostMediaQuery) {
			media.Order(entdb.Asc(postmedia.FieldCreatedAt))
		})
	})
	if c != nil {
		q = q.Where(savedpost.Or(
			savedpost.CreatedAtLT(c.CreatedAt),
			savedpost.And(savedpost.CreatedAtEQ(c.CreatedAt), savedpost.IDLT(c.ID)),
		))
	}
	rows, err := q.Order(entdb.Desc(savedpost.FieldCreatedAt), entdb.Desc(savedpost.FieldID)).Limit(limit + 1).All(ctx)
	if err != nil {
		return nil, nil, app.ErrUnavailable
	}
	more := len(rows) > limit
	if more {
		rows = rows[:limit]
	}
	posts := make([]*entdb.Post, 0, len(rows))
	for _, saved := range rows {
		value, err := saved.Edges.PostOrErr()
		if err != nil {
			return nil, nil, app.ErrUnavailable
		}
		posts = append(posts, value)
	}
	items, err := r.hydratePosts(ctx, posts, viewer)
	if err != nil {
		return nil, nil, err
	}
	if !more || len(rows) == 0 {
		return items, nil, nil
	}
	last := rows[len(rows)-1]
	return items, &domain.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}, nil
}

func (r *Repository) updatePost(ctx context.Context, actor, id, body string, now time.Time) (domain.Post, error) {
	p, e := r.client.Post.Get(ctx, id)
	if entdb.IsNotFound(e) {
		return domain.Post{}, app.ErrNotFound
	}
	if e != nil {
		return domain.Post{}, app.ErrUnavailable
	}
	if p.AuthorID != actor {
		return domain.Post{}, app.ErrForbidden
	}
	p, e = p.Update().SetBody(body).SetUpdatedAt(now).Save(ctx)
	if e != nil {
		return domain.Post{}, app.ErrUnavailable
	}
	return r.item(ctx, p, actor)
}
func (r *Repository) deletePost(ctx context.Context, actor, id string) error {
	p, e := r.client.Post.Get(ctx, id)
	if entdb.IsNotFound(e) {
		return app.ErrNotFound
	}
	if e != nil {
		return app.ErrUnavailable
	}
	if p.AuthorID != actor {
		return app.ErrForbidden
	}
	if e = r.client.Post.DeleteOne(p).Exec(ctx); e != nil {
		return app.ErrUnavailable
	}
	return nil
}
func (r *Repository) setLike(ctx context.Context, actor, id string, set bool) error {
	return r.setRelation(ctx, id, actor, set, true)
}
func (r *Repository) setSave(ctx context.Context, actor, id string, set bool) error {
	return r.setRelation(ctx, id, actor, set, false)
}
func (r *Repository) setRelation(ctx context.Context, pid, uid string, set, like bool) error {
	exists, e := r.client.Post.Query().Where(post.IDEQ(pid)).Exist(ctx)
	if e != nil {
		return app.ErrUnavailable
	}
	if !exists {
		return app.ErrNotFound
	}
	if like {
		if set {
			exists, e := r.client.PostLike.Query().Where(postlike.PostIDEQ(pid), postlike.ProfileIDEQ(uid)).Exist(ctx)
			if e != nil {
				return app.ErrUnavailable
			}
			if exists {
				return nil
			}
			e = r.client.PostLike.Create().SetID(uuid.NewString()).SetPostID(pid).SetProfileID(uid).SetCreatedAt(time.Now().UTC()).Exec(ctx)
			if entdb.IsConstraintError(e) {
				return nil
			}
		} else {
			_, e = r.client.PostLike.Delete().Where(postlike.PostIDEQ(pid), postlike.ProfileIDEQ(uid)).Exec(ctx)
		}
	} else {
		if set {
			exists, e := r.client.SavedPost.Query().Where(savedpost.PostIDEQ(pid), savedpost.ProfileIDEQ(uid)).Exist(ctx)
			if e != nil {
				return app.ErrUnavailable
			}
			if exists {
				return nil
			}
			e = r.client.SavedPost.Create().SetID(uuid.NewString()).SetPostID(pid).SetProfileID(uid).SetCreatedAt(time.Now().UTC()).Exec(ctx)
			if entdb.IsConstraintError(e) {
				return nil
			}
		} else {
			_, e = r.client.SavedPost.Delete().Where(savedpost.PostIDEQ(pid), savedpost.ProfileIDEQ(uid)).Exec(ctx)
		}
	}
	if e != nil {
		return app.ErrUnavailable
	}
	return nil
}
func (r *Repository) createComment(ctx context.Context, c domain.Comment) (domain.Comment, error) {
	exists, e := r.client.Post.Query().Where(post.IDEQ(c.PostID)).Exist(ctx)
	if e != nil {
		return domain.Comment{}, app.ErrUnavailable
	}
	if !exists {
		return domain.Comment{}, app.ErrNotFound
	}
	v, e := r.client.Comment.Create().SetID(c.ID).SetPostID(c.PostID).SetAuthorID(c.Author.ID).SetBody(c.Body).SetCreatedAt(c.CreatedAt).SetUpdatedAt(c.UpdatedAt).Save(ctx)
	if e != nil {
		return domain.Comment{}, app.ErrUnavailable
	}
	return r.mapComment(ctx, v, c.Author.ID)
}
func (r *Repository) mapComment(ctx context.Context, c *entdb.Comment, viewer string) (domain.Comment, error) {
	a, e := c.QueryAuthor().Only(ctx)
	if e != nil {
		return domain.Comment{}, app.ErrUnavailable
	}
	return domain.Comment{ID: c.ID, PostID: c.PostID, Author: domain.AuthorSummary{ID: a.ID, DisplayName: a.DisplayName, AvatarURL: a.AvatarURL}, Body: c.Body, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, ViewerOwns: c.AuthorID == viewer}, nil
}
func (r *Repository) listComments(ctx context.Context, viewer, pid string, c *domain.Cursor, limit int) ([]domain.Comment, string, error) {
	exists, e := r.client.Post.Query().Where(post.IDEQ(pid)).Exist(ctx)
	if e != nil {
		return nil, "", app.ErrUnavailable
	}
	if !exists {
		return nil, "", app.ErrNotFound
	}
	q := r.client.Comment.Query().Where(comment.PostIDEQ(pid))
	if c != nil {
		q = q.Where(comment.Or(comment.CreatedAtLT(c.CreatedAt), comment.And(comment.CreatedAtEQ(c.CreatedAt), comment.IDLT(c.ID))))
	}
	rows, e := q.Order(entdb.Desc(comment.FieldCreatedAt), entdb.Desc(comment.FieldID)).Limit(limit + 1).All(ctx)
	if e != nil {
		return nil, "", app.ErrUnavailable
	}
	more := len(rows) > limit
	if more {
		rows = rows[:limit]
	}
	out := make([]domain.Comment, 0, len(rows))
	for _, v := range rows {
		d, e := r.mapComment(ctx, v, viewer)
		if e != nil {
			return nil, "", app.ErrUnavailable
		}
		out = append(out, d)
	}
	next := ""
	if more {
		v := rows[len(rows)-1]
		next, _ = domain.EncodeCursor(domain.Cursor{CreatedAt: v.CreatedAt, ID: v.ID})
	}
	return out, next, nil
}
func (r *Repository) updateComment(ctx context.Context, actor, id, body string, now time.Time) (domain.Comment, error) {
	c, e := r.client.Comment.Get(ctx, id)
	if entdb.IsNotFound(e) {
		return domain.Comment{}, app.ErrNotFound
	}
	if e != nil {
		return domain.Comment{}, app.ErrUnavailable
	}
	if c.AuthorID != actor {
		return domain.Comment{}, app.ErrForbidden
	}
	c, e = c.Update().SetBody(body).SetUpdatedAt(now).Save(ctx)
	if e != nil {
		return domain.Comment{}, app.ErrUnavailable
	}
	return r.mapComment(ctx, c, actor)
}
func (r *Repository) deleteComment(ctx context.Context, actor, id string) error {
	c, e := r.client.Comment.Get(ctx, id)
	if entdb.IsNotFound(e) {
		return app.ErrNotFound
	}
	if e != nil {
		return app.ErrUnavailable
	}
	if c.AuthorID != actor {
		return app.ErrForbidden
	}
	if e = r.client.Comment.DeleteOne(c).Exec(ctx); e != nil {
		return app.ErrUnavailable
	}
	return nil
}

var _ = errors.Is

func (r *Repository) CreatePost(ctx context.Context, a profiledomain.Profile, body string) (domain.Post, error) {
	now := time.Now().UTC()
	return r.createPost(ctx, domain.Post{ID: uuid.NewString(), Author: domain.NewAuthorSummary(a), Body: body, CreatedAt: now, UpdatedAt: now})
}
func (r *Repository) GetPost(ctx context.Context, v profiledomain.Profile, id string) (domain.Post, error) {
	return r.getPost(ctx, v.ID(), id)
}
func (r *Repository) ListFeed(ctx context.Context, v profiledomain.Profile, c *domain.Cursor, limit int) ([]domain.Post, error) {
	p, e := r.listPosts(ctx, v.ID(), c, limit)
	return p.Items, e
}
func (r *Repository) ListSavedPosts(ctx context.Context, v profiledomain.Profile, c *domain.Cursor, limit int) ([]domain.Post, *domain.Cursor, error) {
	return r.listSavedPosts(ctx, v.ID(), c, limit)
}
func (r *Repository) UpdatePost(ctx context.Context, a profiledomain.Profile, id, body string) (domain.Post, error) {
	return r.updatePost(ctx, a.ID(), id, body, time.Now().UTC())
}
func (r *Repository) DeletePost(ctx context.Context, a profiledomain.Profile, id string) error {
	return r.deletePost(ctx, a.ID(), id)
}
func (r *Repository) SetPostLike(ctx context.Context, a profiledomain.Profile, id string, set bool) (domain.Post, error) {
	if e := r.setLike(ctx, a.ID(), id, set); e != nil {
		return domain.Post{}, e
	}
	return r.getPost(ctx, a.ID(), id)
}
func (r *Repository) SetSavedPost(ctx context.Context, a profiledomain.Profile, id string, set bool) (domain.Post, error) {
	if e := r.setSave(ctx, a.ID(), id, set); e != nil {
		return domain.Post{}, e
	}
	return r.getPost(ctx, a.ID(), id)
}
func (r *Repository) CreateComment(ctx context.Context, a profiledomain.Profile, pid, body string) (domain.Comment, error) {
	now := time.Now().UTC()
	return r.createComment(ctx, domain.Comment{ID: uuid.NewString(), PostID: pid, Author: domain.NewAuthorSummary(a), Body: body, CreatedAt: now, UpdatedAt: now})
}
func (r *Repository) ListComments(ctx context.Context, viewer profiledomain.Profile, pid string, c *domain.Cursor, limit int) ([]domain.Comment, error) {
	v, _, e := r.listComments(ctx, viewer.ID(), pid, c, limit)
	return v, e
}
func (r *Repository) UpdateComment(ctx context.Context, a profiledomain.Profile, id, body string) (domain.Comment, error) {
	return r.updateComment(ctx, a.ID(), id, body, time.Now().UTC())
}
func (r *Repository) DeleteComment(ctx context.Context, a profiledomain.Profile, id string) error {
	return r.deleteComment(ctx, a.ID(), id)
}
