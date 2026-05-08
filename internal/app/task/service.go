package task

import (
	"fmt"
	"net/url"

	"hyperbdr-client/internal/client"
)

type API interface {
	Get(path string, q url.Values) (client.APIResponse, error)
}

type Service struct {
	api API
}

func NewService(api API) Service {
	return Service{api: api}
}

type ListSpec struct {
	SourceID string
	Page     int
	PageSize int
	Query    url.Values
}

type StepsSpec struct {
	TaskID string
	Query  url.Values
}

func (s Service) List(spec ListSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	q.Set("source_id", spec.SourceID)
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/api/v2/getTasks", q)
}

func (s Service) Steps(spec StepsSpec) (client.APIResponse, error) {
	if spec.TaskID == "" {
		return client.APIResponse{}, fmt.Errorf("task-id is required")
	}
	q := cloneValues(spec.Query)
	q.Set("task_id", spec.TaskID)
	return s.api.Get("/hypermotion/v1/job/steps", q)
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
