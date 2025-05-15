package utils

import (
	"fmt"
	"time"
)

func FormatContainerLogTimeFilter(minutes int) string {
	// 将分钟数转换为几个选项：all, 24h, 4h, 1h, 10m
	switch {
	case minutes <= 24*60:
		return "24h"
	case minutes <= 4*60:
		return "4h"
	case minutes <= 60:
		return "1h"
	case minutes <= 10:
		return "10m"
	default:
		return "all"
	}
}

func FormatDuration(sec int64) string {
	// 将秒转换为天、小时、分钟、秒
	duration := time.Duration(sec) * time.Second
	days := duration / (24 * time.Hour)
	hours := (duration % (24 * time.Hour)) / time.Hour
	minutes := (duration % time.Hour) / time.Minute
	seconds := (duration % time.Minute) / time.Second

	// 格式化输出
	return fmt.Sprintf("%d days %d hours %d minutes %d seconds", days, hours, minutes, seconds)
}

func FormatTime(timestamp int64) string {
	// 将 Unix 时间戳转换为时间对象
	t := time.Unix(timestamp, 0)
	// 格式化为 "YYYY-MM-DD HH:MM:SS" 格式
	formattedTime := t.Format("2006-01-02 15:04:05")
	return formattedTime
}

func FormatMemorySize(size uint64) string {
	const (
		_        = iota
		K uint64 = 1 << (10 * iota)
		M
		G
		T
	)

	switch {
	case size >= T:
		return fmt.Sprintf("%.2fT", float64(size)/float64(T))
	case size >= G:
		return fmt.Sprintf("%.2fG", float64(size)/float64(G))
	case size >= M:
		return fmt.Sprintf("%.2fM", float64(size)/float64(M))
	case size >= K:
		return fmt.Sprintf("%.2fK", float64(size)/float64(K))
	default:
		return fmt.Sprintf("%dB", size)
	}
}
