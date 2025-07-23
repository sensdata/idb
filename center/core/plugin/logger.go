package plugin

import (
	"github.com/sensdata/idb/center/global"
)

type PluginLogWriter struct {
}

func (w *PluginLogWriter) Write(p []byte) (n int, err error) {
	global.LOG.Info(string(p))
	return len(p), nil
}
