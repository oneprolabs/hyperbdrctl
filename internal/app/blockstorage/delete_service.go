package blockstorage

import (
	"fmt"
	"strings"

	"hyperbdr-client/internal/client"
)

type DeleteSpec struct {
	ID    string
	IDs   string
	Force bool
}

func (s Service) Delete(spec DeleteSpec) (client.APIResponse, error) {
	storageUUIDs := splitCSV(spec.IDs)
	if spec.ID != "" {
		storageUUIDs = append([]string{spec.ID}, storageUUIDs...)
	}
	if len(storageUUIDs) == 0 {
		return client.APIResponse{}, fmt.Errorf("id or ids is required")
	}
	return s.api.Post("/hypermotion/v1/storages/action", map[string]interface{}{
		"delete_storage": map[string]interface{}{
			"storage_uuids": storageUUIDs,
			"force":         spec.Force,
		},
	})
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
