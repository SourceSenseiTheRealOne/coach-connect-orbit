package controllers

import (
	"testing"

	"github.com/revel/revel"
)

func TestFeedBodyUsesJSONParsedByRevelParamsFilter(t *testing.T) {
	t.Parallel()

	controller := revel.NewControllerEmpty()
	controller.Params = &revel.Params{JSON: []byte(`{"body":"pressing trigger"}`)}
	feed := Feed{Controller: controller}

	body, err := feed.body()
	if err != nil {
		t.Fatalf("body() error = %v", err)
	}
	if body != "pressing trigger" {
		t.Fatalf("body() = %q, want %q", body, "pressing trigger")
	}
}

func TestFeedBodyRejectsMalformedParsedJSON(t *testing.T) {
	t.Parallel()

	controller := revel.NewControllerEmpty()
	controller.Params = &revel.Params{JSON: []byte(`{"body":`)}
	feed := Feed{Controller: controller}

	if _, err := feed.body(); err == nil {
		t.Fatal("body() error = nil, want malformed JSON error")
	}
}
