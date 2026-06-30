package cloudaccount

import (
	"fmt"
	"net/url"
	"strings"

	"hyperbdr-client/internal/client"
)

type ListSpec struct {
	Page        int
	PageSize    int
	StorageType string
	Query       url.Values
}

type DetailSpec struct {
	ID    string
	Query url.Values
}

type DeleteSpec struct {
	ID    string
	Force bool
}

func (s Service) List(spec ListSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	if strings.TrimSpace(spec.StorageType) != "" {
		q.Set("storage_type", normalizeCloudAccountStorageType(spec.StorageType))
	}
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/api/v2/getCloudAccounts", q)
}

func (s Service) Detail(spec DetailSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	return s.api.Get("/hypermotion/v1/cloud_accounts/"+url.PathEscape(spec.ID), cloneValues(spec.Query))
}

func (s Service) Delete(spec DeleteSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	path := "/hypermotion/v1/cloud_accounts/" + url.PathEscape(spec.ID)
	if spec.Force {
		path += "?force=true"
	} else {
		path += "?force=false"
	}
	return s.api.Delete(path, map[string]interface{}{
		"id": spec.ID,
	})
}

func normalizeCloudAccountStorageType(storageType string) string {
	switch strings.ToLower(strings.TrimSpace(storageType)) {
	case "", "blockstorage", "block", "hypergate":
		return "HyperGate"
	case "object", "objectstorage":
		return "objectstorage"
	default:
		return storageType
	}
}

func cloneValues(v url.Values) url.Values {
	cloned := url.Values{}
	for key, values := range v {
		for _, value := range values {
			cloned.Add(key, value)
		}
	}
	return cloned
}

func addInt(q url.Values, key string, value int) {
	if value > 0 {
		q.Set(key, fmt.Sprintf("%d", value))
	}
}
