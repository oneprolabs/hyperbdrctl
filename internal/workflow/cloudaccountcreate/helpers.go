package cloudaccountcreate

const (
	yesLabel                      = "\u662f"
	autoRegionLabel               = "\u81ea\u52a8\u83b7\u53d6"
	autoUploadLabel               = "\u81ea\u52a8\u4e0a\u4f20"
	aliyunObjectDefaultNamePrefix = "\u963f\u91cc\u4e91(\u63a8\u8350\u4f7f\u7528\uff0cSDK v2.0)-"
)

func boolOrNil(v *bool) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func valueOrNil(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}

func emptyStringOrNil(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}

func firstNonEmptyString(values ...interface{}) string {
	for _, value := range values {
		if s, ok := value.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func defaultAliyunObjectCustomName(regionLabel string) string {
	return aliyunObjectDefaultNamePrefix + firstNonEmptyString(regionLabel, "unknown-region")
}
