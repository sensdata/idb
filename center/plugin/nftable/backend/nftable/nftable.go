package nftable

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *NFTable) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
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

	LOG.Info("action response: %v", actionResponse)

	return &actionResponse, nil
}

func (s *NFTable) sendCommand(hostId uint, command string) (*model.CommandResult, error) {
	var commandResult model.CommandResult

	commandRequest := model.Command{
		HostID:  hostId,
		Command: command,
	}

	var commandResponse model.CommandResponse

	resp, err := s.restyClient.R().
		SetBody(commandRequest).
		SetResult(&commandResponse).
		Post("/commands")

	if err != nil {
		LOG.Error("failed to send request: %v", err)
		return &commandResult, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		LOG.Error("failed to send request: %v", err)
		return &commandResult, fmt.Errorf("received error response: %s", resp.Status())
	}

	LOG.Info("cmd response: %v", commandResponse)

	commandResult = commandResponse.Data

	return &commandResult, nil
}

func (s *NFTable) checkRepo(hostID uint, repoPath string) error {
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

func (s *NFTable) updateFile(hostID uint64, op model.FileEdit) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Content_Modify,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("failed to edit file")
		return fmt.Errorf("failed to edit file")
	}

	return nil
}

func (s *NFTable) createFile(hostID uint64, op model.FileCreate) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("failed to create file")
		return fmt.Errorf("failed to create file")
	}

	return nil
}

func (s *NFTable) deleteFile(hostID uint64, op model.FileDelete) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.File_Delete,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("failed to delete file")
		return fmt.Errorf("failed to delete file")
	}

	return nil
}

const safeguardRule = "tcp dport { 9918, 9919 } accept"

func (s *NFTable) checkConfContent(content string) (string, error) {
	LOG.Info("check content: %s", content)
	if strings.Contains(content, safeguardRule) {
		LOG.Info("has safegard rule")
		return content, nil
	}

	lines := strings.Split(content, "\n")
	var buffer bytes.Buffer
	for _, line := range lines {
		buffer.WriteString(line + "\n")
		// 在 input 链的定义中添加规则
		if strings.Contains(line, "chain input") {
			buffer.WriteString(fmt.Sprintf("        %s\n", safeguardRule))
		}
	}
	safeContent := buffer.String()
	LOG.Info("safe content: %s", safeContent)
	return safeContent, nil
}

func (s *NFTable) install(hostID uint64) error {
	// 将安装脚本内容保存到一个临时文件
	logPath := filepath.Join("/tmp", "iDB_nftable_install.log")
	scriptPath := fmt.Sprintf("/tmp/iDB_nftable_%s.sh", time.Now().Format("20060102150405"))
	createFile := model.FileCreate{
		Source:  scriptPath,
		Content: string(installShell),
	}
	err := s.createFile(hostID, createFile)
	if err != nil {
		LOG.Error("Failed to create install shell file")
		return err
	}

	// 执行脚本，获得结果
	result := model.ScriptResult{
		Start: time.Now(),
		End:   time.Now(),
		Out:   "",
		Err:   "",
	}

	scriptExec := model.ScriptExec{
		ScriptPath: scriptPath,
		LogPath:    logPath,
	}

	data, err := utils.ToJSONString(scriptExec)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Script_Exec,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		LOG.Error("Failed to run install script")
		return fmt.Errorf("failed to run install script")
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to get filetree")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		LOG.Error("Error unmarshaling data to filetree: %v", err)
		return fmt.Errorf("json err: %v", err)
	}
	LOG.Info("Install result: %v", result)
	if result.Out == "Failed" {
		return fmt.Errorf("Install failed")
	}

	// 使用默认内容覆盖 /etc/nftables.conf内容
	editFile := model.FileEdit{
		Source:  "/etc/nftables.conf",
		Content: string(templateConf),
	}
	err = s.updateFile(hostID, editFile)
	if err != nil {
		LOG.Error("Failed to edit conf")
		return err
	}

	// 启动服务
	return s.switchTo(hostID, model.SwitchOptions{Option: "nftables"})
}

func (s *NFTable) toggle(hostID uint64, req model.ToggleOptions) error {
	switch req.Option {
	case "on":
		command := "systemctl enable nftables"
		_, err := s.sendCommand(uint(hostID), command)
		if err != nil {
			LOG.Error("Failed to enable nftables")
			return err
		}
		command = "systemctl start nftables"
		_, err = s.sendCommand(uint(hostID), command)
		if err != nil {
			LOG.Error("Failed to start nftables")
			return err
		}

	case "off":
		command := "systemctl stop nftables"
		_, err := s.sendCommand(uint(hostID), command)
		if err != nil {
			LOG.Error("Failed to stop nftables")
			return err
		}
		command = "systemctl disable nftables"
		_, err = s.sendCommand(uint(hostID), command)
		if err != nil {
			LOG.Error("Failed to disable nftables")
			return err
		}

	default:
		return fmt.Errorf("invalid option")
	}

	return nil
}

func (s *NFTable) switchTo(hostID uint64, req model.SwitchOptions) error {

	var content string
	switch req.Option {
	case "nftables":
		content = string(switchToNftables)
	case "iptables":
		content = string(switchToIptables)
	default:
		return fmt.Errorf("invalid option")
	}

	// 将切换脚本内容保存到一个临时文件
	logPath := filepath.Join("/tmp", "iDB_nftable_switch.log")
	scriptPath := fmt.Sprintf("/tmp/iDB_nftable_switch_%s.sh", time.Now().Format("20060102150405"))

	createFile := model.FileCreate{
		Source:  scriptPath,
		Content: content,
	}
	err := s.createFile(hostID, createFile)
	if err != nil {
		LOG.Error("Failed to create switch shell file")
		return err
	}

	// 执行脚本，获得结果
	result := model.ScriptResult{
		Start: time.Now(),
		End:   time.Now(),
		Out:   "",
		Err:   "",
	}

	scriptExec := model.ScriptExec{
		ScriptPath: scriptPath,
		LogPath:    logPath,
	}

	data, err := utils.ToJSONString(scriptExec)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Script_Exec,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		LOG.Error("Failed to run switch script")
		return fmt.Errorf("failed to run switch script")
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to run switch script")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		LOG.Error("Error unmarshaling data to filetree: %v", err)
		return fmt.Errorf("json err: %v", err)
	}
	LOG.Info("Switch result: %v", result)
	out := strings.TrimSpace(result.Out)
	if !strings.HasSuffix(out, "Success") {
		return fmt.Errorf("switch failed")
	}

	return nil
}

func (s *NFTable) getConfList(hostID uint64, req model.QueryGitFile) (*model.PageResult, error) {
	var pageResult = model.PageResult{Total: 0, Items: nil}

	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	gitQuery := model.GitQuery{
		HostID:       uint(hostID),
		RepoPath:     repoPath,
		RelativePath: req.Category,
		Extension:    ".nftable;.linked", //筛选.nftable.linked
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
		return &pageResult, fmt.Errorf("failed to get conf list")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf list: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *NFTable) create(hostID uint64, req model.CreateGitFile, extension string) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+extension)
	} else {
		relativePath = req.Name + extension
	}
	gitCreate := model.GitCreate{
		HostID:       uint(hostID),
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
		return fmt.Errorf("failed to get create conf file")
	}

	return nil
}

func (s *NFTable) getContent(hostID uint64, req model.GetGitFileDetail) (*model.GitFile, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".nftable")
	} else {
		relativePath = req.Name + ".nftable"
	}
	gitGetFile := model.GitGetFile{
		HostID:       uint(hostID),
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
		return nil, fmt.Errorf("failed to get conf detail")
	}

	var gitFile model.GitFile
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &gitFile)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf detail: %v", err)
		return nil, fmt.Errorf("json err: %v", err)
	}

	return &gitFile, nil
}

func (s *NFTable) update(hostID uint64, req model.UpdateGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".nftable")
	} else {
		relativePath = req.Name + ".nftable"
	}
	gitUpdate := model.GitUpdate{
		HostID:       uint(hostID),
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
		return fmt.Errorf("failed to update conf file")
	}

	return nil
}

func (s *NFTable) delete(hostID uint64, req model.DeleteGitFile, extension string) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+extension)
	} else {
		relativePath = req.Name + extension
	}
	gitDelete := model.GitDelete{
		HostID:       uint(hostID),
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
		return fmt.Errorf("failed to delete conf file")
	}

	return nil
}

func (s *NFTable) restore(hostID uint64, req model.RestoreGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".nftable")
	} else {
		relativePath = req.Name + ".nftable"
	}
	gitRestore := model.GitRestore{
		HostID:       uint(hostID),
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
		return fmt.Errorf("failed to restore conf file")
	}

	return nil
}

func (s *NFTable) getConfLog(hostID uint64, req model.GitFileLog) (*model.PageResult, error) {
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
		relativePath = filepath.Join(req.Category, req.Name+".nftable")
	} else {
		relativePath = req.Name + ".nftable"
	}
	gitLog := model.GitLog{
		HostID:       uint(hostID),
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
		return &pageResult, fmt.Errorf("failed to get conf logs")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf logs: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *NFTable) getConfDiff(hostID uint64, req model.GitFileDiff) (string, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".nftable")
	} else {
		relativePath = req.Name + ".nftable"
	}
	gitDiff := model.GitDiff{
		HostID:       uint(hostID),
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
		return "", fmt.Errorf("failed to get conf diff")
	}

	return actionResponse.Data.Action.Data, nil
}

func (s *NFTable) confAction(hostID uint64, req model.ServiceAction) error {

	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".nftable")
	} else {
		relativePath = req.Name + ".nftable"
	}
	gitGetFile := model.GitGetFile{
		HostID:       uint(hostID),
		RepoPath:     repoPath,
		RelativePath: relativePath,
	}

	// 检查repo
	err := s.checkRepo(uint(hostID), repoPath)
	if err != nil {
		return err
	}

	// 获取脚本详情
	data, err := utils.ToJSONString(gitGetFile)
	if err != nil {
		return err
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
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to get conf detail")
	}

	var gitFile model.GitFile
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &gitFile)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf detail: %v", err)
		return fmt.Errorf("json err: %v", err)
	}

	switch req.Action {
	case "activate":
		// 应用时，需要检查content
		content, err := s.checkConfContent(gitFile.Content)
		if err != nil {
			return err
		}

		// 覆盖 /etc/nftables.conf内容
		editFile := model.FileEdit{
			Source:  "/etc/nftables.conf",
			Content: content,
		}
		err = s.updateFile(hostID, editFile)
		if err != nil {
			LOG.Error("Failed to edit conf")
			return err
		}

		// 进行测试
		command := "nft -c -f /etc/nftables.conf"
		commandResult, err := s.sendCommand(uint(hostID), command)
		if err != nil {
			LOG.Error("Failed to test conf")
			return err
		}
		LOG.Info("Conf test result: %s", commandResult.Result)
		if commandResult.Result != "" {
			return fmt.Errorf("test failed")
		}

		// 生效规则
		command = "nft -f /etc/nftables.conf"
		commandResult, err = s.sendCommand(uint(hostID), command)
		if err != nil {
			LOG.Error("Failed to enable conf")
			return err
		}
		LOG.Info("Conf enable result: %s", commandResult.Result)
		if commandResult.Result != "" {
			return fmt.Errorf("enable failed")
		}

	case "deactivate":
		// 使用默认内容覆盖 /etc/nftables.conf内容
		templateContent := string(templateConf)

		// 应用时，需要检查content
		content, err := s.checkConfContent(templateContent)
		if err != nil {
			return err
		}

		// 覆盖 /etc/nftables.conf内容
		editFile := model.FileEdit{
			Source:  "/etc/nftables.conf",
			Content: content,
		}
		err = s.updateFile(hostID, editFile)
		if err != nil {
			LOG.Error("Failed to edit conf")
			return err
		}

		// 进行测试
		command := "nft -c -f /etc/nftables.conf"
		commandResult, err := s.sendCommand(uint(hostID), command)
		if err != nil {
			LOG.Error("Failed to test conf")
			return err
		}
		LOG.Info("Conf test result: %s", commandResult.Result)
		if commandResult.Result != "" {
			return fmt.Errorf("test failed")
		}

		// 生效规则
		command = "nft -f /etc/nftables.conf"
		commandResult, err = s.sendCommand(uint(hostID), command)
		if err != nil {
			LOG.Error("Failed to enable conf")
			return err
		}
		LOG.Info("Conf enable result: %s", commandResult.Result)
		if commandResult.Result != "" {
			return fmt.Errorf("enable failed")
		}

	default:
		return errors.New("unsupported action")
	}

	return nil
}
