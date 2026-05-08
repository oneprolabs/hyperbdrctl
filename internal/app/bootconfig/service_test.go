package bootconfig

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"hyperbdr-client/internal/client"
)

type fakeAPI struct {
	getPath  string
	getQuery url.Values
	postPath string
	postBody interface{}
	gets     []func(string, url.Values) (client.APIResponse, error)
	posts    []func(string, interface{}) (client.APIResponse, error)
}

func (f *fakeAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.getPath = path
	f.getQuery = q
	if len(f.gets) > 0 {
		fn := f.gets[0]
		f.gets = f.gets[1:]
		return fn(path, q)
	}
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func (f *fakeAPI) Post(path string, body interface{}) (client.APIResponse, error) {
	f.postPath = path
	f.postBody = body
	if len(f.posts) > 0 {
		fn := f.posts[0]
		f.posts = f.posts[1:]
		return fn(path, body)
	}
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func TestMetadataFileRejectsInvalidShapes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "array", raw: `[{"storage_id":"storage-1"}]`, want: "file must contain single metadata object"},
		{name: "scalar", raw: `"storage-1"`, want: "file must contain metadata object"},
		{name: "batch create wrapper", raw: `{"batch_create":[{"migration_id":"host-1"}]}`, want: "file must contain metadata object, not batch_create wrapper"},
		{name: "batch update wrapper", raw: `{"batch_update":[{"id":"cfg-1","migration_id":"host-1"}]}`, want: "file must contain metadata object, not batch_update wrapper"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "metadata.json")
			if err := os.WriteFile(path, []byte(tc.raw), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := MetadataFile(path)
			if err == nil || err.Error() != tc.want {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestServiceGetResolvesBootConfigIDAndPostsBatchGet(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"boot_config_id": "cfg-1"}}, nil
			},
		},
		posts: []func(string, interface{}) (client.APIResponse, error){
			func(path string, body interface{}) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"boot_configs": []map[string]interface{}{{"id": "cfg-1"}}}}, nil
			},
		},
	}
	service := NewService(api)

	resp, err := service.Get("host-1")
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/api/v2/getHostDetail" || api.getQuery.Get("host_id") != "host-1" {
		t.Fatalf("get path=%q query=%q", api.getPath, api.getQuery.Encode())
	}
	if api.postPath != "/api/v2/batchGetBootConfigs" {
		t.Fatalf("postPath = %q", api.postPath)
	}
	items := api.postBody.(map[string]interface{})["batch_get"].([]map[string]interface{})
	if len(items) != 1 || items[0]["id"] != "cfg-1" {
		t.Fatalf("postBody = %+v", api.postBody)
	}
	if resp.Data == nil {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestServiceApplyCreatesWhenBootConfigMissing(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{}}, nil
			},
		},
	}
	service := NewService(api)

	result, err := service.Apply("host-1", map[string]interface{}{"storage_id": "storage-1"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Operation != "create" || api.postPath != "/api/v2/batchBootConfigs" {
		t.Fatalf("result = %+v postPath=%q", result, api.postPath)
	}
	item := api.postBody.(map[string]interface{})["batch_create"].([]map[string]interface{})[0]
	if item["migration_id"] != "host-1" {
		t.Fatalf("postBody = %+v", api.postBody)
	}
}

func TestServiceApplyUpdatesWhenBootConfigExists(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"boot_config": map[string]interface{}{"id": "cfg-1"},
				}}, nil
			},
		},
	}
	service := NewService(api)

	result, err := service.Apply("host-1", map[string]interface{}{"storage_id": "storage-1"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Operation != "update" || api.postPath != "/api/v2/batchUpdateBootConfigs" {
		t.Fatalf("result = %+v postPath=%q", result, api.postPath)
	}
	item := api.postBody.(map[string]interface{})["batch_update"].([]map[string]interface{})[0]
	if item["id"] != "cfg-1" || item["migration_id"] != "host-1" {
		t.Fatalf("postBody = %+v", api.postBody)
	}
}

func TestServiceBatchGetPreservesHostOrderAndFallbackIDLookup(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"hosts": []interface{}{
						map[string]interface{}{"id": "host-2", "boot_config_id": "cfg-2"},
						map[string]interface{}{"id": "host-1", "boot_config": map[string]interface{}{"id": "cfg-1"}},
					},
				}}, nil
			},
		},
	}
	service := NewService(api)

	_, err := service.BatchGet([]string{"host-1", "host-2"})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/api/v2/getHosts" || api.getQuery.Get("ids") != "host-1,host-2" {
		t.Fatalf("get path=%q query=%q", api.getPath, api.getQuery.Encode())
	}
	items := api.postBody.(map[string]interface{})["batch_get"].([]map[string]interface{})
	if len(items) != 2 || items[0]["id"] != "cfg-1" || items[1]["id"] != "cfg-2" {
		t.Fatalf("postBody = %+v", api.postBody)
	}
}
