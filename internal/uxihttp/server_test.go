package uxihttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/uxi"
)

type fakeCompleter struct{}

func (fakeCompleter) Complete(context.Context, string, []uxi.ChatMessage) (string, error) {
	return "local answer", nil
}

func newTestServer(t *testing.T) (*Server, *uxi.Store) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "AIFT-OS", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := uxi.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	engine, err := uxi.NewEngine(store, root, fakeCompleter{})
	if err != nil {
		t.Fatal(err)
	}
	server, err := New(engine)
	if err != nil {
		t.Fatal(err)
	}
	return server, store
}

func TestHealthAndConversation(t *testing.T) {
	server, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("health code = %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/sessions", bytes.NewBufferString(`{"title":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create code = %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFederationTreeEndpoint(t *testing.T) {
	server, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/federation/tree", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("tree code=%d body=%s", rr.Code, rr.Body.String())
	}
	var tree uxi.FederationTree
	if err := json.Unmarshal(rr.Body.Bytes(), &tree); err != nil {
		t.Fatal(err)
	}
	if tree.Schema != "aift.federation.tree.v1" || tree.Progress.Repositories != 1 {
		t.Fatalf("tree=%#v", tree)
	}
	if len(tree.Nodes) != 9 || len(tree.Layers) != 7 {
		t.Fatalf("nodes=%d layers=%d", len(tree.Nodes), len(tree.Layers))
	}
}

func TestFederationWorldEndpoint(t *testing.T) {
	server, _ := newTestServer(t)
	locationDir := filepath.Join(server.Engine.AIFTRoot, "AIFT-OS", ".aift")
	if err := os.MkdirAll(locationDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":"aift.location.v1","label":"Sacramento, California","latitude":38.58157,"longitude":-121.4944,"precision":"city","visibility":"federation"}`
	if err := os.WriteFile(filepath.Join(locationDir, "location.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/federation/world", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("world code=%d body=%s", rr.Code, rr.Body.String())
	}
	var world uxi.FederationWorld
	if err := json.Unmarshal(rr.Body.Bytes(), &world); err != nil {
		t.Fatal(err)
	}
	if world.Schema != "aift.federation.world.v1" || world.Progress.Mapped != 1 || len(world.Nodes) != 1 {
		t.Fatalf("world=%#v", world)
	}
	if world.Nodes[0].Latitude != 38.58 || world.Privacy.ExternalRequests {
		t.Fatalf("world=%#v", world)
	}
}

func TestIndexLaunchesStandaloneGames(t *testing.T) {
	server, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("index code=%d", rr.Code)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		"id=\"treeTab\"", "Tree Game", "id=\"worldTab\"", "World Game",
		"window.location.assign(route)", "treeTab:\"/tree\"", "worldTab:\"/world\"",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("index missing %q", expected)
		}
	}
}

func TestStandaloneGameRoutes(t *testing.T) {
	server, _ := newTestServer(t)
	games := []struct {
		path     string
		title    string
		endpoint string
		peer     string
	}{
		{path: "/tree", title: "Tree of Life", endpoint: "/v1/federation/tree", peer: "/world"},
		{path: "/world", title: "World Game", endpoint: "/v1/federation/world", peer: "/tree"},
	}
	for _, game := range games {
		t.Run(game.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, game.path, nil)
			rr := httptest.NewRecorder()
			server.Handler().ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
			}
			if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
				t.Fatalf("content type=%q", contentType)
			}
			body := rr.Body.String()
			for _, expected := range []string{game.title, game.endpoint, game.peer, "requestFullscreen", "pointerdown"} {
				if !strings.Contains(body, expected) {
					t.Fatalf("game %s missing %q", game.path, expected)
				}
			}
		})
	}
}

func TestGovernanceEndpointsRecordButDoNotExecute(t *testing.T) {
	server, store := newTestServer(t)
	session, err := store.CreateSession("governance")
	if err != nil {
		t.Fatal(err)
	}

	plan := `{"objective":"Review Forge","steps":[{"title":"Inspect mission"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+session.ID+"/plans", bytes.NewBufferString(plan))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("plan code=%d body=%s", rr.Code, rr.Body.String())
	}

	action := `{"kind":"forge.patch.apply","target":"AIFT-Forge","risk":"moderate","approval_required":true}`
	req = httptest.NewRequest(http.MethodPost, "/v1/sessions/"+session.ID+"/actions", bytes.NewBufferString(action))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("action code=%d body=%s", rr.Code, rr.Body.String())
	}
	var updated uxi.Session
	if err := json.Unmarshal(rr.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}

	decision := `{"decision":"approved","actor":"human-local-operator"}`
	req = httptest.NewRequest(http.MethodPost, "/v1/sessions/"+session.ID+"/actions/"+updated.Actions[0].ID+"/decision", bytes.NewBufferString(decision))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("decision code=%d body=%s", rr.Code, rr.Body.String())
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Actions[0].Status != uxi.StatusApproved || len(updated.Jobs) != 0 {
		t.Fatalf("updated=%#v", updated)
	}
}

func TestFederationGeometryEndpoint(t *testing.T) {
	server, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/federation/geometry", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("geometry code=%d body=%s", rr.Code, rr.Body.String())
	}
	firstBody := rr.Body.String()
	req = httptest.NewRequest(http.MethodGet, "/v1/federation/geometry", nil)
	second := httptest.NewRecorder()
	server.Handler().ServeHTTP(second, req)
	if second.Code != http.StatusOK || second.Body.String() != firstBody {
		t.Fatalf("geometry response is not deterministic: first=%s second=%s", firstBody, second.Body.String())
	}
	var geometry uxi.FederationGeometry
	if err := json.Unmarshal([]byte(firstBody), &geometry); err != nil {
		t.Fatal(err)
	}
	if geometry.Schema != "aift.federation.geometry.v1" || geometry.Law.Dimensions != 3 || len(geometry.Nodes) != 1 {
		t.Fatalf("geometry=%#v", geometry)
	}
}
