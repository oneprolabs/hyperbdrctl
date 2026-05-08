package sourcecreate

func buildAWS(spec Spec) (string, map[string]interface{}, error) {
	if err := required(spec.AuthURL, "auth-url"); err != nil {
		return "", nil, err
	}
	if err := required(spec.AuthKey, "auth-key"); err != nil {
		return "", nil, err
	}
	if err := required(spec.AuthCert, "auth-cert"); err != nil {
		return "", nil, err
	}
	if err := required(spec.RegionID, "region-id"); err != nil {
		return "", nil, err
	}
	nodeIDs, err := synchNodeIDs(spec.SynchNodeID, spec.SynchNodeIDs)
	if err != nil {
		return "", nil, err
	}
	return "/api/v2/createConnection", baseConnectionBody("aws", nodeIDs, "aws", map[string]interface{}{
		"auth_url":  spec.AuthURL,
		"auth_key":  spec.AuthKey,
		"auth_cert": spec.AuthCert,
		"region_id": spec.RegionID,
	}), nil
}
