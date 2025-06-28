package scriptmanager

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/plugin/manager"
	smpb "github.com/sensdata/idb/center/plugin/scriptmanager/pb"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
)

func init() {
	manager.RegisterFactory("scriptmanager", func(cc *grpc.ClientConn) interface{} {
		return &ScriptManagerWrapper{
			Client: NewGRPCClient(cc),
		}
	})

	api.API.SetUpPluginRouters("plugin/scripts", NewScriptRouter().BuildRoutes())
}

func (h *ScriptRouter) GetScriptList(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	scriptType := c.Query("type")
	if scriptType == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}
	if scriptType != "global" && scriptType != "local" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid type", err)
		return
	}

	category := c.Query("category")
	if category == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid category", err)
		return
	}

	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page", err)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page_size", err)
		return
	}

	req := &smpb.ListScriptsRequest{
		HostId:   uint32(hostID),
		Type:     scriptType,
		Category: category,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	// 获取插件客户端
	plugin, err := manager.PluginMan.GetPlugin("scriptmanager")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "plugin not found"})
		return
	}

	// 类型断言为 gRPC client
	client, ok := plugin.Stub.(ScriptManagerWrapper)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid plugin client"})
		return
	}

	// 调用插件方法
	resp, err := client.ListScripts(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
