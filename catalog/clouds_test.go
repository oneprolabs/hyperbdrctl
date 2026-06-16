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

	if len(got) != 30 {
		t.Fatalf("EnabledBlockClouds() len = %d, want %d", len(got), 30)
	}
	if containsCloudType(got, "huaweicloud") {
		t.Fatalf("EnabledBlockClouds() unexpectedly contains disabled huaweicloud entry")
	}
}

func TestEnabledObjectCloudsReturnsOnlyEnabledEntries(t *testing.T) {
	got := EnabledObjectClouds()

	if len(got) != 22 {
		t.Fatalf("EnabledObjectClouds() len = %d, want %d", len(got), 22)
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
	got, ok := FindBlockCloud("open_telekom")
	if !ok {
		t.Fatalf("FindBlockCloud() ok = false, want true")
	}
	if got.CloudType != "open_telekom_bs" {
		t.Fatalf("FindBlockCloud() cloudType = %q, want %q", got.CloudType, "open_telekom_bs")
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
	got, ok := FindBlockCloud("  OPEN_TELEKOM_BS ")
	if !ok {
		t.Fatalf("FindBlockCloud() ok = false, want true")
	}
	if got.Key != "open_telekom_bs_block" {
		t.Fatalf("FindBlockCloud() key = %q, want %q", got.Key, "open_telekom_bs_block")
	}
}

func TestFindCloudMiss(t *testing.T) {
	if _, ok := FindBlockCloud("does-not-exist"); ok {
		t.Fatalf("FindBlockCloud() ok = true, want false")
	}
}

func TestBlockCloudTypesStableOrdered(t *testing.T) {
	want := []string{
		"aliyun",
		"aliyun_bs",
		"apsara316_bs",
		"apsara318_bs",
		"tencentcloud",
		"tce_bs",
		"tstackenterprise",
		"tstack",
		"huawei_bs",
		"hcso_bs",
		"hwfc80",
		"aws_cn_v2_bs",
		"aws_v2_bs",
		"azure_bs",
		"ucloudstack_bs",
		"qcloud",
		"yidongecloud",
		"yidongjointcloud",
		"esurfingcloud_bs",
		"openstack",
		"tmcloud",
		"oracle_bs",
		"google_bs",
		"smartx_bs",
		"open_telekom_bs",
		"lvneng_bs",
		"zstack",
		"xhere_bs",
		"jinshancloud",
		"fixo_bs",
	}
	if got := BlockCloudTypes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("BlockCloudTypes() = %#v, want %#v", got, want)
	}
}

func TestObjectCloudTypesStableOrdered(t *testing.T) {
	want := []string{
		"aliyun",
		"aliyun_obs",
		"apsara316",
		"apsara318",
		"tencent_obs",
		"tce_obs",
		"huawei_obs",
		"hcso_obs",
		"fusioncompute_obs",
		"volc",
		"aws_cn_v2_obs",
		"aws_v2_obs",
		"ucloud",
		"yidongecloud",
		"ctyun_obs",
		"openstack",
		"tmcloud",
		"open_telekom_obs",
		"vmware_obs",
		"xhere",
		"ens",
		"vmware",
	}
	if got := ObjectCloudTypes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ObjectCloudTypes() = %#v, want %#v", got, want)
	}
}

func TestEnabledHelpersReturnCopies(t *testing.T) {
	got := EnabledBlockClouds()
	got[0].NameEn = "mutated"

	if BlockClouds[0].NameEn != "Alibaba Cloud(Not Recommended)" {
		t.Fatalf("BlockClouds mutated = %q, want original value", BlockClouds[0].NameEn)
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
	block := findCloudEntryByType(t, EnabledBlockClouds(), "open_telekom_bs")
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
