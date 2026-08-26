package i18n

import (
	"strings"
	"testing"
)

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

func TestLocaleCatalogCompleteness(t *testing.T) {
	if len(en) != 1391 {
		t.Fatalf("english catalog entries = %d, want 1391", len(en))
	}
	if len(zhCN) != len(en) {
		t.Fatalf("zh_cn catalog entries = %d, want %d", len(zhCN), len(en))
	}
	for key, value := range en {
		if value == "" {
			t.Fatalf("english key %q has empty translation", key)
		}
		if zhCN[key] == "" {
			t.Fatalf("zh_cn key %q has empty or missing translation", key)
		}
	}
}

func TestBuildLocaleCatalogRejectsInvalidFragments(t *testing.T) {
	tests := []struct {
		name      string
		fragments []localeFragment
		want      string
	}{
		{
			name: "duplicate key",
			fragments: []localeFragment{
				{name: "one", en: map[string]string{"same": "one"}, zh: map[string]string{"same": "一"}},
				{name: "two", en: map[string]string{"same": "two"}, zh: map[string]string{"same": "二"}},
			},
			want: "duplicate en translation key",
		},
		{
			name: "missing zh",
			fragments: []localeFragment{
				{name: "one", en: map[string]string{"missing": "value"}, zh: map[string]string{}},
			},
			want: "missing zh_cn translation",
		},
		{
			name: "missing en",
			fragments: []localeFragment{
				{name: "one", en: map[string]string{}, zh: map[string]string{"missing": "值"}},
			},
			want: "missing en translation",
		},
		{
			name: "empty en",
			fragments: []localeFragment{
				{name: "one", en: map[string]string{"empty": ""}, zh: map[string]string{"empty": "空"}},
			},
			want: "empty en translation",
		},
		{
			name: "empty zh",
			fragments: []localeFragment{
				{name: "one", en: map[string]string{"empty": "value"}, zh: map[string]string{"empty": ""}},
			},
			want: "empty zh_cn translation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := buildLocaleCatalog(tt.fragments)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}
