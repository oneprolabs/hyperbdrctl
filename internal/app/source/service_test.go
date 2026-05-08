package source

import (
	"net/url"
	"testing"

	"hyperbdr-client/internal/client"
)

type fakeAPI struct {
	getPath  string
	getQuery url.Values
}

func (f *fakeAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.getPath = path
	f.getQuery = q
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func TestServiceVMsBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.VMs(VMsSpec{
		ConnectionType: "vmware",
		Registered:     "0",
		Page:           1,
		PageSize:       10,
		Query:          url.Values{"debug": []string{"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/hypermotion/v1/sources/vms" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQuery.Get("connection_type") != "vmware" || api.getQuery.Get("registered") != "0" || api.getQuery.Get("page_size") != "10" || api.getQuery.Get("debug") != "1" {
		t.Fatalf("query = %+v", api.getQuery)
	}
}

func TestServiceVMsNormalizesVSphereToVMware(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.VMs(VMsSpec{
		ConnectionType: "vsphere",
		Page:           1,
		PageSize:       10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getQuery.Get("connection_type") != "vmware" {
		t.Fatalf("connection_type = %q", api.getQuery.Get("connection_type"))
	}
}

func TestServiceListBuildsAgentlessVerificationQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.List(ListSpec{
		SourceType:    "vsphere",
		KW:            "test-vm",
		BindingStatus: "binding",
		Page:          1,
		PageSize:      10,
		Query:         url.Values{"custom_step": []string{"3"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/hypermotion/v1/sources" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQuery.Get("type") != "vmware" || api.getQuery.Get("kw") != "test-vm" || api.getQuery.Get("binding_status") != "binding" || api.getQuery.Get("page") != "1" || api.getQuery.Get("page_size") != "10" || api.getQuery.Get("custom_step") != "3" {
		t.Fatalf("query = %+v", api.getQuery)
	}
}

func TestServiceAgentInstallBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.AgentInstall(AgentInstallSpec{
		Query: url.Values{"custom_step": []string{"3"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/hypermotion/v1/sources" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQuery.Get("type") != "agent" || api.getQuery.Get("custom_step") != "3" {
		t.Fatalf("query = %+v", api.getQuery)
	}
}

func TestServiceAgentlessInstallBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.AgentlessInstall(AgentlessInstallSpec{
		Query: url.Values{"custom_step": []string{"3"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/hypermotion/v1/sources" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQuery.Get("type") != "agentless" || api.getQuery.Get("custom_step") != "3" {
		t.Fatalf("query = %+v", api.getQuery)
	}
}

func TestServiceSynchNodesBuildsQuery(t *testing.T) {
	api := &fakeAPI{}
	service := NewService(api)

	_, err := service.SynchNodes(SynchNodesSpec{
		Type:   "proxy",
		Status: "online",
		Query:  url.Values{"custom_step": []string{"3"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.getPath != "/hypermotion/v1/synch_nodes" {
		t.Fatalf("path = %q", api.getPath)
	}
	if api.getQuery.Get("type") != "proxy" || api.getQuery.Get("status") != "online" || api.getQuery.Get("custom_step") != "3" {
		t.Fatalf("query = %+v", api.getQuery)
	}
}
