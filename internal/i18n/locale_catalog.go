package i18n

import "fmt"

type localeFragment struct {
	name string
	en   map[string]string
	zh   map[string]string
}

var localeFragments = []localeFragment{
	{name: "agent", en: agentEN, zh: agentZH},
	{name: "boot_config_aliyun", en: boot_config_aliyunEN, zh: boot_config_aliyunZH},
	{name: "boot_config_common", en: boot_config_commonEN, zh: boot_config_commonZH},
	{name: "boot_config_huawei", en: boot_config_huaweiEN, zh: boot_config_huaweiZH},
	{name: "boot_config_openstack", en: boot_config_openstackEN, zh: boot_config_openstackZH},
	{name: "cloud_account_aliyun", en: cloud_account_aliyunEN, zh: cloud_account_aliyunZH},
	{name: "cloud_account_common", en: cloud_account_commonEN, zh: cloud_account_commonZH},
	{name: "cloud_account_huawei", en: cloud_account_huaweiEN, zh: cloud_account_huaweiZH},
	{name: "cloud_account_openstack", en: cloud_account_openstackEN, zh: cloud_account_openstackZH},
	{name: "cloud_resource_aliyun", en: cloud_resource_aliyunEN, zh: cloud_resource_aliyunZH},
	{name: "cloud_resource_common", en: cloud_resource_commonEN, zh: cloud_resource_commonZH},
	{name: "cloud_resource_huawei", en: cloud_resource_huaweiEN, zh: cloud_resource_huaweiZH},
	{name: "cloud_resource_openstack", en: cloud_resource_openstackEN, zh: cloud_resource_openstackZH},
	{name: "cloud_sync_gateway_aliyun", en: cloud_sync_gateway_aliyunEN, zh: cloud_sync_gateway_aliyunZH},
	{name: "cloud_sync_gateway_common", en: cloud_sync_gateway_commonEN, zh: cloud_sync_gateway_commonZH},
	{name: "cloud_sync_gateway_huawei", en: cloud_sync_gateway_huaweiEN, zh: cloud_sync_gateway_huaweiZH},
	{name: "cloud_sync_gateway_openstack", en: cloud_sync_gateway_openstackEN, zh: cloud_sync_gateway_openstackZH},
	{name: "common", en: commonEN, zh: commonZH},
	{name: "config", en: configEN, zh: configZH},
	{name: "host", en: hostEN, zh: hostZH},
	{name: "license", en: licenseEN, zh: licenseZH},
	{name: "oss", en: ossEN, zh: ossZH},
	{name: "production_site", en: production_siteEN, zh: production_siteZH},
	{name: "sync_proxy", en: sync_proxyEN, zh: sync_proxyZH},
}

var en, zhCN = mustBuildLocaleCatalog(localeFragments)

func buildLocaleCatalog(fragments []localeFragment) (map[string]string, map[string]string, error) {
	en := make(map[string]string)
	zh := make(map[string]string)
	ownersEN := make(map[string]string)
	ownersZH := make(map[string]string)

	for _, fragment := range fragments {
		if fragment.name == "" {
			return nil, nil, fmt.Errorf("locale fragment has empty name")
		}
		for key, value := range fragment.en {
			if value == "" {
				return nil, nil, fmt.Errorf("locale fragment %q has empty en translation for key %q", fragment.name, key)
			}
			if _, ok := fragment.zh[key]; !ok {
				return nil, nil, fmt.Errorf("locale fragment %q is missing zh_cn translation for key %q", fragment.name, key)
			}
			if previous, ok := ownersEN[key]; ok {
				return nil, nil, fmt.Errorf("duplicate en translation key %q in fragments %q and %q", key, previous, fragment.name)
			}
			if previous, ok := ownersZH[key]; ok {
				return nil, nil, fmt.Errorf("duplicate zh_cn translation key %q in fragments %q and %q", key, previous, fragment.name)
			}
			if fragment.zh[key] == "" {
				return nil, nil, fmt.Errorf("locale fragment %q has empty zh_cn translation for key %q", fragment.name, key)
			}
			en[key] = value
			zh[key] = fragment.zh[key]
			ownersEN[key] = fragment.name
			ownersZH[key] = fragment.name
		}
		for key := range fragment.zh {
			if _, ok := fragment.en[key]; !ok {
				return nil, nil, fmt.Errorf("locale fragment %q is missing en translation for key %q", fragment.name, key)
			}
		}
	}
	return en, zh, nil
}

func mustBuildLocaleCatalog(fragments []localeFragment) (map[string]string, map[string]string) {
	en, zh, err := buildLocaleCatalog(fragments)
	if err != nil {
		panic(err)
	}
	return en, zh
}
