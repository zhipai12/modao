package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// TernaryAny 三元运算
func TernaryAny[T any](ok bool, okDo T, noOkDo T) T {
	if ok {
		return okDo
	}

	return noOkDo
}

// UnderScoreToCamel 将下划线分隔的字符串转换为驼峰式字符串
func UnderScoreToCamel(s string) string {
	builder := strings.Builder{}
	nextUpper := true // 控制是否需要转大写

	for _, r := range s {
		if r == '_' {
			nextUpper = true
			continue
		}

		if nextUpper {
			builder.WriteRune(unicode.ToUpper(r))
			nextUpper = false
		} else {
			builder.WriteRune(unicode.ToLower(r)) // 强制非首字母小写（可选）
		}
	}
	return builder.String()
}

// CheckFileIsExist 判断文件是否存在
func CheckFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateFile 创建文件
func CreateFile(filename string) (err error) {
	dir := filepath.Dir(filename)

	// 创建目录（0755权限：用户rwx，组rx，其他rx）
	if err = os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 显式创建文件（0644权限：用户rw，组r，其他r）
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}

	// 确保文件描述符被释放
	if err = file.Close(); err != nil {
		return fmt.Errorf("关闭文件失败: %w", err)
	}

	return nil
}

// FirstToLower 首字母转小写
func FirstToLower(s string) string {
	return strings.ToLower(string(s[0])) + s[1:]
}

// SnakeToPascal 将下划线分隔的字符串转换为大驼峰命名（首字母大写，其余单词首字母大写，删除下划线）
func SnakeToPascal(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	nextUpper := true // 标记下一个字符是否需要转为大写（处理首字母和下划线后的字母）

	for _, ch := range s {
		if ch == '_' {
			nextUpper = true // 遇到下划线，下一个字符需要大写
			continue
		}

		if nextUpper {
			// 当前字符需要大写（首字母或下划线后的首字母）
			result.WriteString(strings.ToUpper(string(ch)))
			nextUpper = false
		} else {
			// 非首字母统一转为小写，保持驼峰风格
			result.WriteString(strings.ToLower(string(ch)))
		}
	}

	return result.String()
}
