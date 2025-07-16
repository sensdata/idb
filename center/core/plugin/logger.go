package plugin

import (
	"io"
	"strings"

	"github.com/sensdata/idb/center/global"
)

type pluginLogWriter struct {
	name  string // 插件名
	level string // "stdout" or "stderr"
}

func pluginLoggerWriter(name, level string) io.Writer {
	return &pluginLogWriter{name: name, level: level}
}

func (w *pluginLogWriter) Write(p []byte) (n int, err error) {
	line := strings.TrimSuffix(string(p), "\n")

	switch w.level {
	case "stdout":
		global.LOG.Info("[plugin:%s stdout] %s", w.name, line)
	case "stderr":
		global.LOG.Error("[plugin:%s stderr] %s", w.name, line)
	default:
		global.LOG.Warn("[plugin:%s ???] %s", w.name, line)
	}

	return len(p), nil
}
