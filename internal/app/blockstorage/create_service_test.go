package blockstorage

import (
	"net/url"
	"testing"

	"hyperbdr-client/internal/client"
)

type createFakeAPI struct {
	getPath          string
	sawAccountDetail bool
	postCalls        int
	postPath         string
	postBody         interface{}
}

func (f *createFakeAPI) Get(path string, q url.Values) (client.APIResponse, error) {
	f.getPath = path
	if path == "/hypermotion/v1/cloud_accounts/account-1" {
		f.sawAccountDetail = true
		return client.APIResponse{
			Raw: map[string]interface{}{
				"cloud_account": map[string]interface{}{
					"cloud_type":       "aliyun_bs",
					"region_type_list": "cn-beijing",
					"auth_region_id":   "cn-beijing",
				},
			},
		}, nil
	}
	if path == "/api/v3/getCloudInfo" {
		return client.APIResponse{Data: map[string]interface{}{
			"cloud_info": map[string]interface{}{
				"images": []interface{}{
					map[string]interface{}{"image_id": "boot-img-1", "image_name": "Windows 2016"},
				},
			},
		}}, nil
	}
	return client.APIResponse{Data: map[string]interface{}{"ok": true}}, nil
}

func (f *createFakeAPI) Post(path string, body interface{}) (client.APIResponse, error) {
	f.postCalls++
	f.postPath = path
	f.postBody = body
	if path == "/hypermotion/v1/cloud_accounts/account-1/action" {
		switch f.postCalls {
		case 1:
			return client.APIResponse{Data: map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"domain": map[string]interface{}{
						"regions": []interface{}{
							map[string]interface{}{"region_id": "cn-beijing", "region_name": "North China 2 (Beijing)"},
						},
					},
					"zones": []interface{}{
						map[string]interface{}{"id": "cn-beijing-l", "display_name": "Beijing Zone L"},
					},
				},
			}}, nil
		case 2:
			return client.APIResponse{Data: map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"flavors": []interface{}{
						map[string]interface{}{"id": "ecs.e-c1m2.large", "name": "2C4G", "is_recommend": 1},
					},
				},
			}}, nil
		case 3:
			return client.APIResponse{Data: map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"images": []interface{}{
						map[string]interface{}{"id": "img-1", "name": "ubuntu", "os_type": "linux"},
					},
					"system_disk_types": []interface{}{
						map[string]interface{}{"id": "cloud_essd_entry"},
					},
				},
			}}, nil
		case 4:
			return client.APIResponse{Data: map[string]interface{}{
				"cloud_info": map[string]interface{}{
					"networks": []interface{}{
						map[string]interface{}{"id": "net-1", "name": "vpc-a"},
					},
					"subnets": []interface{}{
						map[string]interface{}{"id": "subnet-1", "name": "subnet-a", "network_id": "net-1"},
					},
				},
			}}, nil
		}
	}
	if path == "/hypermotion/v1/storages/action" {
		return client.APIResponse{Data: map[string]interface{}{"uuid": "storage-1"}}, nil
	}
	return client.APIResponse{Data: map[string]interface{}{"uuid": "storage-1"}}, nil
}

func TestServiceCreateUsesRawCloudAccountDetailForAliyunDefaults(t *testing.T) {
	api := &createFakeAPI{}
	service := NewService(api)

	_, err := service.Create(CreateSpec{
		CloudAccountID: "account-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !api.sawAccountDetail {
		t.Fatalf("account detail was not queried; last get path = %q", api.getPath)
	}
	if api.postPath != "/hypermotion/v1/storages/action" {
		t.Fatalf("post path = %q", api.postPath)
	}
}
