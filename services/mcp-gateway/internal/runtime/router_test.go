package runtime

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouterHandlesInitialize(t *testing.T) {
	router := NewRouter(NewService(staticCatalog{}, nil))
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"protocolVersion":"2025-06-18"`) {
		t.Fatalf("response missing protocol version: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"tools"`) {
		t.Fatalf("response missing tools capability: %s", rec.Body.String())
	}
}

func TestRouterHandlesInitializedNotification(t *testing.T) {
	router := NewRouter(NewService(staticCatalog{}, nil))
	body := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", rec.Body.String())
	}
}

func TestRouterRejectsGetMCP(t *testing.T) {
	router := NewRouter(NewService(staticCatalog{}, nil))
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestRouterHandlesToolsList(t *testing.T) {
	router := NewRouter(NewService(staticCatalog{
		tools: []Tool{
			{
				Name:        "search/web_search",
				Description: "搜索网页",
				InputSchema: map[string]any{
					"type": "object",
				},
			},
		},
	}, nil))
	body := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"name":"search/web_search"`) {
		t.Fatalf("response missing tool name: %s", rec.Body.String())
	}
}

func TestRouterHandlesToolsCall(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", request.Method)
		}
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(request.Body); err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if !strings.Contains(buf.String(), `"query":"mcp gateway"`) {
			t.Fatalf("request body = %s", buf.String())
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"results":["ok"]}`))
	}))
	defer upstream.Close()

	router := NewRouter(NewService(staticCatalog{
		tools: []Tool{
			{
				Name:        "search/web_search",
				Description: "搜索网页",
				Method:      http.MethodPost,
				URL:         upstream.URL + "/tools/web-search",
			},
		},
	}, upstream.Client()))
	body := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search/web_search","arguments":{"query":"mcp gateway"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"isError":false`) {
		t.Fatalf("response missing success flag: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `\"results\":[\"ok\"]`) {
		t.Fatalf("response missing upstream output: %s", rec.Body.String())
	}
}

func TestRouterReturnsCallErrorWhenToolDoesNotExist(t *testing.T) {
	router := NewRouter(NewService(staticCatalog{}, nil))
	body := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"search/missing","arguments":{"query":"mcp gateway"}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"isError":true`) {
		t.Fatalf("response missing error flag: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `tool not found`) {
		t.Fatalf("response missing not found message: %s", rec.Body.String())
	}
}

type staticCatalog struct {
	tools []Tool
}

func (catalog staticCatalog) ListRuntimeTools() []Tool {
	return catalog.tools
}

func (catalog staticCatalog) FindRuntimeTool(name string) (Tool, bool) {
	for _, tool := range catalog.tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return Tool{}, false
}
