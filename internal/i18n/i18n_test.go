package i18n

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

var formatPlaceholderPattern = regexp.MustCompile(`%[-+0-9.#]*[a-zA-Z]`)

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
	if len(en) != 1419 {
		t.Fatalf("english catalog entries = %d, want 1419", len(en))
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
		if ja[key] == "" {
			t.Fatalf("ja key %q has empty or missing translation", key)
		}
	}
	if len(ja) != len(en) {
		t.Fatalf("ja catalog entries = %d, want %d", len(ja), len(en))
	}
	if got := New("ja").T("config.saved"); got != "設定を保存しました" {
		t.Fatalf("ja config.saved = %q", got)
	}
	for key, english := range en {
		jaPlaceholders := formatPlaceholderPattern.FindAllString(ja[key], -1)
		enPlaceholders := formatPlaceholderPattern.FindAllString(english, -1)
		sort.Strings(jaPlaceholders)
		sort.Strings(enPlaceholders)
		if strings.Join(jaPlaceholders, "\x00") != strings.Join(enPlaceholders, "\x00") {
			t.Fatalf("ja placeholders for %q = %v, want %v", key, jaPlaceholders, enPlaceholders)
		}
	}
}

func TestJapaneseFragmentsCoverEachBaseFragment(t *testing.T) {
	if len(japaneseLocaleFragments) != len(localeFragments) {
		t.Fatalf("japanese fragment count = %d, want %d", len(japaneseLocaleFragments), len(localeFragments))
	}
	if len(japaneseCatalog) != len(en) {
		t.Fatalf("Japanese source catalog entries = %d, want %d", len(japaneseCatalog), len(en))
	}
	for i, base := range localeFragments {
		japanese := japaneseLocaleFragments[i]
		if japanese.name != base.name {
			t.Fatalf("japanese fragment %d name = %q, want %q", i, japanese.name, base.name)
		}
		if len(japanese.ja) != len(base.en) {
			t.Fatalf("japanese fragment %q entries = %d, want %d", base.name, len(japanese.ja), len(base.en))
		}
		for key := range base.en {
			if japaneseCatalog[key] == "" {
				t.Fatalf("Japanese source catalog has no translation for key %q", key)
			}
			if japanese.ja[key] == "" {
				t.Fatalf("Japanese fragment %q has no translation for key %q", base.name, key)
			}
		}
	}
}

func TestJapaneseCatalogHasNoKnownEnglishHelpPhrases(t *testing.T) {
	for english := range japaneseTextCorrections {
		for key, value := range ja {
			if strings.Contains(value, english) {
				t.Fatalf("ja key %q still contains English help text %q", key, english)
			}
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

func TestBuildJapaneseLocaleCatalogRejectsInvalidFragments(t *testing.T) {
	tests := []struct {
		name      string
		fragments []localeFragment
		want      string
	}{
		{
			name: "missing ja",
			fragments: []localeFragment{
				{name: "one", en: map[string]string{"missing": "value"}, ja: map[string]string{}},
			},
			want: "missing ja translation",
		},
		{
			name: "empty ja",
			fragments: []localeFragment{
				{name: "one", en: map[string]string{"empty": "value"}, ja: map[string]string{"empty": ""}},
			},
			want: "empty ja translation",
		},
		{
			name: "unexpected ja key",
			fragments: []localeFragment{
				{name: "one", en: map[string]string{}, ja: map[string]string{"extra": "余分"}},
			},
			want: "ja translation without en key",
		},
		{
			name: "duplicate ja key",
			fragments: []localeFragment{
				{name: "one", en: map[string]string{"same": "one"}, ja: map[string]string{"same": "一"}},
				{name: "two", en: map[string]string{"same": "two"}, ja: map[string]string{"same": "二"}},
			},
			want: "duplicate ja translation key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildJapaneseLocaleCatalog(tt.fragments)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}
