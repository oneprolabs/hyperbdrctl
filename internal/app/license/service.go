package license

import (
	"fmt"
	"net/url"

	"hyperbdr-client/internal/client"
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

type ListSpec struct {
	Page     int
	PageSize int
	Query    url.Values
}

type RegCodeSpec struct {
	Query url.Values
}

type ActivateSpec struct {
	KKTY string
	DDTY string
}

func (s Service) List(spec ListSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/api/v2/getLicenses", q)
}

func (s Service) RegCode(spec RegCodeSpec) (client.APIResponse, error) {
	return s.api.Get("/api/v2/getLicenseRegCode", cloneValues(spec.Query))
}

func (s Service) Activate(spec ActivateSpec) (client.APIResponse, error) {
	if spec.KKTY == "" {
		return client.APIResponse{}, fmt.Errorf("kkty is required")
	}
	if spec.DDTY == "" {
		return client.APIResponse{}, fmt.Errorf("ddty is required")
	}
	body := map[string]string{
		"kkty": spec.KKTY,
		"ddty": spec.DDTY,
	}
	return s.api.Post("/api/v2/activateLicense", body)
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
