package commands

import (
	"strings"
	"testing"
)

func TestSelectHuaweiObjectNetworkAndSubnetValidatesRelationship(t *testing.T) {
	networks := []map[string]interface{}{
		{"id": "network-1", "is_recommend": 1},
		{"id": "network-2"},
	}
	subnets := []map[string]interface{}{
		{"id": "subnet-1", "network_id": "network-1", "zone_id": "zone-1"},
		{"id": "subnet-2", "network_id": "network-2", "zone_id": "zone-2"},
	}

	network, subnet, err := selectHuaweiObjectNetworkAndSubnet(networks, subnets, "", "", "zone-1")
	if err != nil {
		t.Fatal(err)
	}
	if mapString(network, "id") != "network-1" || mapString(subnet, "id") != "subnet-1" {
		t.Fatalf("network=%v subnet=%v", network, subnet)
	}

	_, _, err = selectHuaweiObjectNetworkAndSubnet(networks, subnets, "network-1", "subnet-2", "")
	if err == nil || !strings.Contains(err.Error(), "belongs to network") {
		t.Fatalf("err = %v", err)
	}
}

func TestSelectHuaweiObjectFlavorPrefersRecommended2C4GAndBuildsPath(t *testing.T) {
	rows := []map[string]interface{}{
		{
			"value": "u-2",
			"children": []interface{}{map[string]interface{}{
				"value": "u-2-m-4",
				"children": []interface{}{
					map[string]interface{}{"id": "not-recommended", "vcpus": 2, "ram_GB": 4},
					map[string]interface{}{"id": "recommended", "vcpus": 2, "ram_GB": 4, "is_recommend": 1},
				},
			}},
		},
	}

	row, path, err := selectHuaweiObjectFlavor(rows, "")
	if err != nil {
		t.Fatal(err)
	}
	if mapString(row, "id") != "recommended" {
		t.Fatalf("row = %+v", row)
	}
	want := []string{"u-2", "u-2-m-4", "recommended"}
	if len(path) != len(want) {
		t.Fatalf("path = %v", path)
	}
	for index := range want {
		if path[index] != want[index] {
			t.Fatalf("path = %v", path)
		}
	}
}

func TestSelectHuaweiObjectFlavorRejectsMissing2C4GDefault(t *testing.T) {
	_, _, err := selectHuaweiObjectFlavor([]map[string]interface{}{
		{"id": "large", "vcpus": 4, "ram_GB": 8},
	}, "")
	if err == nil || !strings.Contains(err.Error(), "no 2 vCPU / 4 GiB") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateHuaweiObjectControlFields(t *testing.T) {
	if err := validateHuaweiObjectControlFields(cloudAccountCreateSpec{UseInternalIP: "2"}); err == nil {
		t.Fatal("expected invalid use-internal-ip error")
	}
	if err := validateHuaweiObjectControlFields(cloudAccountCreateSpec{ControlAccessIP: "invalid"}); err == nil {
		t.Fatal("expected invalid control-access-ip error")
	}
	if err := validateHuaweiObjectControlFields(cloudAccountCreateSpec{UseInternalIP: "1", ControlAccessIP: "192.0.2.10"}); err != nil {
		t.Fatal(err)
	}
}

func TestHuaweiObjectUnsupportedCreateFlags(t *testing.T) {
	for _, name := range []string{
		"username", "linux-boot-image-id", "windows-boot-image-id", "boot-image-source", "skip-driver-fix",
		"control-access-ip-radio", "linux-boot-image-host-config-bandwidth-id", "linux-boot-image-host-config-bandwidth-name",
	} {
		if !huaweiObjectUnsupportedCreateFlag(name) {
			t.Fatalf("%s should be rejected", name)
		}
	}
	if huaweiObjectUnsupportedCreateFlag("linux-boot-image-host-config-image-id") {
		t.Fatal("host config image ID should be supported")
	}
}
