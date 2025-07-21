package plugin

import (
	"encoding/json"
	"strings"

	"github.com/sensdata/idb/center/global"
)

type PluginLogWriter struct {
}

func (w *PluginLogWriter) Write(p []byte) (n int, err error) {
	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err != nil {
		return 0, err
	}

	msgVal, ok1 := logEntry["msg"]
	levelVal, ok2 := logEntry["level"]
	if !ok1 || !ok2 {
		global.LOG.Warn("plugin log missing fields", "raw", string(p))
		return len(p), nil
	}

	msg, _ := msgVal.(string)
	level := strings.ToLower(levelVal.(string))

	switch level {
	case "trace", "debug":
		global.LOG.Info("plugin", "msg", msg, "data", logEntry)
	case "info":
		global.LOG.Info("plugin", "msg", msg, "data", logEntry)
	case "warn":
		global.LOG.Warn("plugin", "msg", msg, "data", logEntry)
	case "error":
		global.LOG.Error("plugin", "msg", msg, "data", logEntry)
	default:
		global.LOG.Info("plugin", "msg", msg, "data", logEntry)
	}

	return len(p), nil
}
