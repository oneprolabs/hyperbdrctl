package sourcecreate

import (
	"fmt"
	"strings"
)

func required(value, name string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func synchNodeIDs(single, csv string) ([]string, error) {
	ids := make([]string, 0, 4)
	seen := map[string]bool{}
	if single = strings.TrimSpace(single); single != "" {
		ids = append(ids, single)
		seen[single] = true
	}
	for _, part := range strings.Split(csv, ",") {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		ids = append(ids, part)
		seen[part] = true
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("synch-node-id or synch-node-ids is required")
	}
	return ids, nil
}

func baseConnectionBody(connectionType string, synchNodeIDs []string, key string, authFields map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"connection": map[string]interface{}{
			"type":           connectionType,
			"synch_node_ids": synchNodeIDs,
			key:              authFields,
		},
	}
}
