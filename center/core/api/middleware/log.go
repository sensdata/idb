package middleware

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/global"
)

const (
	// maxLogLength 定义日志内容的最大长度（字符数）
	// 超过此长度的内容会被截断并添加省略标记
	maxLogLength = 2000
)

// logWhitelist 定义需要记录详细日志（Query 和 Body）的路由白名单
// 使用路径前缀匹配，支持完整路径或路径前缀
var logWhitelist = []string{
	"api/v1/actions",
	"api/v1/commands",
	"api/v1/groups",
	"api/v1/home",
	"api/v1/hosts",
	"api/v1/logs",
	"api/v1/mysql",
	"api/v1/pma",
	"api/v1/postgresql",
	"api/v1/public",
	"api/v1/redis",
	"api/v1/rsync",
	"api/v1/scripts",
	"api/v1/settings",
	"api/v1/store",
	"api/v1/terminals",
	"api/v1/transfer",
	"api/v1/users",
}

// sensitivePaths 定义敏感路径，这些路径的 Query 和 Body 不会被记录
// 即使它们在白名单中，也不会记录详细信息
var sensitivePaths = []string{
	"api/v1/auth/sessions",  // 登录接口，包含密码
	"api/v1/users/password", // 更新密码接口，包含密码
}

// RequestLogger 返回一个日志中间件
// 该中间件会记录请求的基本信息，但只有白名单中的路由才会记录 Query 和 Body
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果是 SSE，则绕过日志中间件
		if strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		// 记录请求基本信息（所有请求都记录）
		global.LOG.Info("Request: %s %s", c.Request.Method, path)

		// 检查是否是敏感路径
		isSensitive := isSensitivePath(path)

		// 检查是否在白名单中
		shouldLogDetails := isWhitelisted(path) && !isSensitive

		// 根据请求方法记录详细信息（仅白名单且非敏感路径）
		if shouldLogDetails {
			if c.Request.Method == "GET" {
				query := c.Request.URL.Query()
				if len(query) > 0 {
					queryStr := fmt.Sprintf("%v", query)
					global.LOG.Info("Query: %s", truncateString(queryStr))
				}
			} else if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
				// 读取请求体（用于日志记录）
				var bodyBytes []byte
				if c.Request.Body != nil {
					var readErr error
					bodyBytes, readErr = io.ReadAll(c.Request.Body)
					if readErr != nil {
						// 读取失败时记录错误，但不影响请求处理
						global.LOG.Warn("Failed to read request body for logging: %v", readErr)
						// 如果读取失败，创建一个空的 Body，避免后续处理出错
						c.Request.Body = io.NopCloser(bytes.NewBuffer(nil))
					} else {
						// 重新设置请求体，以便后续处理
						c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					}
				}
				if len(bodyBytes) > 0 {
					// 先转换为字符串，避免二进制数据产生乱码
					bodyStr := string(bodyBytes)
					global.LOG.Info("Body: %s", truncateString(bodyStr))
				}
			}
		} else if isSensitive {
			// 敏感路径只记录基本信息，不记录 Query 和 Body
			global.LOG.Info("Sensitive path, skipping detailed logging")
		}

		// 继续处理请求
		c.Next()

		// 记录响应信息（所有请求都记录）
		global.LOG.Info("Response: %d", c.Writer.Status())
	}
}

// isWhitelisted 检查路径是否在白名单中
// 支持完整路径匹配和路径前缀匹配
func isWhitelisted(path string) bool {
	// 规范化路径：移除前导斜杠（如果存在）
	normalizedPath := strings.TrimPrefix(path, "/")

	for _, whitelistPath := range logWhitelist {
		// 完整路径匹配
		if normalizedPath == whitelistPath {
			return true
		}
		// 路径前缀匹配（支持 api/v1/hosts 匹配 api/v1/hosts/xxx）
		if strings.HasPrefix(normalizedPath, whitelistPath+"/") {
			return true
		}
	}
	return false
}

// isSensitivePath 检查路径是否是敏感路径
// 敏感路径的 Query 和 Body 不会被记录
func isSensitivePath(path string) bool {
	// 规范化路径：移除前导斜杠（如果存在）
	normalizedPath := strings.TrimPrefix(path, "/")

	for _, sensitivePath := range sensitivePaths {
		// 完整路径匹配
		if normalizedPath == sensitivePath {
			return true
		}
		// 路径前缀匹配
		if strings.HasPrefix(normalizedPath, sensitivePath+"/") {
			return true
		}
	}
	return false
}

// truncateString 截断字符串，如果超过最大长度则截断并添加省略标记
// 先转换为字符串后再判断长度，避免二进制数据产生乱码
func truncateString(s string) string {
	// 使用 rune 计数，确保正确处理多字节字符（如中文）
	runes := []rune(s)
	if len(runes) <= maxLogLength {
		return s
	}
	// 截断并添加省略标记
	return string(runes[:maxLogLength]) + fmt.Sprintf("... [truncated, total length: %d]", len(runes))
}
