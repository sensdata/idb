package scriptman

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *ScriptMan) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("/actions") // 修改URL路径

	if err != nil {
		LOG.Error("failed to send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		LOG.Error("received error response: %s", resp.Status())
		return nil, fmt.Errorf("received error response: %s", resp.Status())
	}

	return &actionResponse, nil
}

func (s *ScriptMan) checkRepo(hostID uint, repoPath string) error {
	req := model.GitInit{HostID: hostID, RepoPath: repoPath, IsBare: false}
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Git_Init,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("Failed to init repo %s in host %d", repoPath, hostID)
		return fmt.Errorf("failed to init repo")
	}

	return nil
}

func (s *ScriptMan) getScriptList(req model.QueryScript) (*model.PageResult, error) {
	var pageResult = model.PageResult{Total: 0, Items: nil}

	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "global")
	default:
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "local")
	}
	gitQuery := model.GitQuery{
		HostID:       req.HostID,
		RepoPath:     repoPath,
		RelativePath: req.Category,
		Extension:    ".sh",
		Page:         req.Page,
		PageSize:     req.PageSize,
	}

	// 检查repo
	err := s.checkRepo(gitQuery.HostID, gitQuery.RepoPath)
	if err != nil {
		return &pageResult, nil
	}

	// 查询脚本
	data, err := utils.ToJSONString(gitQuery)
	if err != nil {
		return &pageResult, nil
	}

	actionRequest := model.HostAction{
		HostID: gitQuery.HostID,
		Action: model.Action{
			Action: model.Git_File_List,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &pageResult, err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return &pageResult, fmt.Errorf("failed to get script list")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to script list: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *ScriptMan) getScriptDetail(req model.GetScript) (*model.GitFile, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "global")
	default:
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}
	gitGetFile := model.GitGetFile{
		HostID:       req.HostID,
		RepoPath:     repoPath,
		RelativePath: relativePath,
	}

	// 检查repo
	err := s.checkRepo(gitGetFile.HostID, gitGetFile.RepoPath)
	if err != nil {
		return nil, err
	}

	// 获取脚本详情
	data, err := utils.ToJSONString(gitGetFile)
	if err != nil {
		return nil, err
	}

	actionRequest := model.HostAction{
		HostID: gitGetFile.HostID,
		Action: model.Action{
			Action: model.Git_File,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return nil, err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return nil, fmt.Errorf("failed to get script detail")
	}

	var gitFile model.GitFile
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &gitFile)
	if err != nil {
		LOG.Error("Error unmarshaling data to script detail: %v", err)
		return nil, fmt.Errorf("json err: %v", err)
	}

	return &gitFile, nil
}

func (s *ScriptMan) create(req model.CreateScript) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "global")
	default:
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}
	gitCreate := model.GitCreate{
		HostID:       req.HostID,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		Content:      req.Content,
	}

	// 检查repo
	err := s.checkRepo(gitCreate.HostID, gitCreate.RepoPath)
	if err != nil {
		return err
	}

	// 创建
	data, err := utils.ToJSONString(gitCreate)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: gitCreate.HostID,
		Action: model.Action{
			Action: model.Git_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to get create script file")
	}

	return nil
}

func (s *ScriptMan) update(req model.UpdateScript) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "global")
	default:
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}
	gitUpdate := model.GitUpdate{
		HostID:       req.HostID,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		Content:      req.Content,
	}

	// 检查repo
	err := s.checkRepo(gitUpdate.HostID, gitUpdate.RepoPath)
	if err != nil {
		return err
	}

	// 更新
	data, err := utils.ToJSONString(gitUpdate)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: gitUpdate.HostID,
		Action: model.Action{
			Action: model.Git_Update,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to update script file")
	}

	return nil
}

func (s *ScriptMan) delete(req model.DeleteScript) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "global")
	default:
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}
	gitDelete := model.GitDelete{
		HostID:       req.HostID,
		RepoPath:     repoPath,
		RelativePath: relativePath,
	}

	// 检查repo
	err := s.checkRepo(gitDelete.HostID, gitDelete.RepoPath)
	if err != nil {
		return err
	}

	// 删除
	data, err := utils.ToJSONString(gitDelete)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: gitDelete.HostID,
		Action: model.Action{
			Action: model.Git_Delete,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to delete script file")
	}

	return nil
}

func (s *ScriptMan) restore(req model.RestoreScript) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "global")
	default:
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}
	gitRestore := model.GitRestore{
		HostID:       req.HostID,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		CommitHash:   req.CommitHash,
	}

	// 检查repo
	err := s.checkRepo(gitRestore.HostID, gitRestore.RepoPath)
	if err != nil {
		return err
	}

	// 更新
	data, err := utils.ToJSONString(gitRestore)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: gitRestore.HostID,
		Action: model.Action{
			Action: model.Git_Restore,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to restore script file")
	}

	return nil
}

func (s *ScriptMan) getScriptLog(req model.ScriptLog) (*model.PageResult, error) {
	var pageResult = model.PageResult{Total: 0, Items: nil}

	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "global")
	default:
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}
	gitLog := model.GitLog{
		HostID:       req.HostID,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}

	// 检查repo
	err := s.checkRepo(gitLog.HostID, gitLog.RepoPath)
	if err != nil {
		return &pageResult, err
	}

	// 获取脚本详情
	data, err := utils.ToJSONString(gitLog)
	if err != nil {
		return &pageResult, err
	}

	actionRequest := model.HostAction{
		HostID: gitLog.HostID,
		Action: model.Action{
			Action: model.Git_Log,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &pageResult, err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return &pageResult, fmt.Errorf("failed to get script logs")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to script logs: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *ScriptMan) getScriptDiff(req model.ScriptDiff) (string, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "global")
	default:
		repoPath = filepath.Join(s.scriptConfig.Script.DataPath, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}
	gitDiff := model.GitDiff{
		HostID:       req.HostID,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		CommitHash:   req.CommitHash,
	}

	// 检查repo
	err := s.checkRepo(gitDiff.HostID, gitDiff.RepoPath)
	if err != nil {
		return "", err
	}

	// 获取脚本差异
	data, err := utils.ToJSONString(gitDiff)
	if err != nil {
		return "", err
	}

	actionRequest := model.HostAction{
		HostID: gitDiff.HostID,
		Action: model.Action{
			Action: model.Git_Diff,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return "", err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return "", fmt.Errorf("failed to get script diff")
	}

	return actionResponse.Data.Action.Data, nil
}

func (s *ScriptMan) execute(req model.ExecuteScript) (*model.ScriptResult, error) {
	result := model.ScriptResult{
		Start: time.Now(),
		End:   time.Now(),
		Out:   "",
		Err:   "",
	}

	logPath := filepath.Join(s.scriptConfig.Script.LogPath, "script-run.log")

	scriptExec := model.ScriptExec{
		ScriptPath: req.ScriptPath,
		LogPath:    logPath,
	}

	data, err := utils.ToJSONString(scriptExec)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Script_Exec,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to get filetree")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to filetree: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *ScriptMan) getScriptRunLog(hostID uint64) (string, error) {

	logPath := filepath.Join(s.scriptConfig.Script.LogPath, "script-run.log")
	req := model.FileContentReq{
		HostID: uint(hostID),
		Path:   logPath,
	}

	data, err := utils.ToJSONString(req)
	if err != nil {
		return "", err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.File_Content,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return "", err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return "", fmt.Errorf("failed to get file content")
	}

	var fileInfo model.FileInfo
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &fileInfo)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content: %v", err)
		return "", fmt.Errorf("json err: %v", err)
	}

	return fileInfo.Content, nil
}
