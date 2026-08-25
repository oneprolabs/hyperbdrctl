package sourcecreate

import "testing"

func TestBuildRequestVMwareMapsToVSphere(t *testing.T) {
	path, body, err := BuildRequest(Spec{
		Type:        "vmware",
		SynchNodeID: "node-1",
		AuthURL:     "https://vcenter.invalid",
		AuthKey:     "user",
		AuthCert:    "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/api/v2/createConnection" {
		t.Fatalf("path = %q", path)
	}
	connection := body["connection"].(map[string]interface{})
	if connection["type"] != "vsphere" {
		t.Fatalf("connection.type = %v", connection["type"])
	}
	if got := connection["synch_node_ids"].([]string); len(got) != 1 || got[0] != "node-1" {
		t.Fatalf("synch_node_ids = %#v", got)
	}
	vsphere := connection["vsphere"].(map[string]interface{})
	if vsphere["auth_url"] != "https://vcenter.invalid" || vsphere["auth_key"] != "user" || vsphere["auth_cert"] != "secret" {
		t.Fatalf("vsphere payload = %#v", vsphere)
	}
}

func TestBuildRequestAWSKeepsAWSPayload(t *testing.T) {
	_, body, err := BuildRequest(Spec{
		Type:        "aws",
		SynchNodeID: "node-1",
		AuthURL:     "aws.example",
		AuthKey:     "ak",
		AuthCert:    "sk",
		RegionID:    "cn-test-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	connection := body["connection"].(map[string]interface{})
	if connection["type"] != "aws" {
		t.Fatalf("connection.type = %v", connection["type"])
	}
	aws := connection["aws"].(map[string]interface{})
	if aws["region_id"] != "cn-test-1" {
		t.Fatalf("aws.region_id = %v", aws["region_id"])
	}
}

func TestBuildRequestValidation(t *testing.T) {
	tests := []struct {
		name string
		spec Spec
		want string
	}{
		{
			name: "unsupported type",
			spec: Spec{Type: "vsphere"},
			want: "type must be one of vmware, aws",
		},
		{
			name: "missing node id",
			spec: Spec{Type: "vmware", AuthURL: "u", AuthKey: "k", AuthCert: "c"},
			want: "synch-node-id is required",
		},
		{
			name: "missing aws region",
			spec: Spec{Type: "aws", SynchNodeID: "node-1", AuthURL: "u", AuthKey: "k", AuthCert: "c"},
			want: "region-id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := BuildRequest(tt.spec)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("err = %v, want %q", err, tt.want)
			}
		})
	}
}
