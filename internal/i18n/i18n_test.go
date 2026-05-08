package i18n

import "testing"

func TestFallback(t *testing.T) {
	loc := New("zh_cn")
	if got := loc.T("config.saved"); got != "配置已保存" {
		t.Fatalf("zh translation = %q", got)
	}
	if got := loc.T("missing.key"); got != "missing.key" {
		t.Fatalf("missing fallback = %q", got)
	}
	if got := New("bad").Lang(); got != "en" {
		t.Fatalf("bad lang fallback = %q", got)
	}
}
