package source

import (
	"fmt"
	"net/url"

	"hyperbdr-client/internal/client"
)

type API interface {
	Get(path string, q url.Values) (client.APIResponse, error)
	Delete(path string, body interface{}) (client.APIResponse, error)
}

type Service struct {
	api API
}

func NewService(api API) Service {
	return Service{api: api}
}

type ListSpec struct {
	SourceType    string
	KW            string
	BindingStatus string
	Page          int
	PageSize      int
	Query         url.Values
}

type DetailSpec struct {
	ID            string
	SourceType    string
	BindingStatus string
	Page          int
	PageSize      int
	Query         url.Values
}

type VMsSpec struct {
	ConnectionType string
	ConnectionUUID string
	Registered     string
	KW             string
	Page           int
	PageSize       int
	Query          url.Values
}

type AgentlessInstallSpec struct {
	Query url.Values
}

type AgentInstallSpec struct {
	Query url.Values
}

type SynchNodesSpec struct {
	Type   string
	Status string
	Query  url.Values
}

type DeleteSpec struct {
	ID    string
	Force bool
}

func (s Service) List(spec ListSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	if sourceType := normalizeReadSourceType(spec.SourceType); sourceType != "" {
		q.Set("type", sourceType)
	}
	if spec.KW != "" {
		q.Set("kw", spec.KW)
	}
	if spec.BindingStatus != "" {
		q.Set("binding_status", spec.BindingStatus)
	}
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/hypermotion/v1/sources", q)
}

func (s Service) Detail(spec DetailSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	q := cloneValues(spec.Query)
	q.Set("connection_id", spec.ID)
	if sourceType := normalizeReadSourceType(spec.SourceType); sourceType != "" {
		q.Set("type", sourceType)
	}
	if spec.BindingStatus != "" {
		q.Set("binding_status", spec.BindingStatus)
	}
	addInt(q, "page_num", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/api/v2/getConnectionDetail", q)
}

func (s Service) VMs(spec VMsSpec) (client.APIResponse, error) {
	if spec.ConnectionType == "" {
		return client.APIResponse{}, fmt.Errorf("connection-type is required")
	}
	q := cloneValues(spec.Query)
	q.Set("connection_type", normalizeReadSourceType(spec.ConnectionType))
	if spec.ConnectionUUID != "" {
		q.Set("connection_uuid", spec.ConnectionUUID)
	}
	if spec.Registered != "" {
		q.Set("registered", spec.Registered)
	}
	if spec.KW != "" {
		q.Set("kw", spec.KW)
	}
	addInt(q, "page", spec.Page)
	addInt(q, "page_size", spec.PageSize)
	return s.api.Get("/hypermotion/v1/sources/vms", q)
}

func (s Service) AgentlessInstall(spec AgentlessInstallSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	q.Set("type", "agentless")
	return s.api.Get("/hypermotion/v1/sources", q)
}

func (s Service) AgentInstall(spec AgentInstallSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	q.Set("type", "agent")
	return s.api.Get("/hypermotion/v1/sources", q)
}

func (s Service) SynchNodes(spec SynchNodesSpec) (client.APIResponse, error) {
	q := cloneValues(spec.Query)
	if spec.Type != "" {
		q.Set("type", spec.Type)
	}
	if spec.Status != "" {
		q.Set("status", spec.Status)
	}
	return s.api.Get("/hypermotion/v1/synch_nodes", q)
}

func (s Service) Delete(spec DeleteSpec) (client.APIResponse, error) {
	if spec.ID == "" {
		return client.APIResponse{}, fmt.Errorf("id is required")
	}
	path := "/hypermotion/v1/sources/" + url.PathEscape(spec.ID)
	if spec.Force {
		path += "?force=true"
	}
	return s.api.Delete(path, nil)
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

func normalizeReadSourceType(sourceType string) string {
	switch sourceType {
	case "vsphere":
		return "vmware"
	default:
		return sourceType
	}
}
