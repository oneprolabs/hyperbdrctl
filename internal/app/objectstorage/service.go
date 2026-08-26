package objectstorage

import (
	"fmt"
	"net/url"

	"hyperbdr-client/internal/client"
	workflowcreate "hyperbdr-client/internal/workflow/objectstoragecreate"
)

type API interface {
	Get(path string, q url.Values) (client.APIResponse, error)
	Post(path string, body interface{}) (client.APIResponse, error)
}

type Service struct {
	api API
}

func NewService(api API) Service {
	return Service{api: api}
}

type BucketsSpec struct {
	AuthURL         string
	RegionID        string
	AccessKeyID     string
	AccessKeySecret string
	Protocol        string
	BucketLookup    string
	UseTLS          bool
}

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

type AssociatedResourcesSpec struct {
	ID string
}

type CreateProfile = workflowcreate.Profile

type CreateSpec = workflowcreate.Spec

type PreparedCreateRequest struct {
	Path string
	Body map[string]interface{}
}

type DeleteSpec struct {
	ID    string
	Force bool
}

func (s Service) List(spec ListSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	q.Set("type", spec.StorageType)
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/api/v2/getStorages", q)
}

func (s Service) Detail(spec DetailSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	q := cloneValues(spec.Query)
	q.Set("storage_id", spec.ID)
	return s.api.Get("/api/v2/getStorageDetailInfo", q)
}

func (s Service) AssociatedResources(spec AssociatedResourcesSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	q := url.Values{}
	q.Set("storage_id", spec.ID)
	q.Set("with_statistics", "false")
	return s.api.Get("/api/v2/getStorageAssociatedResources", q)
}

func (s Service) Buckets(spec BucketsSpec) (client.APIResponse, error) {
	if spec.AuthURL == "" {
		return client.APIResponse{}, fmt.Errorf("auth-url is required")
	}
	if spec.AccessKeyID == "" {
		return client.APIResponse{}, fmt.Errorf("access-key-id is required")
	}
	if spec.AccessKeySecret == "" {
		return client.APIResponse{}, fmt.Errorf("access-key-secret is required")
	}
	if spec.Protocol == "" {
		spec.Protocol = "s3"
	}

	return s.api.Post("/api/v2/objectStorageBuckets", map[string]interface{}{
		"auth_type":     "aksk",
		"auth_url":      spec.AuthURL,
		"region_id":     spec.RegionID,
		"auth_key":      spec.AccessKeyID,
		"auth_cert":     spec.AccessKeySecret,
		"use_tls":       spec.UseTLS,
		"protocol":      spec.Protocol,
		"bucket_lookup": workflowcreate.NormalizeBucketLookup(spec.BucketLookup),
	})
}

func (s Service) Create(profile CreateProfile, spec CreateSpec) (client.APIResponse, error) {
	prepared, err := s.PrepareCreate(profile, spec)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Post(prepared.Path, prepared.Body)
}

func (s Service) Delete(spec DeleteSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	return s.api.Post("/api/v2/deleteStorage", map[string]interface{}{
		"storage_id": spec.ID,
		"force":      spec.Force,
	})
}

func (s Service) PrepareCreate(profile CreateProfile, spec CreateSpec) (PreparedCreateRequest, error) {
	prepared, err := workflowcreate.BuildRequest(profile, spec)
	if err != nil {
		return PreparedCreateRequest{}, err
	}
	return PreparedCreateRequest{Path: prepared.Path, Body: prepared.Body}, nil
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
