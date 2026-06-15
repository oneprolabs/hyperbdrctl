package cloudinfo

import "testing"

func TestBuildAuthResourceEnvelopeAutoDetectsBootLoaderFlavors(t *testing.T) {
	envelope := BuildAuthResourceEnvelope("huawei", "huawei_obs", "objectstorage", map[string]interface{}{
		"cloud_info": map[string]interface{}{
			"boot_loader_flavors": []interface{}{
				map[string]interface{}{"id": "boot-flavor-1", "name": "2C4G"},
			},
		},
	}, nil, nil)

	if len(envelope.Sections) != 1 {
		t.Fatalf("sections = %+v", envelope.Sections)
	}
	section := envelope.Sections[0]
	if section.Resource != "boot_loader_flavors" || section.TitleKey != "resource.boot_loader_flavors" {
		t.Fatalf("section = %+v", section)
	}
}

func TestNormalizeHuaweiFlavorRowsPrefersCanonicalFields(t *testing.T) {
	rows := NormalizeHuaweiFlavorRows([]map[string]interface{}{
		{
			"id":           "flavor-1",
			"name":         "flavor-1",
			"vcpus":        2,
			"ram_GB":       4,
			"ram":          8,
			"zone_id":      "cn-north-1a",
			"GHz":          "2.2GHz",
			"quota_rate":   "0.2 / 0.8 Gbit/s",
			"quota_pps":    "100,000 PPS",
			"max_nic_num":  12,
			"max_disk_num": 24,
		},
	})

	if got := rows[0]["ram_gb"]; got != 4 {
		t.Fatalf("ram_gb = %#v", got)
	}
	if got := rows[0]["ghz"]; got != "2.2GHz" {
		t.Fatalf("ghz = %#v", got)
	}
}

func TestNormalizeHuaweiFlavorRowsFlattensNestedTree(t *testing.T) {
	rows := NormalizeHuaweiFlavorRows([]map[string]interface{}{
		{
			"id":   "u-2",
			"name": "2C",
			"children": []interface{}{
				map[string]interface{}{
					"id":   "u-2-m-4",
					"name": "4GB",
					"children": []interface{}{
						map[string]interface{}{
							"id":           "c3.large.2",
							"name":         "c3.large.2",
							"vcpus":        2,
							"ram_GB":       4,
							"zone_id":      "cn-north-1a",
							"quota_rate":   "0.6 / 1.5 Gbit/s",
							"quota_pps":    "300,000 PPS",
							"max_nic_num":  12,
							"max_disk_num": 24,
						},
					},
				},
			},
		},
	})

	if len(rows) != 1 {
		t.Fatalf("rows = %+v", rows)
	}
	if got := rows[0]["id"]; got != "c3.large.2" {
		t.Fatalf("id = %#v", got)
	}
	if got := rows[0]["zone_id"]; got != "cn-north-1a" {
		t.Fatalf("zone_id = %#v", got)
	}
}
