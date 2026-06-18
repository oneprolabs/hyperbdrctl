package output

import (
	"bytes"
	"strings"
	"testing"

	"hyperbdr-client/internal/i18n"
)

func TestTableAlignsWideRunes(t *testing.T) {
	var out bytes.Buffer
	rows := []map[string]interface{}{
		{"id": "1", "name": "abc", "status": "ok"},
		{"id": "2", "name": "动态盘", "status": "ok"},
	}
	err := Table(&out, i18n.New("en"), rows, []Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.status", Field: "status"},
	})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines = %q", out.String())
	}
	statusPositions := make([]int, len(lines))
	for i, line := range lines {
		idx := strings.Index(line, "Status")
		if i > 0 {
			idx = strings.Index(line, "ok")
		}
		if idx < 0 {
			t.Fatalf("missing status column in line %q", line)
		}
		statusPositions[i] = displayWidth(line[:idx])
	}
	if statusPositions[0] != statusPositions[1] || statusPositions[1] != statusPositions[2] {
		t.Fatalf("status columns are not aligned: positions=%v output=\n%s", statusPositions, out.String())
	}
}

func TestDisplayWidth(t *testing.T) {
	if got := displayWidth("动态盘"); got != 6 {
		t.Fatalf("display width = %d", got)
	}
	if got := displayWidth("abc"); got != 3 {
		t.Fatalf("display width = %d", got)
	}
}

func TestVerticalUsesColumnOrderAndNumberedBlocks(t *testing.T) {
	var out bytes.Buffer
	rows := []map[string]interface{}{
		{"id": "1", "name": "alpha", "status": "ok"},
		{"id": "2", "name": "beta", "status": "failed"},
	}
	err := Vertical(&out, i18n.New("en"), rows, []Column{
		{HeaderKey: "table.id", Field: "id"},
		{HeaderKey: "table.name", Field: "name"},
		{HeaderKey: "table.status", Field: "status"},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"*************************** 1. row ***************************",
		"*************************** 2. row ***************************",
		"ID      1",
		"Name    alpha",
		"Status  ok",
		"ID      2",
		"Name    beta",
		"Status  failed",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("vertical output missing %q: %q", want, text)
		}
	}
}

func TestVerticalAcceptsEmptyRows(t *testing.T) {
	var out bytes.Buffer
	if err := Vertical(&out, i18n.New("en"), nil, []Column{
		{HeaderKey: "table.id", Field: "id"},
	}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "" {
		t.Fatalf("expected empty output, got %q", out.String())
	}
}
