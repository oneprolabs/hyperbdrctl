package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLicensesList(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{
				"page_size": 1,
				"total":     1,
				"pages": []map[string]interface{}{
					{
						"id":             "license-1",
						"amount":         20,
						"unused":         15,
						"used":           5,
						"status":         "valid",
						"display_status": "valid",
						"start_at":       "2026-01-01T00:00:00Z",
						"expire_at":      "2027-01-01T00:00:00Z",
					},
				},
			},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "license", "list", "--page", "2", "--page-size", "3"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/getLicenses" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != "page=2&page_size=3" {
		t.Fatalf("query = %q", gotQuery)
	}
	if !strings.Contains(out.String(), "license-1") || !strings.Contains(out.String(), "Amount") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestLicensesRegCode(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"kkty": "reg-code-value"},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "license", "reg-code"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/getLicenseRegCode" {
		t.Fatalf("path = %q", gotPath)
	}
	text := out.String()
	for _, want := range []string{"== Registration Code ==", "-------------------------", "reg-code-value"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestLicensesJSONKeepsRawFields(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":     "00000000",
			"data":     map[string]interface{}{"kkty": "raw-value"},
			"trace_id": "trace-1",
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "--output", "json", "license", "reg-code"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"kkty"`) || !strings.Contains(out.String(), `"trace_id"`) {
		t.Fatalf("json output = %q", out.String())
	}
}

func TestLicensesActivateWithFlags(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"license": map[string]interface{}{"id": "license-1", "state": "valid"}},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "license", "activate", "--kkty", "k", "--ddty", "d"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/activateLicense" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody["kkty"] != "k" || gotBody["ddty"] != "d" {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), "license-1") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestLicensesActivateWithFile(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "activate.json")
	if err := os.WriteFile(bodyPath, []byte(`{"kkty":"file-k","ddty":"file-d"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"ok": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "license", "activate", "--file", bodyPath), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["kkty"] != "file-k" || gotBody["ddty"] != "file-d" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func setUserDirs(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	t.Setenv("APPDATA", filepath.Join(dir, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(dir, "AppData", "Local"))
}

func withHost(t *testing.T, host string, args ...string) []string {
	t.Helper()
	t.Setenv("HYPERBDR_HOST", host)
	return args
}
