package scriptmanager

import (
	"net/http"

	"github.com/sensdata/idb/core/plugin"
)

type ScriptRouter struct {
}

func NewScriptRouter() *ScriptRouter {
	return &ScriptRouter{}
}

func (h *ScriptRouter) BuildRoutes() []plugin.PluginRoute {
	return []plugin.PluginRoute{
		{Method: http.MethodGet, Path: "/:host", Handler: h.GetScriptList},
	}
}
