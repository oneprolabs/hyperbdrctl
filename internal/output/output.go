package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"hyperbdr-client/internal/i18n"
)

type Column struct {
	HeaderKey string
	Field     string
}

type KeyValueRow struct {
	Key   string
	Value interface{}
}

func JSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func Table(w io.Writer, loc i18n.Localizer, rows []map[string]interface{}, cols []Column) error {
	headers := make([]string, len(cols))
	widths := make([]int, len(cols))
	for i, col := range cols {
		headers[i] = loc.T(col.HeaderKey)
		widths[i] = displayWidth(headers[i])
	}
	values := make([][]string, 0, len(rows))
	for _, row := range rows {
		line := make([]string, len(cols))
		for i, col := range cols {
			line[i] = scalar(row[col.Field])
			if width := displayWidth(line[i]); width > widths[i] {
				widths[i] = width
			}
		}
		values = append(values, line)
	}
	if err := writeTableLine(w, headers, widths); err != nil {
		return err
	}
	for _, line := range values {
		if err := writeTableLine(w, line, widths); err != nil {
			return err
		}
	}
	return nil
}

func KeyValue(w io.Writer, loc i18n.Localizer, m map[string]interface{}) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	width := 0
	for _, k := range keys {
		if w := displayWidth(k); w > width {
			width = w
		}
	}
	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "%s%s  %s\n", k, strings.Repeat(" ", width-displayWidth(k)), scalar(m[k])); err != nil {
			return err
		}
	}
	return nil
}

func OrderedKeyValue(w io.Writer, loc i18n.Localizer, rows []KeyValueRow) error {
	width := 0
	for _, row := range rows {
		if w := displayWidth(row.Key); w > width {
			width = w
		}
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "%s%s  %s\n", row.Key, strings.Repeat(" ", width-displayWidth(row.Key)), scalar(row.Value)); err != nil {
			return err
		}
	}
	return nil
}

func writeTableLine(w io.Writer, values []string, widths []int) error {
	for i, value := range values {
		if i > 0 {
			if _, err := io.WriteString(w, "  "); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, value); err != nil {
			return err
		}
		if i < len(values)-1 {
			padding := widths[i] - displayWidth(value)
			if padding > 0 {
				if _, err := io.WriteString(w, strings.Repeat(" ", padding)); err != nil {
					return err
				}
			}
		}
	}
	_, err := io.WriteString(w, "\n")
	return err
}

func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r == '\t' || r == '\n' || r == '\r' {
			width++
			continue
		}
		if isWideRune(r) {
			width += 2
		} else {
			width++
		}
	}
	return width
}

func isWideRune(r rune) bool {
	return unicode.In(r,
		unicode.Han,
		unicode.Hangul,
		unicode.Hiragana,
		unicode.Katakana,
	) || (r >= 0xFF01 && r <= 0xFF60) || (r >= 0xFFE0 && r <= 0xFFE6)
}

func scalar(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case bool:
		return fmt.Sprintf("%t", t)
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return string(b)
	}
}
