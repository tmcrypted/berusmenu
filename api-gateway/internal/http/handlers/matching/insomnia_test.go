package matching

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestInsomniaCollectionContainsMatchingRoutes(t *testing.T) {
	data, err := os.ReadFile("../../../../insomnia/berusmenu-api.json")
	if err != nil {
		t.Fatalf("read insomnia collection: %v", err)
	}

	var collection struct {
		Resources []struct {
			Method string `json:"method"`
			URL    string `json:"url"`
			Type   string `json:"_type"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("parse insomnia collection: %v", err)
	}

	actual := map[string]bool{}
	for _, resource := range collection.Resources {
		if resource.Type != "request" || resource.Method == "" || resource.URL == "" {
			continue
		}
		path := strings.TrimPrefix(resource.URL, "{{ _.base_url }}")
		if i := strings.IndexByte(path, '?'); i >= 0 {
			path = path[:i]
		}
		actual[resource.Method+" "+path] = true
	}

	expected := []string{
		"POST /v1/vacancies",
		"GET /v1/vacancies",
		"GET /v1/vacancies/1",
		"POST /v1/vacancies/1/responses",
		"GET /v1/vacancies/1/responses",
		"POST /v1/responses/1/withdraw",
		"POST /v1/responses/1/accept",
		"POST /v1/assignments/1/cancel",
		"POST /v1/assignments/1/complete",
		"POST /v1/assignments/1/reviews",
		"PATCH /v1/reviews/1",
		"GET /v1/reviews",
		"POST /v1/reviews/1/replies",
		"GET /v1/reviews/1/replies",
		"PATCH /v1/review-replies/1",
		"GET /v1/users/1/rating",
		"GET /v1/moderation/matching/vacancies",
		"POST /v1/moderation/matching/responses/1/accept",
		"POST /v1/moderation/matching/assignments/1/cancel",
		"POST /v1/moderation/matching/assignments/1/complete",
	}

	for _, route := range expected {
		if !actual[route] {
			t.Fatalf("missing matching route in insomnia collection: %s", route)
		}
	}
}
