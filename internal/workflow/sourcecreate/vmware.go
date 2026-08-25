package sourcecreate

func buildVMware(spec Spec) (string, map[string]interface{}, error) {
	if err := required(spec.AuthURL, "auth-url"); err != nil {
		return "", nil, err
	}
	if err := required(spec.AuthKey, "auth-key"); err != nil {
		return "", nil, err
	}
	if err := required(spec.AuthCert, "auth-cert"); err != nil {
		return "", nil, err
	}
	nodeIDs, err := synchNodeIDs(spec.SynchNodeID)
	if err != nil {
		return "", nil, err
	}
	return "/api/v2/createConnection", baseConnectionBody("vsphere", nodeIDs, "vsphere", map[string]interface{}{
		"auth_url":  spec.AuthURL,
		"auth_key":  spec.AuthKey,
		"auth_cert": spec.AuthCert,
	}), nil
}
