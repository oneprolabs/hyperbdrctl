package commands

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func writeSectionTitle(ctx *context, title string) error {
	line := fmt.Sprintf("== %s ==", title)
	if _, err := fmt.Fprintln(ctx.out, line); err != nil {
		return err
	}
	width := utf8.RuneCountInString(line) * 2
	if width < 25 {
		width = 25
	}
	_, err := fmt.Fprintln(ctx.out, strings.Repeat("-", width))
	return err
}
