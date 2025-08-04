package cmds

import "strings"

func sanitizeFilename(name string) string {
	// 定义非法字符集合
	invalidChars := `/\:*?"<>|`

	// 替换非法字符为下划线
	sanitized := strings.Map(func(r rune) rune {
		if strings.ContainsRune(invalidChars, r) {
			return '_'
		}
		return r
	}, name)

	// 移除首尾空格
	sanitized = strings.TrimSpace(sanitized)

	// 如果名称为空，返回默认值
	if sanitized == "" {
		return "unknown"
	}

	return sanitized
}
