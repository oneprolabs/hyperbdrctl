package commands

import "testing"

func TestTaskColumns(t *testing.T) {
	columns := taskColumns()
	got := make([]string, 0, len(columns))
	for _, column := range columns {
		got = append(got, column.Field)
	}
	want := []string{
		"id",
		"display_type",
		"source",
		"source_id",
		"start_at",
		"end_at",
		"display_status",
	}
	if len(got) != len(want) {
		t.Fatalf("columns = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("columns = %v, want %v", got, want)
		}
	}
}
