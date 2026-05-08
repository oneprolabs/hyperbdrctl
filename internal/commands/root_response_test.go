package commands

import (
	"bytes"
	"strings"
	"testing"

	"hyperbdr-client/internal/client"
	"hyperbdr-client/internal/config"
	"hyperbdr-client/internal/i18n"
)

func TestWriteResponseUnwrapsSingleNestedObjectForHumanOutput(t *testing.T) {
	var out bytes.Buffer
	ctx := &context{
		out: &out,
		cfg: config.Resolved{Config: config.Config{Output: ""}},
		loc: i18n.New("en"),
	}

	resp := client.APIResponse{
		Data: map[string]interface{}{
			"cloud_account": map[string]interface{}{
				"id":   "account-1",
				"name": "my-aliyun-bs-account",
			},
		},
	}

	if err := writeResponse(ctx, resp, "", nil); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if strings.Contains(text, "cloud_account") {
		t.Fatalf("output should unwrap wrapper key: %q", text)
	}
	for _, want := range []string{"{", "\"id\": \"account-1\"", "\"name\": \"my-aliyun-bs-account\""} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestWriteResponseFormatsNestedMapAsIndentedJSONForHumanOutput(t *testing.T) {
	var out bytes.Buffer
	ctx := &context{
		out: &out,
		cfg: config.Resolved{Config: config.Config{Output: ""}},
		loc: i18n.New("en"),
	}

	resp := client.APIResponse{
		Data: map[string]interface{}{
			"id":   "account-1",
			"name": "my-aliyun-bs-account",
			"storage_statistics": map[string]interface{}{
				"total_storage_num": float64(0),
			},
		},
	}

	if err := writeResponse(ctx, resp, "", nil); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if strings.Contains(text, "storage_statistics  ") {
		t.Fatalf("output should not use inline key-value JSON: %q", text)
	}
	for _, want := range []string{"{", "\"storage_statistics\": {", "  \"id\": \"account-1\"", "    \"total_storage_num\": 0"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}

func TestWriteValueKeepsScalarMapAsKeyValueOutput(t *testing.T) {
	var out bytes.Buffer
	ctx := &context{
		out: &out,
		cfg: config.Resolved{Config: config.Config{Output: ""}},
		loc: i18n.New("en"),
	}

	if err := writeValue(ctx, map[string]interface{}{
		"host":     "https://example.invalid",
		"insecure": true,
	}); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if strings.Contains(text, "{") {
		t.Fatalf("scalar-only maps should stay in key-value form: %q", text)
	}
	for _, want := range []string{"host", "https://example.invalid", "insecure", "true"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output = %q, missing %q", text, want)
		}
	}
}
