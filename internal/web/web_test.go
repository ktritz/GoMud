package web

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func TestSameOriginWebSocketRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "http://game.example/ws", nil)
	req.Host = "game.example:8080"
	req.Header.Set("Origin", "https://game.example")

	if !sameOriginWebSocketRequest(req) {
		t.Fatal("sameOriginWebSocketRequest() = false, want true for same host")
	}
}

func TestSameOriginWebSocketRequestRejectsCrossHostOrigin(t *testing.T) {
	req := httptest.NewRequest("GET", "http://game.example/ws", nil)
	req.Host = "game.example:8080"
	req.Header.Set("Origin", "https://evil.example")

	if sameOriginWebSocketRequest(req) {
		t.Fatal("sameOriginWebSocketRequest() = true, want false for cross-host origin")
	}
}

func TestServeTemplateReturnsOnParseError(t *testing.T) {
	mudlog.SetupLogger(nil, "debug", "", false)

	tempRoot := t.TempDir()
	badTemplatePath := filepath.Join(tempRoot, "broken.html")
	if err := os.WriteFile(badTemplatePath, []byte("{{ if }}"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	oldRoot := httpRoot
	oldPlugins := webPlugins
	httpRoot = tempRoot
	webPlugins = nil
	t.Cleanup(func() {
		httpRoot = oldRoot
		webPlugins = oldPlugins
	})

	req := httptest.NewRequest("GET", "http://example/broken", nil)
	rec := httptest.NewRecorder()

	serveTemplate(rec, req)

	if rec.Code != 500 {
		t.Fatalf("serveTemplate() status = %d, want 500", rec.Code)
	}
}
