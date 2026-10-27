package gen

import "strings"

// tableName 决定表名：显式 table 优先，否则消息名转蛇形再变复数。
func tableName(msgName, override string) string {
	if override != "" {
		return override
	}
	return pluralize(snake(msgName))
}

// snake 把大驼峰（含连续大写缩写）转成蛇形：WebdavAccount → webdav_account。
func snake(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				prevLower := s[i-1] >= 'a' && s[i-1] <= 'z'
				prevDigit := s[i-1] >= '0' && s[i-1] <= '9'
				nextLower := i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z'
				if prevLower || prevDigit || nextLower {
					b.WriteByte('_')
				}
			}
			b.WriteByte(c + ('a' - 'A'))
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// pluralize 做最简单的英语复数化，够覆盖本项目表名；不规则表名用 table 选项覆盖。
func pluralize(s string) string {
	switch {
	case s == "":
		return ""
	case strings.HasSuffix(s, "y") && len(s) > 1 && !isVowel(s[len(s)-2]):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(s, "s"), strings.HasSuffix(s, "x"), strings.HasSuffix(s, "z"),
		strings.HasSuffix(s, "ch"), strings.HasSuffix(s, "sh"):
		return s + "es"
	default:
		return s + "s"
	}
}

func isVowel(c byte) bool {
	switch c {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}
