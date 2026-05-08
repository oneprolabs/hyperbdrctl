package catalog

import "strings"

type CloudEntry struct {
	Key       string `json:"key"`
	Provider  string `json:"provider"`
	CloudType string `json:"cloud_type"`
	NameZhCN  string `json:"name_zh_cn"`
	NameEn    string `json:"name_en"`
	Enabled   bool   `json:"enabled"`
}

var BlockClouds = []CloudEntry{
	{
		Key:       "aliyun_block",
		Provider:  "aliyun_disused",
		CloudType: "aliyun",
		NameZhCN:  "阿里云(即将退役，不推荐)",
		NameEn:    "Alibaba Cloud(Not Recommended)",
		Enabled:   true,
	},
	{
		Key:       "aliyun_bs_block",
		Provider:  "aliyun",
		CloudType: "aliyun_bs",
		NameZhCN:  "阿里云(推荐使用，SDK v2.0)",
		NameEn:    "Alibaba Cloud(Recommended, SDK v2.0)",
		Enabled:   true,
	},
	{
		Key:       "apsara316_bs_block",
		Provider:  "apsara316",
		CloudType: "apsara316_bs",
		NameZhCN:  "阿里云 专有云(v3.16.x)",
		NameEn:    "Alibaba Cloud Apsara Stack(v3.16.x)",
		Enabled:   true,
	},
	{
		Key:       "apsara318_bs_block",
		Provider:  "apsara318",
		CloudType: "apsara318_bs",
		NameZhCN:  "阿里云 专有云(v3.18.x)",
		NameEn:    "Alibaba Cloud Apsara Stack(v3.18.x)",
		Enabled:   true,
	},
	{
		Key:       "tencentcloud_block",
		Provider:  "tencentcloud",
		CloudType: "tencentcloud",
		NameZhCN:  "腾讯云",
		NameEn:    "Tencent Cloud",
		Enabled:   true,
	},
	{
		Key:       "tce_bs_block",
		Provider:  "tce",
		CloudType: "tce_bs",
		NameZhCN:  "腾讯云 专有云企业版",
		NameEn:    "Tencent Cloud Enterprise",
		Enabled:   true,
	},
	{
		Key:       "tstackenterprise_block",
		Provider:  "tstackenterprise",
		CloudType: "tstackenterprise",
		NameZhCN:  "腾讯云 TStack企业版",
		NameEn:    "Tencent Cloud TStack Enterprise",
		Enabled:   true,
	},
	{
		Key:       "tstack_block",
		Provider:  "tstack",
		CloudType: "tstack",
		NameZhCN:  "腾讯云 TStack旗舰版",
		NameEn:    "Tencent Cloud TStack Ultimate",
		Enabled:   true,
	},
	{
		Key:       "huawei_bs_block",
		Provider:  "huawei",
		CloudType: "huawei_bs",
		NameZhCN:  "华为云(推荐使用，SDK v3.1.86)",
		NameEn:    "Huawei Cloud(Recommended, SDK v3.1.86)",
		Enabled:   true,
	},
	{
		Key:       "huaweicloud_block",
		Provider:  "huaweicloud",
		CloudType: "huaweicloud",
		NameZhCN:  "华为云(已废弃，请使用华为云 SDK v3.1.86)",
		NameEn:    "Huawei Cloud(Obsolete, Use Huawei Cloud SDK v3.1.86)",
		Enabled:   false,
	},
	{
		Key:       "hcso_bs_block",
		Provider:  "hcso",
		CloudType: "hcso_bs",
		NameZhCN:  "华为云 HCS Online(v23.3)",
		NameEn:    "Huawei Cloud Stack Online(v23.3)",
		Enabled:   true,
	},
	{
		Key:       "hwfc80_block",
		Provider:  "hwfc80",
		CloudType: "hwfc80",
		NameZhCN:  "华为云Stack(HCS)(v8.2.x / v8.3.x)",
		NameEn:    "Huawei Cloud Stack(HCS)(v8.2.x / v8.3.x)",
		Enabled:   true,
	},
	{
		Key:       "hwfclegacy_block",
		Provider:  "hwfclegacy",
		CloudType: "hwfclegacy",
		NameZhCN:  "华为云Stack(HCS)(v8.1.x)",
		NameEn:    "Huawei Cloud Stack(HCS)(v8.1.x)",
		Enabled:   false,
	},
	{
		Key:       "fusioncompute_bs_block",
		Provider:  "fusioncompute",
		CloudType: "fusioncompute_bs",
		NameZhCN:  "FusionCompute(v8.6.x)",
		NameEn:    "FusionCompute(v8.6.x)",
		Enabled:   false,
	},
	{
		Key:       "awsebs_block",
		Provider:  "awsebs",
		CloudType: "awsebs",
		NameZhCN:  "AWS",
		NameEn:    "AWS",
		Enabled:   false,
	},
	{
		Key:       "aws_cn_v2_bs_block",
		Provider:  "aws_cn_v2",
		CloudType: "aws_cn_v2_bs",
		NameZhCN:  "AWS中国(SDK v1.34.93)",
		NameEn:    "AWS China(SDK v1.34.93)",
		Enabled:   true,
	},
	{
		Key:       "aws_v2_bs_block",
		Provider:  "aws_v2",
		CloudType: "aws_v2_bs",
		NameZhCN:  "AWS(SDK v1.34.93)",
		NameEn:    "AWS(SDK v1.34.93)",
		Enabled:   true,
	},
	{
		Key:       "azurecloud_block",
		Provider:  "azurecloud",
		CloudType: "azurecloud",
		NameZhCN:  "Azure",
		NameEn:    "Azure",
		Enabled:   false,
	},
	{
		Key:       "azure_bs_block",
		Provider:  "azure",
		CloudType: "azure_bs",
		NameZhCN:  "Microsoft Azure(SDK v30.3)",
		NameEn:    "Microsoft Azure(SDK v30.3)",
		Enabled:   true,
	},
	{
		Key:       "ucloudstack_bs_block",
		Provider:  "ucloudstack",
		CloudType: "ucloudstack_bs",
		NameZhCN:  "UCloudStack",
		NameEn:    "UCloudStack",
		Enabled:   true,
	},
	{
		Key:       "qcloud_block",
		Provider:  "qcloud",
		CloudType: "qcloud",
		NameZhCN:  "青云",
		NameEn:    "QingCloud",
		Enabled:   true,
	},
	{
		Key:       "yidongecloud_block",
		Provider:  "yidongecloud",
		CloudType: "yidongecloud",
		NameZhCN:  "移动云",
		NameEn:    "ecloud",
		Enabled:   true,
	},
	{
		Key:       "yidongjointcloud_block",
		Provider:  "yidongjointcloud",
		CloudType: "yidongjointcloud",
		NameZhCN:  "移动和云",
		NameEn:    "ecloud JC",
		Enabled:   true,
	},
	{
		Key:       "esurfingcloud_bs_block",
		Provider:  "esurfingcloud",
		CloudType: "esurfingcloud_bs",
		NameZhCN:  "天翼云4.0",
		NameEn:    "eSurfingCloud4.0",
		Enabled:   true,
	},
	{
		Key:       "openstack_block",
		Provider:  "openstack",
		CloudType: "openstack",
		NameZhCN:  "OpenStack社区版本(Juno+)",
		NameEn:    "OpenStackCommunity(Juno+)",
		Enabled:   true,
	},
	{
		Key:       "tmcloud_block",
		Provider:  "tmcloud",
		CloudType: "tmcloud",
		NameZhCN:  "TM CAE",
		NameEn:    "TM CAE",
		Enabled:   true,
	},
	{
		Key:       "oracle_bs_block",
		Provider:  "oracle",
		CloudType: "oracle_bs",
		NameZhCN:  "甲骨文云(SDK v2.126.3)",
		NameEn:    "Oracle Cloud(SDK v2.126.3)",
		Enabled:   true,
	},
	{
		Key:       "google_bs_block",
		Provider:  "google",
		CloudType: "google_bs",
		NameZhCN:  "Google Cloud(SDK v1.19.0)",
		NameEn:    "Google Cloud(SDK v1.19.0)",
		Enabled:   true,
	},
	{
		Key:       "smartx_bs_block",
		Provider:  "smartx",
		CloudType: "smartx_bs",
		NameZhCN:  "SMTX OS(v6.x.x)",
		NameEn:    "SMTX OS(v6.x.x)",
		Enabled:   true,
	},
	{
		Key:       "open_telekom_bs_block",
		Provider:  "open_telekom",
		CloudType: "open_telekom_bs",
		NameZhCN:  "Open Telekom Cloud(SDK v3.1.86)",
		NameEn:    "Open Telekom Cloud(SDK v3.1.86)",
		Enabled:   true,
	},
	{
		Key:       "lvneng_bs_block",
		Provider:  "lvneng",
		CloudType: "lvneng_bs",
		NameZhCN:  "绿能云",
		NameEn:    "GridCloud",
		Enabled:   true,
	},
	{
		Key:       "zstack_block",
		Provider:  "zstack",
		CloudType: "zstack",
		NameZhCN:  "ZStack(v4.x.x)",
		NameEn:    "ZStack(v4.x.x)",
		Enabled:   true,
	},
	{
		Key:       "xhere_bs_block",
		Provider:  "xhere",
		CloudType: "xhere_bs",
		NameZhCN:  "XHERE(NeutonOS_3.x)",
		NameEn:    "XHERE(NeutonOS_3.x)",
		Enabled:   true,
	},
	{
		Key:       "jinshancloud_block",
		Provider:  "jinshancloud",
		CloudType: "jinshancloud",
		NameZhCN:  "金山云",
		NameEn:    "Jinshan Cloud",
		Enabled:   true,
	},
	{
		Key:       "fixo_bs_block",
		Provider:  "fixo",
		CloudType: "fixo_bs",
		NameZhCN:  "FiXo Cloud BS",
		NameEn:    "FiXo Cloud BS",
		Enabled:   true,
	},
}

var ObjectClouds = []CloudEntry{
	{
		Key:       "aliyun_object",
		Provider:  "aliyun_disused",
		CloudType: "aliyun",
		NameZhCN:  "阿里云(即将退役，不推荐)",
		NameEn:    "Alibaba Cloud(Not Recommended)",
		Enabled:   true,
	},
	{
		Key:       "aliyun_obs_object",
		Provider:  "aliyun",
		CloudType: "aliyun_obs",
		NameZhCN:  "阿里云(推荐使用，SDK v2.0)",
		NameEn:    "Alibaba Cloud(Recommended, SDK v2.0)",
		Enabled:   true,
	},
	{
		Key:       "apsara316_object",
		Provider:  "apsara316",
		CloudType: "apsara316",
		NameZhCN:  "阿里云 专有云(v3.16.x)",
		NameEn:    "Alibaba Cloud Apsara Stack(v3.16.x)",
		Enabled:   true,
	},
	{
		Key:       "apsara318_object",
		Provider:  "apsara318",
		CloudType: "apsara318",
		NameZhCN:  "阿里云 专有云(v3.18.x)",
		NameEn:    "Alibaba Cloud Apsara Stack(v3.18.x)",
		Enabled:   true,
	},
	{
		Key:       "tencent_obs_object",
		Provider:  "tencent",
		CloudType: "tencent_obs",
		NameZhCN:  "腾讯云",
		NameEn:    "Tencent Cloud",
		Enabled:   true,
	},
	{
		Key:       "tce_obs_object",
		Provider:  "tce",
		CloudType: "tce_obs",
		NameZhCN:  "腾讯云 专有云企业版",
		NameEn:    "Tencent Cloud Enterprise",
		Enabled:   true,
	},
	{
		Key:       "huawei_obs_object",
		Provider:  "huawei",
		CloudType: "huawei_obs",
		NameZhCN:  "华为云(推荐使用，SDK v3.1.86)",
		NameEn:    "Huawei Cloud(Recommended, SDK v3.1.86)",
		Enabled:   true,
	},
	{
		Key:       "huaweicloud_object",
		Provider:  "huaweicloud",
		CloudType: "huaweicloud",
		NameZhCN:  "华为云(已废弃，请使用华为云 SDK v3.1.86)",
		NameEn:    "Huawei Cloud(Obsolete, Use Huawei Cloud SDK v3.1.86)",
		Enabled:   false,
	},
	{
		Key:       "hcso_obs_object",
		Provider:  "hcso",
		CloudType: "hcso_obs",
		NameZhCN:  "华为云 HCS Online(v23.3)",
		NameEn:    "Huawei Cloud Stack Online(v23.3)",
		Enabled:   true,
	},
	{
		Key:       "fusioncompute_obs_object",
		Provider:  "fusioncompute",
		CloudType: "fusioncompute_obs",
		NameZhCN:  "FusionCompute(v8.6.x)",
		NameEn:    "FusionCompute(v8.6.x)",
		Enabled:   true,
	},
	{
		Key:       "volc_object",
		Provider:  "volc",
		CloudType: "volc",
		NameZhCN:  "火山引擎",
		NameEn:    "Volcengine",
		Enabled:   true,
	},
	{
		Key:       "aws_cn_obs_object",
		Provider:  "aws_cn",
		CloudType: "aws_cn_obs",
		NameZhCN:  "AWS中国",
		NameEn:    "AWS China",
		Enabled:   false,
	},
	{
		Key:       "aws_cn_v2_obs_object",
		Provider:  "aws_cn_v2",
		CloudType: "aws_cn_v2_obs",
		NameZhCN:  "AWS中国(SDK v1.34.93)",
		NameEn:    "AWS China(SDK v1.34.93)",
		Enabled:   true,
	},
	{
		Key:       "aws_obs_object",
		Provider:  "aws",
		CloudType: "aws_obs",
		NameZhCN:  "AWS国际",
		NameEn:    "AWS",
		Enabled:   false,
	},
	{
		Key:       "aws_v2_obs_object",
		Provider:  "aws_v2",
		CloudType: "aws_v2_obs",
		NameZhCN:  "AWS(SDK v1.34.93)",
		NameEn:    "AWS(SDK v1.34.93)",
		Enabled:   true,
	},
	{
		Key:       "ucloud_object",
		Provider:  "ucloud",
		CloudType: "ucloud",
		NameZhCN:  "UCloud",
		NameEn:    "UCloud",
		Enabled:   true,
	},
	{
		Key:       "yidongecloud_object",
		Provider:  "yidongecloud",
		CloudType: "yidongecloud",
		NameZhCN:  "移动云",
		NameEn:    "ecloud",
		Enabled:   true,
	},
	{
		Key:       "ctyun_obs_object",
		Provider:  "ctyun",
		CloudType: "ctyun_obs",
		NameZhCN:  "天翼云合营云",
		NameEn:    "ctyun JC",
		Enabled:   true,
	},
	{
		Key:       "openstack_object",
		Provider:  "openstack",
		CloudType: "openstack",
		NameZhCN:  "OpenStack社区版本(Juno+)",
		NameEn:    "OpenStackCommunity(Juno+)",
		Enabled:   true,
	},
	{
		Key:       "tmcloud_object",
		Provider:  "tmcloud",
		CloudType: "tmcloud",
		NameZhCN:  "TM CAE",
		NameEn:    "TM CAE",
		Enabled:   true,
	},
	{
		Key:       "oracle_bs_object",
		Provider:  "oracle",
		CloudType: "oracle_bs",
		NameZhCN:  "甲骨文云(SDK v2.126.3)",
		NameEn:    "Oracle Cloud(SDK v2.126.3)",
		Enabled:   false,
	},
	{
		Key:       "google_obs_object",
		Provider:  "google",
		CloudType: "google_obs",
		NameZhCN:  "Google Cloud(SDK v1.19.0)",
		NameEn:    "Google Cloud(SDK v1.19.0)",
		Enabled:   false,
	},
	{
		Key:       "open_telekom_obs_object",
		Provider:  "open_telekom",
		CloudType: "open_telekom_obs",
		NameZhCN:  "Open Telekom Cloud(SDK v3.1.86)",
		NameEn:    "Open Telekom Cloud(SDK v3.1.86)",
		Enabled:   true,
	},
	{
		Key:       "lvneng_obs_object",
		Provider:  "lvneng",
		CloudType: "lvneng_obs",
		NameZhCN:  "绿能云",
		NameEn:    "GridCloud",
		Enabled:   false,
	},
	{
		Key:       "vmware_obs_object",
		Provider:  "vmware",
		CloudType: "vmware_obs",
		NameZhCN:  "VMware vCenter Server",
		NameEn:    "VMware vCenter Server",
		Enabled:   true,
	},
	{
		Key:       "xhere_object",
		Provider:  "xhere",
		CloudType: "xhere",
		NameZhCN:  "XHERE(NeutonOS_3.x)",
		NameEn:    "XHERE(NeutonOS_3.x)",
		Enabled:   true,
	},
	{
		Key:       "ens_object",
		Provider:  "ens",
		CloudType: "ens",
		NameZhCN:  "GDS万国数据本地云",
		NameEn:    "GDS",
		Enabled:   true,
	},
	{
		Key:       "vmware_object",
		Provider:  "vmware_disused",
		CloudType: "vmware",
		NameZhCN:  "VMware(即将退役，不推荐)",
		NameEn:    "VMware(Not Recommended)",
		Enabled:   true,
	},
}

func EnabledBlockClouds() []CloudEntry {
	return enabledClouds(BlockClouds)
}

func EnabledObjectClouds() []CloudEntry {
	return enabledClouds(ObjectClouds)
}

func FindBlockCloud(value string) (CloudEntry, bool) {
	return findCloud(BlockClouds, value)
}

func FindObjectCloud(value string) (CloudEntry, bool) {
	return findCloud(ObjectClouds, value)
}

func BlockCloudTypes() []string {
	return cloudTypes(EnabledBlockClouds())
}

func ObjectCloudTypes() []string {
	return cloudTypes(EnabledObjectClouds())
}

func enabledClouds(items []CloudEntry) []CloudEntry {
	out := make([]CloudEntry, 0, len(items))
	for _, item := range items {
		if item.Enabled {
			out = append(out, item)
		}
	}
	return out
}

func findCloud(items []CloudEntry, value string) (CloudEntry, bool) {
	needle := normalizeValue(value)
	if needle == "" {
		return CloudEntry{}, false
	}
	for _, item := range items {
		if matchesCloud(item, needle) {
			return item, true
		}
	}
	return CloudEntry{}, false
}

func cloudTypes(items []CloudEntry) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.CloudType)
	}
	return out
}

func matchesCloud(item CloudEntry, needle string) bool {
	return normalizeValue(item.Key) == needle ||
		normalizeValue(item.Provider) == needle ||
		normalizeValue(item.CloudType) == needle
}

func normalizeValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
