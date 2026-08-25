package catalog

import (
	"reflect"
	"testing"
)

func TestProvidersAreUniqueWithinEachCatalog(t *testing.T) {
	assertUniqueProviders(t, "BlockClouds", BlockClouds)
	assertUniqueProviders(t, "ObjectClouds", ObjectClouds)
}

func TestEnabledBlockCloudsReturnsOnlyEnabledEntries(t *testing.T) {
	got := EnabledBlockClouds()

	if len(got) != 3 {
		t.Fatalf("EnabledBlockClouds() len = %d, want %d", len(got), 3)
	}
	if containsCloudType(got, "huaweicloud") {
		t.Fatalf("EnabledBlockClouds() unexpectedly contains disabled huaweicloud entry")
	}
}

func TestEnabledObjectCloudsReturnsOnlyEnabledEntries(t *testing.T) {
	got := EnabledObjectClouds()

	if len(got) != 3 {
		t.Fatalf("EnabledObjectClouds() len = %d, want %d", len(got), 3)
	}
	if containsCloudType(got, "aws_obs") {
		t.Fatalf("EnabledObjectClouds() unexpectedly contains disabled aws_obs entry")
	}
}

func TestFindBlockCloudByCloudType(t *testing.T) {
	got, ok := FindBlockCloud("aliyun_bs")
	if !ok {
		t.Fatalf("FindBlockCloud() ok = false, want true")
	}
	if got.Key != "aliyun_bs_block" {
		t.Fatalf("FindBlockCloud() key = %q, want %q", got.Key, "aliyun_bs_block")
	}
}

func TestFindObjectCloudByCloudType(t *testing.T) {
	got, ok := FindObjectCloud("aliyun_obs")
	if !ok {
		t.Fatalf("FindObjectCloud() ok = false, want true")
	}
	if got.Key != "aliyun_obs_object" {
		t.Fatalf("FindObjectCloud() key = %q, want %q", got.Key, "aliyun_obs_object")
	}
}

func TestFindBlockCloudByProvider(t *testing.T) {
	got, ok := FindBlockCloud("huawei")
	if !ok {
		t.Fatalf("FindBlockCloud() ok = false, want true")
	}
	if got.CloudType != "huawei_bs" {
		t.Fatalf("FindBlockCloud() cloudType = %q, want %q", got.CloudType, "huawei_bs")
	}
}

func TestFindObjectCloudByCloudTypeSharedProvider(t *testing.T) {
	got, ok := FindObjectCloud("openstack")
	if !ok {
		t.Fatalf("FindObjectCloud() ok = false, want true")
	}
	if got.Key != "openstack_object" {
		t.Fatalf("FindObjectCloud() key = %q, want %q", got.Key, "openstack_object")
	}
}

func TestFindCloudTrimsWhitespaceAndIgnoresCase(t *testing.T) {
	got, ok := FindBlockCloud("  HUAWEI_BS ")
	if !ok {
		t.Fatalf("FindBlockCloud() ok = false, want true")
	}
	if got.Key != "huawei_bs_block" {
		t.Fatalf("FindBlockCloud() key = %q, want %q", got.Key, "huawei_bs_block")
	}
}

func TestFindCloudDoesNotReturnDisabledEntry(t *testing.T) {
	if _, ok := FindBlockCloud("open_telekom_bs"); ok {
		t.Fatalf("FindBlockCloud() returned disabled open_telekom_bs entry")
	}
	if _, ok := FindObjectCloud("aws_v2_obs"); ok {
		t.Fatalf("FindObjectCloud() returned disabled aws_v2_obs entry")
	}
}

func TestFindCloudMiss(t *testing.T) {
	if _, ok := FindBlockCloud("does-not-exist"); ok {
		t.Fatalf("FindBlockCloud() ok = true, want false")
	}
}

func TestBlockCloudTypesStableOrdered(t *testing.T) {
	want := []string{
		"aliyun_bs",
		"huawei_bs",
		"openstack",
	}
	if got := BlockCloudTypes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("BlockCloudTypes() = %#v, want %#v", got, want)
	}
}

func TestObjectCloudTypesStableOrdered(t *testing.T) {
	want := []string{
		"aliyun_obs",
		"huawei_obs",
		"openstack",
	}
	if got := ObjectCloudTypes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ObjectCloudTypes() = %#v, want %#v", got, want)
	}
}

func TestEnabledHelpersReturnCopies(t *testing.T) {
	got := EnabledBlockClouds()
	got[0].NameEn = "mutated"

	original := findCloudEntryByType(t, BlockClouds, "aliyun_bs")
	if original.NameEn != "Alibaba Cloud(Recommended, SDK v2.0)" {
		t.Fatalf("BlockClouds mutated = %q, want original value", original.NameEn)
	}
}

func TestBlockCloudArchitectures(t *testing.T) {
	tests := []struct {
		name      string
		cloudType string
		want      string
	}{
		{name: "bs provider uses AtomyV2", cloudType: "aliyun_bs", want: AtomyV2},
		{name: "legacy provider uses NotAtomy", cloudType: "openstack", want: NotAtomy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findCloudEntryByType(t, BlockClouds, tt.cloudType)
			if got.Architecture != tt.want {
				t.Fatalf("BlockClouds %q architecture = %q, want %q", tt.cloudType, got.Architecture, tt.want)
			}
		})
	}
}

func TestObjectCloudArchitectures(t *testing.T) {
	tests := []struct {
		name      string
		cloudType string
		want      string
	}{
		{name: "obs provider uses AtomyV2", cloudType: "aliyun_obs", want: AtomyV2},
		{name: "legacy provider uses NotAtomy", cloudType: "vmware", want: NotAtomy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findCloudEntryByType(t, ObjectClouds, tt.cloudType)
			if got.Architecture != tt.want {
				t.Fatalf("ObjectClouds %q architecture = %q, want %q", tt.cloudType, got.Architecture, tt.want)
			}
		})
	}
}

func TestEnabledCloudsPreserveArchitecture(t *testing.T) {
	block := findCloudEntryByType(t, EnabledBlockClouds(), "aliyun_bs")
	if block.Architecture != AtomyV2 {
		t.Fatalf("EnabledBlockClouds() architecture for %q = %q, want %q", block.CloudType, block.Architecture, AtomyV2)
	}

	object := findCloudEntryByType(t, EnabledObjectClouds(), "openstack")
	if object.Architecture != NotAtomy {
		t.Fatalf("EnabledObjectClouds() architecture for %q = %q, want %q", object.CloudType, object.Architecture, NotAtomy)
	}
}

func containsCloudType(items []CloudEntry, cloudType string) bool {
	for _, item := range items {
		if item.CloudType == cloudType {
			return true
		}
	}
	return false
}

func findCloudEntryByType(t *testing.T, items []CloudEntry, cloudType string) CloudEntry {
	t.Helper()

	for _, item := range items {
		if item.CloudType == cloudType {
			return item
		}
	}

	t.Fatalf("cloud type %q not found", cloudType)
	return CloudEntry{}
}

func assertUniqueProviders(t *testing.T, name string, items []CloudEntry) {
	t.Helper()

	seen := make(map[string]string, len(items))
	for _, item := range items {
		if prevKey, ok := seen[item.Provider]; ok {
			t.Fatalf("%s provider %q is duplicated by %q and %q", name, item.Provider, prevKey, item.Key)
		}
		seen[item.Provider] = item.Key
	}
}
