package host

import (
	"net/url"
	"testing"
	"time"

	"hyperbdr-client/internal/client"
)

type fakeAPI struct {
	getPath    string
	getQuery   url.Values
	postPath   string
	postBody   interface{}
	deletePath string
	deleteBody interface{}
	gets       []func(string, url.Values) (client.APIResponse, error)
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
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func (f *fakeAPI) Delete(path string, body interface{}) (client.APIResponse, error) {
	f.deletePath = path
	f.deleteBody = body
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func TestServiceSyncBuildsBatchSyncBody(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	mode := "full"
	speed := 100
	_, err := service.Sync(SyncSpec{ID: "host-1", Mode: &mode, TransferSpeed: &speed})
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/api/v2/batchSync" {
		t.Fatalf("path = %q", api.postPath)
	}
	item := api.postBody.(map[string]interface{})["batch_sync"].([]map[string]interface{})[0]
	if item["migration_id"] != "host-1" || item["sync_mode"] != "full" || item["transfer_speed"] != 100 {
		t.Fatalf("body = %+v", api.postBody)
	}
}

func TestServiceListBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.List(ListSpec{
		Status:   "running",
		Page:     2,
		PageSize: 20,
		Query:    url.Values{"debug": []string{"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/api/v2/getHosts" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQuery.Get("status") != "running" || api.getQuery.Get("page") != "2" || api.getQuery.Get("page_size") != "20" || api.getQuery.Get("debug") != "1" {
		t.Fatalf("query = %+v", api.getQuery)
	}
}

func TestServiceSnapshotsRequiresID(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.Snapshots(SnapshotsSpec{})
	if err == nil || err.Error() != "id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestServiceBootUsesLatestSnapshotWhenSnapshotIDOmitted(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				if path != "/api/v2/getHostDetail" {
					t.Fatalf("path = %q", path)
				}
				if q.Get("host_id") != "host-1" || q.Get("sheet") != "snapshot" {
					t.Fatalf("query = %+v", q)
				}
				return client.APIResponse{Data: map[string]interface{}{
					"snapshots": []interface{}{
						map[string]interface{}{"id": "snap-old", "created_at": "2026-06-15T10:00:00Z"},
						map[string]interface{}{"id": "snap-new", "created_at": "2026-06-16T10:00:00Z"},
					},
				}}, nil
			},
		},
	}
	service := NewService(api)

	_, err := service.Boot(BootSpec{
		ID:    "host-1",
		Known: map[string]string{"boot_instance_purpose": "drill"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.postPath != "/api/v2/batchBoot" {
		t.Fatalf("path = %q", api.postPath)
	}
	item := api.postBody.(map[string]interface{})["batch_boot"].([]map[string]interface{})[0]
	if item["migration_id"] != "host-1" || item["snapshot_id"] != "snap-new" {
		t.Fatalf("body = %+v", api.postBody)
	}
}

func TestServiceBootKeepsExplicitSnapshotID(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.Boot(BootSpec{
		ID:    "host-1",
		Known: map[string]string{"snapshot_id": "snap-explicit"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(api.gets) != 0 || api.getPath != "" {
		t.Fatalf("boot should not query snapshots when snapshot_id is explicit")
	}
	item := api.postBody.(map[string]interface{})["batch_boot"].([]map[string]interface{})[0]
	if item["snapshot_id"] != "snap-explicit" {
		t.Fatalf("body = %+v", api.postBody)
	}
}

func TestServiceBootFailsWhenHostHasNoSnapshots(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{"snapshots": []interface{}{}}}, nil
			},
		},
	}
	service := NewService(api)

	_, err := service.Boot(BootSpec{ID: "host-1"})
	if err == nil || err.Error() != "host has no snapshots" {
		t.Fatalf("err = %v", err)
	}
}

func TestServiceWaitUsesHostDetailPolling(t *testing.T) {
	api := &fakeAPI{
		gets: []func(string, url.Values) (client.APIResponse, error){
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"status":         "sync_doing",
					"display_status": "sync_doing",
					"task_id":        "task-1",
				}}, nil
			},
			func(path string, q url.Values) (client.APIResponse, error) {
				return client.APIResponse{Data: map[string]interface{}{
					"status":         "sync_snapshot_done",
					"display_status": "sync_snapshot_done",
					"task_id":        "task-1",
				}}, nil
			},
		},
	}
	service := NewService(api)

	result, err := service.Wait(WaitSpec{
		ID:        "host-1",
		Operation: "sync",
		Interval:  0,
		Timeout:   10 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Failed {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Rows) != 1 || result.Rows[0]["result"] != "success" {
		t.Fatalf("rows = %+v", result.Rows)
	}
}
