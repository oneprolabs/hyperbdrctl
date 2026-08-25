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

func synchNodeIDs(single string) ([]string, error) {
	single = strings.TrimSpace(single)
	if single == "" {
		return nil, fmt.Errorf("synch-node-id is required")
	}
	return []string{single}, nil
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
