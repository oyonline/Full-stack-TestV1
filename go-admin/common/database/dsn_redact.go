package database

import "strings"

// RedactDataSourceName 返回用于日志输出的 DSN：隐去 user:password 中的密码段。
// 典型 MySQL DSN：user:secret@tcp(127.0.0.1:3306)/db?params
// 未知格式时原样返回（避免误删连接信息）。
func RedactDataSourceName(source string) string {
	for _, marker := range []string{"@tcp(", "@unix("} {
		idx := strings.Index(source, marker)
		if idx <= 0 {
			continue
		}
		head := source[:idx]
		tail := source[idx:]
		colon := strings.Index(head, ":")
		if colon < 0 {
			return source
		}
		return head[:colon+1] + "***" + tail
	}
	return source
}
