package entry

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// @Tags Rsync
// @Summary List rsync task
// @Description List rsync task
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param page query uint true "Page"
// @Param page_size query uint true "Page size"
// @Success 200 {object} model.RsyncListTaskResponse
// @Router /rsync/{host}/task [get]
func (s *BaseApi) RsyncListTask(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RsyncListTaskRequest
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	list, err := rsyncService.ListTask(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, list)
}

// @Tags Rsync
// @Summary Query rsync task
// @Description Query rsync task
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param id query string true "Task ID"
// @Success 200 {object} model.RsyncClientTask
// @Router /rsync/{host}/task/query [get]
func (s *BaseApi) RsyncQueryTask(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RsyncQueryTaskRequest
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	task, err := rsyncService.QueryTask(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, task)
}

// @Tags Rsync
// @Summary Create rsync task
// @Description Create rsync task
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param request body model.RsyncClientCreateTaskRequest true "Request"
// @Success 200 {object} model.RsyncCreateTaskResponse
// @Router /rsync/{host}/task [post]
func (s *BaseApi) RsyncCreateTask(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RsyncClientCreateTaskRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	rsp, err := rsyncService.CreateTask(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, rsp)
}

// @Tags Rsync
// @Summary Delete rsync task
// @Description Delete rsync task
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param id query string true "Task ID"
// @Success 200
// @Router /rsync/{host}/task [delete]
func (s *BaseApi) RsyncDeleteTask(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RsyncDeleteTaskRequest
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	err = rsyncService.DeleteTask(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

// @Tags Rsync
// @Summary Cancel rsync task
// @Description Cancel rsync task
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param request body model.RsyncCancelTaskRequest true "Request"
// @Success 200
// @Router /rsync/{host}/task/cancel [post]
func (s *BaseApi) RsyncCancelTask(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RsyncCancelTaskRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = rsyncService.CancelTask(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

// @Tags Rsync
// @Summary Retry rsync task
// @Description Retry rsync task
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param request body model.RsyncRetryTaskRequest true "Request"
// @Success 200
// @Router /rsync/{host}/task/retry [post]
func (s *BaseApi) RsyncRetryTask(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RsyncRetryTaskRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = rsyncService.RetryTask(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, nil)
}

// @Tags Rsync
// @Summary Test rsync task
// @Description Test rsync task
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param request body model.RsyncTestTaskRequest true "Request"
// @Success 200 {object} model.RsyncTaskLog
// @Router /rsync/{host}/task/test [post]
func (s *BaseApi) RsyncTest(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RsyncTestTaskRequest
	if err := CheckBindAndValidate(&req, c); err != nil {
		return
	}

	log, err := rsyncService.TestTask(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, log)
}

// @Tags Rsync
// @Summary Query rsync task logs
// @Description Query rsync task logs
// @Accept json
// @Produce json
// @Param host path string true "Host"
// @Param id query string true "Task ID"
// @Param page query int true "Page"
// @Param page_size query int true "Page Size"
// @Success 200 {object} model.RsyncTaskLogListResponse
// @Router /rsync/{host}/task/log [get]
func (s *BaseApi) RsyncTaskLogList(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.RsyncTaskLogListRequest
	if err := CheckQueryAndValidate(&req, c); err != nil {
		return
	}

	list, err := rsyncService.TaskLogList(uint(hostID), req)
	if err != nil {
		ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
		return
	}

	SuccessWithData(c, list)
}
