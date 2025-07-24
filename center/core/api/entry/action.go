package entry

import (
	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/logstream/pkg/types"
	"github.com/sensdata/idb/core/model"
)

func (b *BaseApi) SendAction(c *gin.Context) {
	var req model.HostAction
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result, err := actionService.SendAction(req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, result)
}

func (b *BaseApi) CreateLogStreamTask(c *gin.Context) {
	var req model.CreateTask
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 生成任务
	metadata := map[string]interface{}{
		"log_path": req.LogPath,
	}
	task, err := global.LogStream.CreateTask(types.TaskTypeRemote, metadata)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), err)
		return
	}
	SuccessWithData(c, model.CreateTaskResponse{
		LogPath: task.LogPath,
	})
}
