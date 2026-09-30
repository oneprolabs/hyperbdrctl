package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestJapaneseProviderSpecificCreateHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "cloud account aliyun block",
			args: []string{"--lang", "ja", "cloud-account", "create", "--cloud-type", "aliyun", "--storage-type", "block", "--help"},
			want: "Alibaba Cloud ブロックストレージ クラウド アカウントを作成します。",
		},
		{
			name: "cloud account huawei object",
			args: []string{"--lang", "ja", "cloud-account", "create", "--cloud-type", "huawei", "--storage-type", "object", "--help"},
			want: "Huawei Cloud オブジェクトストレージクラウドアカウントを作成します。",
		},
		{
			name: "cloud sync gateway aliyun",
			args: []string{"--lang", "ja", "cloud-sync-gateway", "create", "--cloud-type", "aliyun", "--help"},
			want: "Alibaba Cloud クラウド同期ゲートウェイを作成します。",
		},
		{
			name: "cloud sync gateway openstack",
			args: []string{"--lang", "ja", "cloud-sync-gateway", "create", "--cloud-type", "openstack", "--help"},
			want: "OpenStack クラウド同期ゲートウェイを作成します。",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if err := Execute(tt.args, &out, &errOut); err != nil {
				t.Fatalf("Execute() error = %v, stderr = %s", err, errOut.String())
			}
			if !strings.Contains(out.String(), tt.want) {
				t.Fatalf("help is missing Japanese text %q:\n%s", tt.want, out.String())
			}
			if strings.Contains(out.String(), "Parameter Sources:") || strings.Contains(out.String(), "Resource Discovery:") {
				t.Fatalf("provider-specific help contains an English section heading:\n%s", out.String())
			}
		})
	}
}
