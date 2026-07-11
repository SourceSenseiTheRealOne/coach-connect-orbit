package httpdto

import (
	"encoding/json"
	"testing"

	domain "github.com/SourceSenseiTheRealOne/coach-connect/mvp-v2/backend/internal/domain/feed"
)

func TestFeedDTOUsesExplicitServerDerivedOwnership(t *testing.T) {
	post, err := json.Marshal(NewPost(domain.Post{ViewerOwns: true}))
	if err != nil {
		t.Fatal(err)
	}
	comment, err := json.Marshal(NewComment(domain.Comment{
		Author:     domain.AuthorSummary{ID: "profile-1", DisplayName: "Alex Coach"},
		ViewerOwns: true,
	}))
	if err != nil {
		t.Fatal(err)
	}

	for name, payload := range map[string][]byte{"post": post, "comment": comment} {
		var value map[string]any
		if err := json.Unmarshal(payload, &value); err != nil {
			t.Fatal(err)
		}
		if value["viewerOwns"] != true {
			t.Fatalf("%s viewerOwns = %v, want true", name, value["viewerOwns"])
		}
		if _, leaked := value["viewerCanEdit"]; leaked {
			t.Fatalf("%s leaked legacy viewerCanEdit", name)
		}
	}
}

func TestCommentDTOIncludesPublicAuthorSummary(t *testing.T) {
	comment := NewComment(domain.Comment{Author: domain.AuthorSummary{
		ID:          "profile-1",
		DisplayName: "Alex Coach",
	}})
	if comment.Author.ID != "profile-1" || comment.Author.DisplayName != "Alex Coach" {
		t.Fatalf("author = %#v", comment.Author)
	}
}
