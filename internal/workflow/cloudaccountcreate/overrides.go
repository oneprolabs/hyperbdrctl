package cloudaccountcreate

func finalizeCreateBody(spec Spec, body map[string]interface{}) map[string]interface{} {
	cloudAccountValue, ok := body["cloud_account"]
	if !ok {
		return body
	}
	cloudAccount, ok := cloudAccountValue.(map[string]interface{})
	if !ok {
		return body
	}

	if metadataValue, ok := cloudAccount["metadata"]; ok {
		if metadata, ok := metadataValue.(map[string]interface{}); ok {
			cloudAccount["metadata"] = mergeMap(metadata, spec.MetadataOverrides, map[string]struct{}{
				"cloud_type": {},
			})
		}
	}

	body = mergeAllowedRootOverrides(body, spec.RequestOverrides)
	body["cloud_account"] = cloudAccount
	return body
}

func mergeAllowedRootOverrides(body map[string]interface{}, overrides map[string]interface{}) map[string]interface{} {
	if len(overrides) == 0 {
		return body
	}
	for _, key := range []string{"auto_upload_images", "only_verify"} {
		value, ok := overrides[key]
		if ok {
			body[key] = value
		}
	}
	return body
}

func mergeMap(base map[string]interface{}, overrides map[string]interface{}, protected map[string]struct{}) map[string]interface{} {
	if len(overrides) == 0 {
		return base
	}
	if base == nil {
		base = map[string]interface{}{}
	}
	for key, value := range overrides {
		if _, blocked := protected[key]; blocked {
			continue
		}
		if existing, ok := base[key].(map[string]interface{}); ok {
			overrideMap, ok := value.(map[string]interface{})
			if ok {
				base[key] = mergeMap(existing, overrideMap, nil)
				continue
			}
		}
		base[key] = value
	}
	return base
}
