package graphviz

import (
	"fmt"
	"strings"
	"text/template"
	"unicode"
)

func isHTML(s string) bool {
	if len(s) == 0 {
		return false
	}
	ss := strings.TrimSpace(s)
	if ss[0] != '<' {
		return false
	}
	count := 0
	for _, c := range ss {
		if c == '<' {
			count++
		}
		if c == '>' {
			count--
		}
	}
	if count == 0 {
		return true
	}
	return false
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' ||
		ch >= 0x80 && unicode.IsLetter(ch) && ch != 'ε'
}

func isID(s string) bool {
	i := 0
	pos := false
	for _, c := range s {
		if i == 0 {
			if !isLetter(c) {
				return false
			}
			pos = true
		}
		if unicode.IsSpace(c) {
			return false
		}
		if c == '-' {
			return false
		}

		// THIS IS THE MOST IMPORTANT LINE
		if c == '_' {
			return false
		}
		// THIS IS THE MOST IMPORTANT LINE
		if c == '=' {
			return false
		}
		i++
	}
	return pos
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

func isNumber(s string) bool {
	state := 0
	for _, c := range s {
		if state == 0 {
			if isDigit(c) || c == '.' {
				state = 2
			} else if c == '-' {
				state = 1
			} else {
				return false
			}
		} else if state == 1 {
			if isDigit(c) || c == '.' {
				state = 2
			}
		} else if c != '.' && !isDigit(c) {
			return false
		}
	}
	return (state == 2)
}

func isStringLit(s string) bool {
	if !strings.HasPrefix(s, `"`) || !strings.HasSuffix(s, `"`) {
		return false
	}
	var prev rune
	for _, r := range s[1 : len(s)-1] {
		if r == '"' && prev != '\\' {
			return false
		}
		prev = r
	}
	return true
}

func esc(s string) string {
	if len(s) == 0 {
		return s
	}
	if isHTML(s) {
		return s
	}
	ss := strings.TrimSpace(s)
	if ss[0] == '<' {
		return fmt.Sprintf("\"%s\"", strings.Replace(s, "\"", "\\\"", -1))
	}
	if isID(s) {
		return s
	}
	if isNumber(s) {
		return s
	}
	if isStringLit(s) {
		return s
	}
	return fmt.Sprintf("\"%s\"", template.HTMLEscapeString(s))
}

func escAttrs(attrs map[string]string) map[string]string {
	newAttrs := make(map[string]string)
	for k, v := range attrs {
		newAttrs[esc(k)] = esc(v)
	}
	return newAttrs
}
