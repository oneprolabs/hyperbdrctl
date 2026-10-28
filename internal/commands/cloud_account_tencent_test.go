package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"hyperbdr-client/internal/i18n"
)

func tencentCreateArgs() []string {
	return []string{"cloud-account", "create", "--cloud-type", "tencent", "--storage-type", "block", "--access-key-id", "test-ak", "--access-key-secret", "test-sk", "--app-id", "1000000001", "--auth-region-id", "ap-beijing"}
}

func TestTencentCloudAccountCreatePreviewAndAcceptedResponse(t *testing.T) {
	setUserDirs(t, t.TempDir())
	t.Setenv("HYPERBDR_SCENE", "migration")
	posts := 0
	var posted map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		if r.Method != "POST" || r.URL.Path != "/hypermotion/v1/cloud_accounts" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
			return
		}
		if r.Header.Get("X-SCENE") != "migration" || r.Header.Get("X-LANG") != "zh" {
			t.Errorf("headers = %v", r.Header)
		}
		if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"cloud_account":{"uuid":"account-tencent","cloud_type":"tencent_bs","storage_type":"HyperGate","status":"available","account_name":"test-account"}}`))
	}))
	defer srv.Close()
	args := append([]string{"--lang", "zh_cn", "--output", "json"}, tencentCreateArgs()...)
	args = append(args, "--account-name", "test-account", "--set", "region_type_list_name=华北地区(北京)")
	var out, errOut bytes.Buffer
	previewArgs := append(append([]string{}, args...), "--preview-request")
	if err := Execute(withHost(t, srv.URL, previewArgs...), &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if posts != 0 {
		t.Fatalf("preview sent %d requests", posts)
	}
	var preview map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := Execute(withHost(t, srv.URL, args...), &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if posts != 1 || !reflect.DeepEqual(posted, preview) {
		t.Fatalf("posts=%d posted=%#v preview=%#v", posts, posted, preview)
	}
	account := posted["cloud_account"].(map[string]interface{})
	metadata := account["metadata"].(map[string]interface{})
	if account["cloud_type"] != "tencent_bs" || account["cloud_auth_type"] != "aksk" || account["storage_type"] != nil || metadata["app_id"] != "1000000001" || metadata["auth_region_id"] != "ap-beijing" {
		t.Fatalf("account = %#v", account)
	}
	if _, ok := account["storage_type"]; !ok {
		t.Fatal("storage_type null was omitted")
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	response := raw["cloud_account"].(map[string]interface{})
	if response["uuid"] != "account-tencent" || response["cloud_type"] != "tencent_bs" || response["account_name"] != "test-account" {
		t.Fatalf("raw response = %#v", raw)
	}
}

func TestTencentCloudAccountMetadataPrecedence(t *testing.T) {
	setUserDirs(t, t.TempDir())
	file := filepath.Join(t.TempDir(), "metadata.json")
	if err := os.WriteFile(file, []byte(`{"access_key_id":"file-ak","access_key_secret":"file-sk","app_id":1000000002,"auth_region_id":"ap-guangzhou","account_name":"file-name"}`), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"cloud-account", "create", "--cloud-type", "tencent", "--storage-type", "block", "--file", file, "--set-json", "app_id=1000000003", "--set", "app_id=1000000004", "--app-id", "1000000005", "--auth-region-id", "ap-beijing", "--account-name", "flag-name", "--preview-request"}
	var out, errOut bytes.Buffer
	if err := Execute(args, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	metadata := body["cloud_account"].(map[string]interface{})["metadata"].(map[string]interface{})
	if metadata["app_id"] != "1000000005" || metadata["access_key_id"] != "file-ak" || metadata["auth_region_id"] != "ap-beijing" || metadata["region_type_list"] != "ap-beijing" || metadata["account_name"] != "flag-name" {
		t.Fatalf("metadata = %#v", metadata)
	}
}

func TestTencentCloudAccountRejectsInvalidArgumentsBeforeRequest(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"missing App ID", []string{"--access-key-id", "test-ak", "--access-key-secret", "test-sk", "--auth-region-id", "ap-beijing"}, "app-id is required"},
		{"missing region", []string{"--access-key-id", "test-ak", "--access-key-secret", "test-sk", "--app-id", "1000000001"}, "auth-region-id is required"},
		{"password", []string{"--password", "pw"}, "unknown flag: --password"},
		{"username", []string{"--username", "user"}, "unknown flag: --username"},
		{"auth URL", []string{"--auth-url", "https://example.invalid"}, "unknown flag: --auth-url"},
		{"auth override", []string{"--cloud-auth-type", "password"}, "cloud-auth-type cannot be used"},
		{"region flag", []string{"--region-id", "ap-beijing"}, "use --auth-region-id"},
		{"region name flag", []string{"--region-name", "Beijing"}, "use --auth-region-id"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setUserDirs(t, t.TempDir())
			args := append([]string{"cloud-account", "create", "--cloud-type", "tencent", "--storage-type", "block"}, tc.args...)
			var out, errOut bytes.Buffer
			err := Execute(withHost(t, "https://example.invalid", args...), &out, &errOut)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestTencentCloudAccountHelpIsLocalizedAndOffline(t *testing.T) {
	for _, lang := range []string{"en", "zh_cn", "ja"} {
		t.Run(lang, func(t *testing.T) {
			setUserDirs(t, t.TempDir())
			var out, errOut bytes.Buffer
			if err := Execute([]string{"--lang", lang, "cloud-account", "create", "--cloud-type", "tencent", "--storage-type", "block", "--help"}, &out, &errOut); err != nil {
				t.Fatal(err)
			}
			text := out.String()
			for _, want := range []string{"--access-key-id string", "--access-key-secret string", "--app-id string", "--auth-region-id string", "ap-beijing", "--preview-request", "--set-json"} {
				if !strings.Contains(text, want) {
					t.Fatalf("missing %q in %s help: %s", want, lang, text)
				}
			}
			for _, unwanted := range []string{"--region-id string", "--region-name string", "--password string", "--cloud-auth-type string", "flag.app-id", "help.cloud_account.block.tencent", "%s"} {
				if strings.Contains(text, unwanted) {
					t.Fatalf("unexpected %q in help: %s", unwanted, text)
				}
			}
			// Help describes inputs and operations, not implementation status or scope disclaimers.
			description := i18n.New(lang).T("cmd.cloud_accounts.create.block.tencent.long")
			for _, unwanted := range []string{
				"validated", "已验证", "検証済み", "tencent_bs",
				"cloud sync gateway", "云同步网关", "クラウド同期ゲートウェイ",
				"sent as a string", "提交请求时", "文字列として送信",
				"Positive decimal integer", "正十进制整数", "正の十進整数",
				"Region metadata", "地域元数据", "メタデータを使用",
			} {
				if strings.Contains(text+description, unwanted) {
					t.Fatalf("%s help contains non-operational text %q", lang, unwanted)
				}
			}
			if lang == "zh_cn" {
				for _, want := range []string{"腾讯云 App ID", "创建腾讯云块存储云账号。", "可按需补充以下动态参数", "云账号名称", "创建成功后，等待任务完成", "查看云账号详情"} {
					if !strings.Contains(text, want) {
						t.Fatalf("Chinese help missing %q: %s", want, text)
					}
				}
				if strings.Count(text, "创建腾讯云块存储云账号") != 2 {
					t.Fatalf("Chinese help repeats the command description: %s", text)
				}
			}
		})
	}
}

func TestTencentCloudAccountRegionDiscoveryIncludesAppID(t *testing.T) {
	setUserDirs(t, t.TempDir())
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != "POST" || r.URL.Path != "/api/v3/postCloudInfoForAuth" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
			return
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		account := body["cloud_account"].(map[string]interface{})
		metadata := account["metadata"].(map[string]interface{})
		if account["cloud_type"] != "tencent_bs" || account["storage_type"] != "HyperGate" || metadata["app_id"] != "1000000001" || body["fetch_res"] != "regions" {
			t.Errorf("body = %#v", body)
		}
		_, _ = w.Write([]byte(`{"code":"00000000","data":{"cloud_info":{"regions":[{"id":"ap-beijing","name":"Beijing"}]}}}`))
	}))
	defer srv.Close()
	var out, errOut bytes.Buffer
	if err := Execute(withHost(t, srv.URL, "--output", "json", "cloud-resource", "fetch", "--cloud-type", "tencent", "--storage-type", "block", "--access-key-id", "test-ak", "--access-key-secret", "test-sk", "--app-id", "1000000001", "--fetch-res", "regions"), &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if requests != 1 || !strings.Contains(out.String(), "ap-beijing") {
		t.Fatalf("requests=%d output=%s", requests, out.String())
	}
}
