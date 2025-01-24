package terminal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// parseTmuxTime 解析tmux的时间字符串，支持相对时间和绝对时间格式
func parseTmuxTime(timeStr string) (time.Time, error) {
	// 处理空字符串
	if strings.TrimSpace(timeStr) == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	// 尝试解析相对时间格式
	if isRelativeTime(timeStr) {
		return parseRelativeTime(timeStr)
	}

	// 尝试解析绝对时间格式
	return parseAbsoluteTime(timeStr)
}

// isRelativeTime 判断是否为相对时间格式
func isRelativeTime(timeStr string) bool {
	relativePattern := regexp.MustCompile(`(?i)\d+\s+(second|minute|hour|day|week|month|year)s?\s+ago`)
	return relativePattern.MatchString(timeStr)
}

// parseRelativeTime 解析相对时间格式
func parseRelativeTime(timeStr string) (time.Time, error) {
	// 提取数字和时间单位
	parts := strings.Fields(strings.ToLower(timeStr))
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("invalid relative time format: %s", timeStr)
	}

	// 解析数字
	value, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid number in relative time: %s", timeStr)
	}

	// 获取时间单位
	unit := strings.TrimSuffix(parts[1], "s") // 移除可能的复数形式

	// 计算时间
	now := time.Now()
	switch unit {
	case "second":
		return now.Add(time.Duration(-value) * time.Second), nil
	case "minute":
		return now.Add(time.Duration(-value) * time.Minute), nil
	case "hour":
		return now.Add(time.Duration(-value) * time.Hour), nil
	case "day":
		return now.AddDate(0, 0, -value), nil
	case "week":
		return now.AddDate(0, 0, -value*7), nil
	case "month":
		return now.AddDate(0, -value, 0), nil
	case "year":
		return now.AddDate(-value, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unknown time unit: %s", unit)
	}
}

// parseAbsoluteTime 解析绝对时间格式
func parseAbsoluteTime(timeStr string) (time.Time, error) {
	// 尝试多种常见的时间格式
	formats := []string{
		"Mon Jan 2 15:04:05 2006",
		"Mon Jan 2 15:04:05 MST 2006",
		"2006-01-02 15:04:05",
		"Jan 2 15:04:05",
		"2006/01/02 15:04:05",
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("unable to parse absolute time '%s': %v", timeStr, lastErr)
}
