package license

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

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
	File string
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

func (s Service) ActivateBody(spec ActivateSpec) (map[string]string, error) {
	body := map[string]string{}
	if spec.File != "" {
		b, err := os.ReadFile(spec.File)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &body); err != nil {
			return nil, err
		}
	}
	if spec.KKTY != "" {
		body["kkty"] = spec.KKTY
	}
	if spec.DDTY != "" {
		body["ddty"] = spec.DDTY
	}
	return body, nil
}

func (s Service) Activate(body map[string]string) (client.APIResponse, error) {
	if body["kkty"] == "" {
		return client.APIResponse{}, fmt.Errorf("kkty is required")
	}
	if body["ddty"] == "" {
		return client.APIResponse{}, fmt.Errorf("ddty is required")
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
