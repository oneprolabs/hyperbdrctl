package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIRequestGetPassesQueryAndRawJSON(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	var gotPath string
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":     "00000000",
			"trace_id": "trace-1",
			"data":     map[string]interface{}{"ok": true},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"--output", "json",
		"api", "request",
		"--path", "/api/v2/getTargetCloudInfo",
		"--cloud-account-id", "account-1",
		"--query", "fetch_res=region",
		"--query", "render_mode=1",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v2/getTargetCloudInfo" {
		t.Fatalf("path = %q", gotPath)
	}
	for _, want := range []string{"cloud_account_id=account-1", "fetch_res=region", "render_mode=1"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query = %q, missing %q", gotQuery, want)
		}
	}
	if !strings.Contains(out.String(), `"trace_id": "trace-1"`) {
		t.Fatalf("raw json output = %q", out.String())
	}
}

func TestAPIRequestPostFileAndHeader(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	bodyPath := filepath.Join(dir, "postCloudInfo.json")
	if err := os.WriteFile(bodyPath, []byte(`{"resources_options":{"flavors":{},"os_types":{}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	var gotMethod, gotPath, gotHeader, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotHeader = r.Header.Get("X-Encrypt-Enable")
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(b)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"api", "request",
		"--method", "POST",
		"--path", "/api/v3/postCloudInfo",
		"--file", bodyPath,
		"--header", "X-Encrypt-Enable=true",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "POST" || gotPath != "/api/v3/postCloudInfo" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotHeader != "true" {
		t.Fatalf("X-Encrypt-Enable = %q", gotHeader)
	}
	if !strings.Contains(gotBody, `"resources_options":{"flavors":{},"os_types":{}}`) {
		t.Fatalf("body = %q", gotBody)
	}
}

func TestAPIRequestInlineBody(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(b)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"api", "request",
		"--method", "PATCH",
		"--path", "/api/v2/example",
		"--body", `{"name":"inline"}`,
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody != `{"name":"inline"}` {
		t.Fatalf("body = %q", gotBody)
	}
}

func TestAPIRequestUsesSceneFromEnv(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)
	t.Setenv("HYPERBDR_SCENE", "migration")

	var gotScene string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotScene = r.Header.Get("X-SCENE")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "00000000", "data": map[string]interface{}{"ok": true}})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, srv.URL,
		"api", "request",
		"--path", "/api/v2/getTargetCloudInfo",
	), &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	if gotScene != "migration" {
		t.Fatalf("X-SCENE = %q", gotScene)
	}
}

func TestAPIRequestValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "file and body", args: []string{"api", "request", "--path", "/api/v2/example", "--file", "x.json", "--body", "{}"}, want: "file and body are mutually exclusive"},
		{name: "absolute URL", args: []string{"api", "request", "--path", "https://example.invalid/api"}, want: "path must be a host-relative path starting with /"},
		{name: "relative without slash", args: []string{"api", "request", "--path", "api/v2/example"}, want: "path must be a host-relative path starting with /"},
		{name: "invalid method", args: []string{"api", "request", "--method", "TRACE", "--path", "/api/v2/example"}, want: "method must be one of GET, POST, PUT, PATCH, DELETE"},
		{name: "protected header", args: []string{"api", "request", "--path", "/api/v2/example", "--header", "X-Auth-Token=bad"}, want: "header X-Auth-Token is managed by hyperbdrctl and cannot be overridden"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			setUserDirs(t, dir)
			var out, errOut bytes.Buffer
			err := Execute(tt.args, &out, &errOut)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("err = %v, want %q", err, tt.want)
			}
		})
	}
}
