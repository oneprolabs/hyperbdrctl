package cloudinfo

import "testing"

func TestStorageNetworkRowsFlattensNestedNetworkFields(t *testing.T) {
	rows := StorageNetworkRows(map[string]interface{}{
		"cloud_info": map[string]interface{}{
			"network_addrs_for_write_data": []interface{}{
				map[string]interface{}{"name": "public_endpoint", "value": "public_endpoint"},
			},
			"nested": map[string]interface{}{
				"network_addr_for_read_data": "internal_endpoint",
			},
		},
	})

	if len(rows) != 2 {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0]["type"] != "write" || rows[0]["name"] != "public_endpoint" {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[1]["type"] != "read" || rows[1]["name"] != "internal_endpoint" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestStorageNetworkRowsUsesObjectFallbackName(t *testing.T) {
	rows := StorageNetworkRows(map[string]interface{}{
		"networks": []interface{}{
			map[string]interface{}{"id": "net-1"},
		},
	})

	if len(rows) != 1 || rows[0]["type"] != "network" {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0]["name"] == "" {
		t.Fatalf("rows = %+v", rows)
	}
}
