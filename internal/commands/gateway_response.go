package commands

import (
	"hyperbdr-client/internal/client"
	normalizegateway "hyperbdr-client/internal/normalize/gateway"
)

func gatewayResponseData(resp client.APIResponse) interface{} {
	return normalizegateway.Data(resp)
}

func gatewayResponseMap(resp client.APIResponse) map[string]interface{} {
	payload, _ := gatewayResponseData(resp).(map[string]interface{})
	return payload
}

func gatewayResponseCloudInfo(resp client.APIResponse) map[string]interface{} {
	return normalizegateway.NestedMap(gatewayResponseData(resp), "cloud_info")
}
