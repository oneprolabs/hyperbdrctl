package sourcecreate

import (
	"fmt"
	"strings"
)

type adapter struct {
	build func(Spec) (string, map[string]interface{}, error)
}

var adapters = map[string]adapter{
	"vmware": {build: buildVMware},
	"aws":    {build: buildAWS},
}

func BuildRequest(spec Spec) (string, map[string]interface{}, error) {
	sourceType := normalizeType(spec.Type)
	adapter, ok := adapters[sourceType]
	if !ok {
		return "", nil, fmt.Errorf("type must be one of vmware, aws")
	}
	spec.Type = sourceType
	return adapter.build(spec)
}

func normalizeType(sourceType string) string {
	return strings.ToLower(strings.TrimSpace(sourceType))
}
