package cloudaccount

import (
	"errors"

	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/workflow"
	workflowcreate "hyperbdr-client/internal/workflow/cloudaccountcreate"
)

var (
	errRawCreateBodyRequired        = errors.New("cloud account create body is required")
	errRawCreateStorageTypeRequired = errors.New("cloud account create body storage_type is required")
	errRawCreateStorageTypeInvalid  = errors.New("cloud account create body storage_type must be HyperGate, block, or objectstorage")
)

type CreateSpec = workflowcreate.Spec

type CreateRawSpec struct {
	Body map[string]interface{}
}

type PreparedCreateRequest struct {
	Path string
	Body map[string]interface{}
}

func (s Service) Create(spec CreateSpec) (client.APIResponse, error) {
	prepared, err := s.PrepareCreate(spec)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Post(prepared.Path, prepared.Body)
}

func (s Service) PrepareCreate(spec CreateSpec) (PreparedCreateRequest, error) {
	path, reqBody, err := workflowcreate.BuildRequest(spec)
	if err != nil {
		return PreparedCreateRequest{}, err
	}
	return PreparedCreateRequest{
		Path: path,
		Body: reqBody,
	}, nil
}

func (s Service) CreateRaw(spec CreateRawSpec) (client.APIResponse, error) {
	prepared, err := s.PrepareCreateRaw(spec)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Post(prepared.Path, prepared.Body)
}

func (s Service) PrepareCreateRaw(spec CreateRawSpec) (PreparedCreateRequest, error) {
	if spec.Body == nil {
		return PreparedCreateRequest{}, errRawCreateBodyRequired
	}

	if _, err := rawCreateStorageType(spec.Body); err != nil {
		return PreparedCreateRequest{}, err
	}

	path := "/hypermotion/v1/cloud_accounts"

	return PreparedCreateRequest{
		Path: path,
		Body: spec.Body,
	}, nil
}

func rawCreateStorageType(body map[string]interface{}) (string, error) {
	storageType, ok := nestedStorageType(body)
	if !ok {
		return "", errRawCreateStorageTypeRequired
	}

	switch workflow.CloudAccountCreateKey("", storageType).StorageType {
	case workflow.BlockStorageType:
		return workflow.BlockStorageType, nil
	case "objectstorage":
		return "objectstorage", nil
	default:
		return "", errRawCreateStorageTypeInvalid
	}
}

func nestedStorageType(body map[string]interface{}) (string, bool) {
	if cloudAccountValue, ok := body["cloud_account"]; ok {
		cloudAccount, ok := cloudAccountValue.(map[string]interface{})
		if !ok {
			return "", false
		}
		if value, ok := cloudAccount["storage_type"].(string); ok && value != "" {
			return value, true
		}
	}
	if value, ok := body["storage_type"].(string); ok && value != "" {
		return value, true
	}
	return "", false
}
