package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestLicensesActivateFetchesRegCodeBeforePost(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPaths []string
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getLicenseRegCode":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"kkty": "reg-code-value"},
			})
		case "/api/v2/activateLicense":
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{"license": map[string]interface{}{"id": "license-1", "state": "valid"}},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "license", "activate", "--ddty", "d"), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPaths) != 2 {
		t.Fatalf("paths = %#v", gotPaths)
	}
	if gotPaths[0] != "/api/v2/getLicenseRegCode" || gotPaths[1] != "/api/v2/activateLicense" {
		t.Fatalf("paths = %#v", gotPaths)
	}
	if gotBody["kkty"] != "reg-code-value" || gotBody["ddty"] != "d" {
		t.Fatalf("body = %+v", gotBody)
	}
	if !strings.Contains(out.String(), "license-1") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestLicensesActivateRequiresDDTY(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "license", "activate"), &out, &errOut)
	if err == nil || err.Error() != "ddty is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestLicensesActivateStopsWhenRegCodeLookupFails(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "license", "activate", "--ddty", "d"), &out, &errOut)
	if err == nil {
		t.Fatal("expected error")
	}
	if requests != 1 {
		t.Fatalf("requests = %d", requests)
	}
}

func setUserDirs(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	t.Setenv("APPDATA", filepath.Join(dir, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(dir, "AppData", "Local"))
}

func TestLicensesActivateStopsWhenRegCodeMissing(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v2/getLicenseRegCode":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "00000000",
				"data": map[string]interface{}{},
			})
		case "/api/v2/activateLicense":
			t.Fatal("activate should not be called when kkty is missing")
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL, "license", "activate", "--ddty", "d"), &out, &errOut)
	if err == nil || err.Error() != "kkty is required" {
		t.Fatalf("err = %v", err)
	}
	if len(gotPaths) != 1 || gotPaths[0] != "/api/v2/getLicenseRegCode" {
		t.Fatalf("paths = %#v", gotPaths)
	}
}

func withHost(t *testing.T, host string, args ...string) []string {
	t.Helper()
	t.Setenv("HYPERBDR_HOST", host)
	return args
}
