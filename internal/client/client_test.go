package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"hyperbdr-client/internal/config"
)

func TestHeadersAndLang(t *testing.T) {
	tests := []struct {
		name     string
		cliLang  string
		wantLang string
	}{
		{name: "zh cn maps to api zh", cliLang: "zh_cn", wantLang: "zh"},
		{name: "en remains en", cliLang: "en", wantLang: "en"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotLang, gotScene, gotToken string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotLang = r.Header.Get("X-LANG")
				gotScene = r.Header.Get("X-SCENE")
				gotToken = r.Header.Get("X-Auth-Token")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"items": []interface{}{}}})
			}))
			defer srv.Close()

			dir := t.TempDir()
			tokenPath := filepath.Join(dir, "token.json")
			if err := config.SaveToken(tokenPath, config.Token{Token: "cached-token"}); err != nil {
				t.Fatal(err)
			}
			c, err := New(config.Resolved{Config: config.Config{Host: srv.URL, Scene: "migration", Lang: tt.cliLang}, CachePath: tokenPath})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := c.Get("/api/v2/getHosts", nil); err != nil {
				t.Fatal(err)
			}
			if gotLang != tt.wantLang || gotScene != "migration" || gotToken != "cached-token" {
				t.Fatalf("headers lang=%q scene=%q token=%q", gotLang, gotScene, gotToken)
			}
		})
	}
}

func TestDebugLog(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Auth-Token") == "" {
			t.Fatal("missing auth token")
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":     "00000000",
			"trace_id": "trace-debug",
			"data":     map[string]interface{}{},
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.json")
	if err := config.SaveToken(tokenPath, config.Token{Token: "secret-token"}); err != nil {
		t.Fatal(err)
	}
	var debug bytes.Buffer
	c, err := NewWithDebug(config.Resolved{Config: config.Config{Host: srv.URL, Scene: "dr", Lang: "en", Debug: true}, CachePath: tokenPath}, &debug)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get("/api/v2/getHosts", nil); err != nil {
		t.Fatal(err)
	}
	got := debug.String()
	if !strings.Contains(got, "DEBUG method=GET") || !strings.Contains(got, "status=200") || !strings.Contains(got, "trace_id=trace-debug") {
		t.Fatalf("debug log = %q", got)
	}
	if strings.Contains(got, "secret-token") {
		t.Fatalf("debug log leaked token: %q", got)
	}
}

func TestDebugLogIncludesSanitizedRequestBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":     "00000000",
			"trace_id": "trace-post-debug",
			"data":     map[string]interface{}{},
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.json")
	if err := config.SaveToken(tokenPath, config.Token{Token: "secret-token"}); err != nil {
		t.Fatal(err)
	}
	var debug bytes.Buffer
	c, err := NewWithDebug(config.Resolved{Config: config.Config{Host: srv.URL, Scene: "dr", Lang: "en", Debug: true}, CachePath: tokenPath}, &debug)
	if err != nil {
		t.Fatal(err)
	}
	body := map[string]interface{}{
		"migration_id": "host-1",
		"password":     "top-secret",
		"metadata": map[string]interface{}{
			"user_password": "nested-secret",
			"pool_id":       "pool-1",
		},
	}
	if _, err := c.Post("/api/v2/batchBootConfigs", body); err != nil {
		t.Fatal(err)
	}
	got := debug.String()
	if !strings.Contains(got, `DEBUG method=POST`) || !strings.Contains(got, `trace_id=trace-post-debug`) {
		t.Fatalf("debug log = %q", got)
	}
	if !strings.Contains(got, `body={"metadata":{"pool_id":"pool-1","user_password":"***"},"migration_id":"host-1","password":"***"}`) {
		t.Fatalf("debug log missing sanitized body: %q", got)
	}
	if strings.Contains(got, "top-secret") || strings.Contains(got, "nested-secret") || strings.Contains(got, "secret-token") {
		t.Fatalf("debug log leaked secret: %q", got)
	}
}

func TestAutoLoginWhenTokenMissing(t *testing.T) {
	tests := []struct {
		name   string
		method string
		call   func(*Client) error
	}{
		{name: "get", method: http.MethodGet, call: func(c *Client) error {
			_, err := c.Get("/api/v2/getHosts", nil)
			return err
		}},
		{name: "post", method: http.MethodPost, call: func(c *Client) error {
			_, err := c.Post("/api/v2/batchSync", map[string]interface{}{"ok": true})
			return err
		}},
		{name: "delete", method: http.MethodDelete, call: func(c *Client) error {
			_, err := c.Delete("/hypermotion/v1/hosts", map[string]interface{}{"ids": []string{"host-1"}})
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var loginCalled, requestCalled bool
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v2/loginNoImageCaptcha":
					loginCalled = true
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"token": "fresh-token"}})
				default:
					requestCalled = true
					if r.Method != tt.method {
						t.Fatalf("method = %q, want %q", r.Method, tt.method)
					}
					if got := r.Header.Get("X-Auth-Token"); got != "fresh-token" {
						t.Fatalf("token = %q", got)
					}
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"ok": true}})
				}
			}))
			defer srv.Close()

			dir := t.TempDir()
			c, err := New(config.Resolved{
				Config:    config.Config{Host: srv.URL, Username: "u", Password: "p", Scene: "dr", Lang: "en"},
				CachePath: filepath.Join(dir, "token.json"),
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := tt.call(c); err != nil {
				t.Fatal(err)
			}
			if !loginCalled || !requestCalled {
				t.Fatalf("loginCalled=%v requestCalled=%v", loginCalled, requestCalled)
			}
		})
	}
}

func TestAutoLoginAfterUnauthorized(t *testing.T) {
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/loginNoImageCaptcha":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"token": "fresh-token"}})
		default:
			requests++
			if requests == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if got := r.Header.Get("X-Auth-Token"); got != "fresh-token" {
				t.Fatalf("token = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"ok": true}})
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.json")
	if err := config.SaveToken(tokenPath, config.Token{Token: "stale-token"}); err != nil {
		t.Fatal(err)
	}
	c, err := New(config.Resolved{Config: config.Config{Host: srv.URL, Username: "u", Password: "p", Scene: "dr", Lang: "en"}, CachePath: tokenPath})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get("/api/v2/getHosts", nil); err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d", requests)
	}
}

func TestAcceptsAPICode200AsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "200",
			"detail":  "",
			"message": "ok",
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.json")
	if err := config.SaveToken(tokenPath, config.Token{Token: "cached-token"}); err != nil {
		t.Fatal(err)
	}
	c, err := New(config.Resolved{Config: config.Config{Host: srv.URL, Scene: "dr", Lang: "en"}, CachePath: tokenPath})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Post("/hypermotion/v1/hosts/register", map[string]interface{}{"ids": []string{"vm-1"}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Code != "200" {
		t.Fatalf("code = %q", resp.Code)
	}
}

func TestPathQueryIsPreservedAsRealQuery(t *testing.T) {
	var gotPath string
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"ok": true},
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.json")
	if err := config.SaveToken(tokenPath, config.Token{Token: "cached-token"}); err != nil {
		t.Fatal(err)
	}
	c, err := New(config.Resolved{Config: config.Config{Host: srv.URL, Scene: "dr", Lang: "en"}, CachePath: tokenPath})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Request(http.MethodPost, "/hypermotion/v1/cloud_accounts?rt_msg=1", url.Values{"force": []string{"true"}}, map[string]interface{}{"ok": true}, nil); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/hypermotion/v1/cloud_accounts" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != "force=true&rt_msg=1" && gotQuery != "rt_msg=1&force=true" {
		t.Fatalf("query = %q", gotQuery)
	}
}

func TestErrorMessagePriority(t *testing.T) {
	tests := []struct {
		name   string
		header string
		body   string
		want   string
	}{
		{name: "header", header: "from-header", body: `{"faultstring":"from-body"}`, want: "from-header"},
		{name: "faultstring", body: `{"faultstring":"from-body"}`, want: "from-body"},
		{name: "raw", body: `plain`, want: "plain"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.header != "" {
					w.Header().Set("Server-Error-Message", tt.header)
				}
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			c, err := New(config.Resolved{Config: config.Config{Host: srv.URL, Scene: "dr", Lang: "en"}})
			if err != nil {
				t.Fatal(err)
			}
			_, err = c.Get("/x", nil)
			httpErr, ok := err.(HTTPError)
			if !ok {
				t.Fatalf("err = %T %v", err, err)
			}
			if httpErr.Message != tt.want {
				t.Fatalf("message = %q", httpErr.Message)
			}
		})
	}
}

func TestAPIErrorOnNonSuccessCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":  "10000001",
			"title": "business failed",
			"error": map[string]interface{}{},
		})
	}))
	defer srv.Close()

	c, err := New(config.Resolved{Config: config.Config{Host: srv.URL, Scene: "dr", Lang: "en"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Get("/x", nil)
	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("err = %T %v", err, err)
	}
	if apiErr.Code != "10000001" || apiErr.Message != "business failed" {
		t.Fatalf("api err = %+v", apiErr)
	}
}

func TestDeleteAcceptsScalarSuccessResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %q", r.Method)
		}
		_, _ = w.Write([]byte(`"deleted"`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.json")
	if err := config.SaveToken(tokenPath, config.Token{Token: "cached-token"}); err != nil {
		t.Fatal(err)
	}
	c, err := New(config.Resolved{Config: config.Config{Host: srv.URL, Scene: "dr", Lang: "en"}, CachePath: tokenPath})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Delete("/hypermotion/v1/cloud_accounts/account-1", map[string]interface{}{"id": "account-1"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := resp.Data.(string); !ok || got != "deleted" {
		t.Fatalf("resp = %+v", resp)
	}
}
