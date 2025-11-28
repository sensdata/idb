package entry

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/plugin"
	"github.com/sensdata/idb/center/core/plugin/shared"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

func getRsync() (shared.Rsync, error) {
	// 获取插件客户端
	plugin, err := plugin.PLUGINSERVER.GetPlugin("idb-rsync")
	if err != nil {
		return nil, err
	}
	// 类型断言为 gRPC client
	client, ok := plugin.Stub.(shared.Rsync)
	if !ok {
		return nil, errors.New("invalid plugin client")
	}
	return client, nil
}

// @Tags Transfer
// @Summary List transfer task
// @Description List transfer task
// @Accept json
// @Produce json
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.RsyncListTaskResponse
// @Router /transfer/task [get]
func (a *BaseApi) TransferListTask(c *gin.Context) {
	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page", err)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid page_size", err)
		return
	}

	client, err := getRsync()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	resp, err := client.ListTask(&model.RsyncListTaskRequest{
		Page:     int(page),
		PageSize: int(pageSize),
	})
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, resp)
}

// @Tags Transfer
// @Summary Query transfer task
// @Description Query transfer task
// @Accept json
// @Produce json
// @Param id query string true "Task ID"
// @Success 200 {object} model.RsyncTaskInfo
// @Router /transfer/task/query [get]
func (a *BaseApi) TransferQueryTask(c *gin.Context) {
	taskID := c.Query("id")
	if taskID == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid id", nil)
		return
	}

	client, err := getRsync()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	resp, err := client.QueryTask(&model.RsyncQueryTaskRequest{
		ID: taskID,
	})
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, resp)
}

// @Tags Transfer
// @Summary Create transfer task
// @Description Create transfer task
// @Accept json
// @Produce json
// @Param task body model.RsyncCreateTask true "Task"
// @Success 200 {object} model.RsyncCreateTaskResponse
// @Router /transfer/task [post]
func (a *BaseApi) TransferCreateTask(c *gin.Context) {
	var req model.RsyncCreateTask
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// 查找host
	hostRepo := repo.NewHostRepo()
	srcHost, err := hostRepo.Get(hostRepo.WithByID(uint(req.SrcHostId)))
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
	}
	dstHost, err := hostRepo.Get(hostRepo.WithByID(uint(req.DstHostId)))
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
	}

	client, err := getRsync()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	resp, err := client.CreateTask(&model.RsyncCreateTaskRequest{
		Mode: req.Mode,
		Src:  req.Src,
		Dst:  req.Dst,
		SrcHost: model.RsyncHost{
			ID:       uint64(srcHost.ID),
			Host:     srcHost.Addr,
			Port:     srcHost.Port,
			User:     srcHost.User,
			AuthMode: srcHost.AuthMode,
			Password: srcHost.Password,
			KeyPath:  srcHost.PrivateKey,
		},
		DstHost: model.RsyncHost{
			ID:       uint64(dstHost.ID),
			Host:     dstHost.Addr,
			Port:     dstHost.Port,
			User:     dstHost.User,
			AuthMode: dstHost.AuthMode,
			Password: dstHost.Password,
			KeyPath:  dstHost.PrivateKey,
		},
	})
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, resp)
}

// @Tags Transfer
// @Summary Delete transfer task
// @Description Delete transfer task
// @Accept json
// @Produce json
// @Param id query string true "Task ID"
// @Success 200
// @Router /transfer/task [delete]
func (a *BaseApi) TransferDeleteTask(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid id", nil)
		return
	}

	client, err := getRsync()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	if err := client.DeleteTask(&model.RsyncDeleteTaskRequest{
		ID: id,
	}); err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Transfer
// @Summary Cancel transfer task
// @Description Cancel transfer task
// @Accept json
// @Produce json
// @Param request body model.RsyncCancelTaskRequest true "Request"
// @Success 200
// @Router /transfer/task/cancel [post]
func (a *BaseApi) TransferCancelTask(c *gin.Context) {
	var req model.RsyncCancelTaskRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	client, err := getRsync()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	if err := client.CancelTask(&model.RsyncCancelTaskRequest{
		ID: req.ID,
	}); err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}

// @Tags Transfer
// @Summary Retry transfer task
// @Description Retry transfer task
// @Accept json
// @Produce json
// @Param request body model.RsyncRetryTaskRequest true "Request"
// @Success 200
// @Router /transfer/task/retry [post]
func (a *BaseApi) TransferRetryTask(c *gin.Context) {
	var req model.RsyncRetryTaskRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	client, err := getRsync()
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	if err := client.RetryTask(&model.RsyncRetryTaskRequest{
		ID: req.ID,
	}); err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}
	SuccessWithData(c, nil)
}
