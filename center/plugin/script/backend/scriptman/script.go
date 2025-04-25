package scriptman

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sensdata/idb/center/core/api/service"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/files"
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
		HostID: uint(hostID),
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

func (s *ScriptMan) handleHostID(reqType string, hostID uint64) (uint, error) {
	var hid = uint(hostID)
	// 如果是global, 操作本机
	if reqType == "global" {
		defaultHost, err := s.hostRepo.Get(s.hostRepo.WithByDefault())
		if err != nil {
			return 0, err
		}
		hid = defaultHost.ID
	}
	return hid, nil
}

func (s *ScriptMan) getCategories(hostID uint64, req model.QueryGitFile) (*model.PageResult, error) {
	var pageResult = model.PageResult{Total: 0, Items: nil}

	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return &pageResult, err
	}

	gitQuery := model.GitQuery{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: req.Category,
		Extension:    "directory",
		Page:         req.Page,
		PageSize:     req.PageSize,
	}

	// 检查repo
	err = s.checkRepo(gitQuery.HostID, gitQuery.RepoPath)
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

func (s *ScriptMan) createCategory(hostID uint64, req model.CreateGitCategory) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}
	gitCreate := model.GitCreate{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: req.Category,
		Dir:          true,
		Content:      "",
	}

	// 检查repo
	err = s.checkRepo(gitCreate.HostID, gitCreate.RepoPath)
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

func (s *ScriptMan) updateCategory(hostID uint64, req model.UpdateGitCategory) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	gitUpdate := model.GitUpdate{
		HostID:          hid,
		RepoPath:        repoPath,
		RelativePath:    req.Category,
		NewRelativePath: req.NewName,
		Dir:             true,
		Content:         "",
	}

	// 检查repo
	err = s.checkRepo(gitUpdate.HostID, gitUpdate.RepoPath)
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

func (s *ScriptMan) deleteCategory(hostID uint64, req model.DeleteGitCategory) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	gitDelete := model.GitDelete{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: req.Category,
		Dir:          true,
	}

	// 检查repo
	err = s.checkRepo(gitDelete.HostID, gitDelete.RepoPath)
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

func (s *ScriptMan) getScriptList(hostID uint64, req model.QueryGitFile) (*model.PageResult, error) {
	var pageResult = model.PageResult{Total: 0, Items: nil}

	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return &pageResult, err
	}

	gitQuery := model.GitQuery{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: req.Category,
		Extension:    ".sh",
		Page:         req.Page,
		PageSize:     req.PageSize,
	}

	// 检查repo
	err = s.checkRepo(gitQuery.HostID, gitQuery.RepoPath)
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

func (s *ScriptMan) getScriptDetail(hostID uint64, req model.GetGitFileDetail) (*model.GitFile, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return nil, err
	}

	gitGetFile := model.GitGetFile{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
	}

	// 检查repo
	err = s.checkRepo(gitGetFile.HostID, gitGetFile.RepoPath)
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

func (s *ScriptMan) create(hostID uint64, req model.CreateGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	gitCreate := model.GitCreate{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		Dir:          false,
		Content:      req.Content,
	}

	// 检查repo
	err = s.checkRepo(gitCreate.HostID, gitCreate.RepoPath)
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

func (s *ScriptMan) update(hostID uint64, req model.UpdateGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}
	var newName string
	if req.NewName != "" {
		newName = req.NewName
	} else {
		newName = req.Name
	}
	var newRelativePath string
	var newCategory string
	if req.NewCategory != "" {
		newCategory = req.NewCategory
	} else {
		newCategory = req.Category
	}
	if newCategory != "" {
		newRelativePath = filepath.Join(newCategory, newName+".sh")
	} else {
		newRelativePath = newName + ".sh"
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	gitUpdate := model.GitUpdate{
		HostID:          hid,
		RepoPath:        repoPath,
		RelativePath:    relativePath,
		NewRelativePath: newRelativePath,
		Dir:             false,
		Content:         req.Content,
	}

	// 检查repo
	err = s.checkRepo(gitUpdate.HostID, gitUpdate.RepoPath)
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

func (s *ScriptMan) delete(hostID uint64, req model.DeleteGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	gitDelete := model.GitDelete{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		Dir:          false,
	}

	// 检查repo
	err = s.checkRepo(gitDelete.HostID, gitDelete.RepoPath)
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

func (s *ScriptMan) restore(hostID uint64, req model.RestoreGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	gitRestore := model.GitRestore{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		CommitHash:   req.CommitHash,
	}

	// 检查repo
	err = s.checkRepo(gitRestore.HostID, gitRestore.RepoPath)
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

func (s *ScriptMan) getScriptLog(hostID uint64, req model.GitFileLog) (*model.PageResult, error) {
	var pageResult = model.PageResult{Total: 0, Items: nil}

	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return &pageResult, err
	}

	gitLog := model.GitLog{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}

	// 检查repo
	err = s.checkRepo(gitLog.HostID, gitLog.RepoPath)
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

func (s *ScriptMan) getScriptDiff(hostID uint64, req model.GitFileDiff) (string, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".sh")
	} else {
		relativePath = req.Name + ".sh"
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return "", err
	}

	gitDiff := model.GitDiff{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		CommitHash:   req.CommitHash,
	}

	// 检查repo
	err = s.checkRepo(gitDiff.HostID, gitDiff.RepoPath)
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

func (s *ScriptMan) syncGlobal(hostID uint) error {
	LOG.Info("Start syncing global scripts for host %d", hostID)

	defaultHost, err := s.hostRepo.Get(s.hostRepo.WithByDefault())
	if err != nil {
		LOG.Error("Failed to get default host: %v", err)
		return err
	}
	if hostID == defaultHost.ID {
		LOG.Error("Attempting to sync global scripts on default host (ID: %d)", hostID)
		return fmt.Errorf("can't sync global script in default host")
	}

	settingService := service.NewISettingsService()
	settingInfo, _ := settingService.Settings()
	scheme := "http"
	if settingInfo.Https == "yes" {
		scheme = "https"
		LOG.Info("Using HTTPS for sync")
	}
	host := global.Host
	if settingInfo.BindDomain != "" && settingInfo.BindDomain != host {
		host = settingInfo.BindDomain
		LOG.Info("Using custom domain: %s", host)
	}
	remoteUrl := fmt.Sprintf("%s://%s:%d/api/v1/git/scripts/global", scheme, host, settingInfo.BindPort)
	repoPath := filepath.Join(s.pluginConf.Items.WorkDir, "global")

	LOG.Info("Syncing from %s to %s", remoteUrl, repoPath)

	gitSync := model.GitSync{
		HostID:    hostID,
		RepoPath:  repoPath,
		RemoteUrl: remoteUrl,
	}

	data, err := utils.ToJSONString(gitSync)
	if err != nil {
		LOG.Error("Failed to marshal git sync data: %v", err)
		return err
	}

	LOG.Info("Sending sync request to agent")
	actionRequest := model.HostAction{
		HostID: gitSync.HostID,
		Action: model.Action{
			Action: model.Git_Sync,
			Data:   data,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		LOG.Error("Failed to send sync action: %v", err)
		return err
	}
	if !actionResponse.Data.Action.Result {
		LOG.Error("Sync action failed on agent")
		return fmt.Errorf("failed to sync global script")
	}

	LOG.Info("Successfully synced global scripts for host %d", hostID)
	return nil
}

func (s *ScriptMan) execute(hostID uint, req model.ExecuteScript) (*model.ScriptResult, error) {
	result := model.ScriptResult{
		LogPath: "",
		Start:   time.Now(),
		End:     time.Now(),
		Out:     "",
		Err:     "",
	}

	scriptName := filepath.Base(req.ScriptPath)
	scriptName = strings.TrimSuffix(scriptName, filepath.Ext(scriptName))
	// 根据脚本路径创建同名日志目录
	timestamp := time.Now().Local().Format("20060102_150405")
	scriptDir := filepath.Dir(req.ScriptPath)
	logPath := filepath.Join(scriptDir, scriptName, timestamp+".log")

	scriptExec := model.ScriptExec{
		ScriptPath: req.ScriptPath,
		LogPath:    logPath,
	}

	data, err := utils.ToJSONString(scriptExec)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
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
		global.LOG.Error("Error unmarshaling data to script result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *ScriptMan) getScriptRunLogs(hostID uint, scriptPath string, page int, pageSize int) (*model.PageResult, error) {
	var result = model.PageResult{Total: 0, Items: nil}

	// 枚举所有运行日志
	scriptName := filepath.Base(scriptPath)
	scriptName = strings.TrimSuffix(scriptName, filepath.Ext(scriptName))
	// 根据脚本路径创建同名日志目录
	scriptDir := filepath.Dir(scriptPath)
	logDir := filepath.Join(scriptDir, scriptName)

	req := model.FileOption{
		FileOption: files.FileOption{
			Path:       logDir,
			Search:     "",
			Expand:     true,
			ShowHidden: false,
			Dir:        false,
			Page:       page,
			PageSize:   pageSize,
		},
	}

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.File_Search,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to search file")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file list: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *ScriptMan) getScriptRunLogsDetail(hostID uint, logPath string) (*model.FileInfo, error) {
	var info model.FileInfo
	// 读取文件内容
	req := model.FileContentReq{
		Path:   logPath,
		Expand: true,
	}
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &info, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Content,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &info, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &info, fmt.Errorf("failed to get file content")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &info)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content: %v", err)
		return &info, fmt.Errorf("json err: %v", err)
	}

	return &info, nil
}
