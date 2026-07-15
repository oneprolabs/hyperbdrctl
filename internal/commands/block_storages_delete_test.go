package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBlockStoragesHelpShowsDeleteCommand(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	if err := Execute([]string{"cloud-sync-gateway", "--help"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"\nCommands:\n", "delete", "Usage Notes:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q: %q", want, text)
		}
	}
}

func TestBlockStoragesDeleteHelpUsesFourSectionLayout(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	cases := []struct {
		args []string
		want []string
	}{
		{
			args: []string{"cloud-sync-gateway", "delete", "--help"},
			want: []string{
				"Usage:",
				"\nFlags:\n",
				"Usage Notes:",
				"--id string",
				"--ids string",
				"--force",
				"cloud-sync-gateway delete --id <storage_id>",
				"cloud-sync-gateway delete --ids <storage_id_1,storage_id_2>",
			},
		},
		{
			args: []string{"--lang", "zh_cn", "cloud-sync-gateway", "delete", "--help"},
			want: []string{
				"用法:",
				"\n参数:\n",
				"使用说明:",
				"--id string",
				"--ids string",
				"--force",
				"cloud-sync-gateway delete --id <storage_id>",
				"cloud-sync-gateway delete --ids <storage_id_1,storage_id_2>",
			},
		},
	}

	for _, tt := range cases {
		var out, errOut bytes.Buffer
		if err := Execute(tt.args, &out, &errOut); err != nil {
			t.Fatalf("args=%v err=%v", tt.args, err)
		}

		text := out.String()
		for _, want := range tt.want {
			if !strings.Contains(text, want) {
				t.Fatalf("args=%v help missing %q: %q", tt.args, want, text)
			}
		}
		if strings.Contains(text, "\nCommands:\n") {
			t.Fatalf("leaf help should not include commands section args=%v: %q", tt.args, text)
		}
		for _, unwanted := range []string{"\nExamples:\n", "\nNotes:\n", "\nWorkflow:\n", "\nRelated Commands:\n", "\nNext Steps:\n"} {
			if strings.Contains(text, unwanted) {
				t.Fatalf("leaf help should not include %q args=%v: %q", unwanted, tt.args, text)
			}
		}
		for _, unwanted := range []string{"default false", "默认值 false"} {
			if strings.Contains(text, unwanted) {
				t.Fatalf("force help should not include %q args=%v: %q", unwanted, tt.args, text)
			}
		}
		assertNoHelpFooter(t, text)
	}
}

func TestBlockStoragesDeleteByIDCallsStorageAction(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"deleted": true},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	if err := Execute(withHost(t, srv.URL, "--output", "json", "cloud-sync-gateway", "delete", "--id", "storage-1"), &out, &errOut); err != nil {
		t.Fatal(err)
	}

	if gotPath != "/hypermotion/v1/storages/action" {
		t.Fatalf("path = %q", gotPath)
	}
	deleteStorage := gotBody["delete_storage"].(map[string]interface{})
	storageUUIDs := deleteStorage["storage_uuids"].([]interface{})
	if len(storageUUIDs) != 1 || storageUUIDs[0] != "storage-1" {
		t.Fatalf("storage_uuids = %#v", deleteStorage["storage_uuids"])
	}
	if deleteStorage["force"] != false {
		t.Fatalf("delete_storage = %#v", deleteStorage)
	}
}

func TestBlockStoragesDeleteByIDsCallsStorageAction(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"deleted": true},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	if err := Execute(withHost(t, srv.URL, "cloud-sync-gateway", "delete", "--ids", "storage-1,storage-2", "--force"), &out, &errOut); err != nil {
		t.Fatal(err)
	}

	deleteStorage := gotBody["delete_storage"].(map[string]interface{})
	storageUUIDs := deleteStorage["storage_uuids"].([]interface{})
	want := []string{"storage-1", "storage-2"}
	if len(storageUUIDs) != len(want) {
		t.Fatalf("storage_uuids = %#v", storageUUIDs)
	}
	for i, item := range want {
		if storageUUIDs[i] != item {
			t.Fatalf("storage_uuids = %#v", storageUUIDs)
		}
	}
	if deleteStorage["force"] != true {
		t.Fatalf("delete_storage = %#v", deleteStorage)
	}
}

func TestBlockStoragesDeleteCombinesIDAndIDsInOrder(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "00000000",
			"data": map[string]interface{}{"deleted": true},
		})
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	if err := Execute(withHost(t, srv.URL, "cloud-sync-gateway", "delete", "--id", "storage-1", "--ids", "storage-2, storage-3"), &out, &errOut); err != nil {
		t.Fatal(err)
	}

	deleteStorage := gotBody["delete_storage"].(map[string]interface{})
	storageUUIDs := deleteStorage["storage_uuids"].([]interface{})
	want := []string{"storage-1", "storage-2", "storage-3"}
	if len(storageUUIDs) != len(want) {
		t.Fatalf("storage_uuids = %#v", storageUUIDs)
	}
	for i, item := range want {
		if storageUUIDs[i] != item {
			t.Fatalf("storage_uuids = %#v", storageUUIDs)
		}
	}
}

func TestBlockStoragesDeleteRequiresIDOrIDs(t *testing.T) {
	dir := t.TempDir()
	setUserDirs(t, dir)

	var out, errOut bytes.Buffer
	err := Execute(withHost(t, "https://example.invalid", "cloud-sync-gateway", "delete"), &out, &errOut)
	if err == nil || err.Error() != "id or ids is required" {
		t.Fatalf("err = %v", err)
	}
}
