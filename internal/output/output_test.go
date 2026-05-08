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
	err := Table(&out, i18n.New("zh_cn"), rows, []Column{
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
		idx := strings.Index(line, "状态")
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
