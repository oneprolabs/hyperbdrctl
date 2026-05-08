package source

import (
	"hyperbdr-client/internal/client"
	workflowcreate "hyperbdr-client/internal/workflow/sourcecreate"
)

type CreateAPI interface {
	Post(path string, body interface{}) (client.APIResponse, error)
}

type CreateService struct {
	api CreateAPI
}

type CreateSpec = workflowcreate.Spec

type PreparedCreateRequest struct {
	Path string
	Body map[string]interface{}
}

func NewCreateService(api CreateAPI) CreateService {
	return CreateService{api: api}
}

func (s CreateService) Create(spec CreateSpec) (client.APIResponse, error) {
	prepared, err := s.PrepareCreate(spec)
	if err != nil {
		return client.APIResponse{}, err
	}
	return s.api.Post(prepared.Path, prepared.Body)
}

func (s CreateService) PrepareCreate(spec CreateSpec) (PreparedCreateRequest, error) {
	path, body, err := workflowcreate.BuildRequest(spec)
	if err != nil {
		return PreparedCreateRequest{}, err
	}
	return PreparedCreateRequest{
		Path: path,
		Body: body,
	}, nil
}
