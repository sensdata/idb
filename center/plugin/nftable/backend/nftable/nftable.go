package nftable

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sensdata/idb/center/core/api/service"
	"github.com/sensdata/idb/center/global"
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

func (s *NFTable) fileExist(hostID uint, path string) (bool, error) {
	command := fmt.Sprintf("test -f %s && echo true || echo false", path)
	commandResult, err := s.sendCommand(hostID, command)
	if err != nil {
		return false, err
	}
	trimmed := strings.TrimSpace(commandResult.Result)
	LOG.Info("file exist: %s, raw result: %q, trimmed: %q", path, commandResult.Result, trimmed)
	return trimmed == "true", nil
}

func (s *NFTable) fileContent(hostID uint, path string) (string, error) {
	var content string
	req := model.FileContentReq{
		Path:   path,
		Expand: true,
	}
	data, err := utils.ToJSONString(req)
	if err != nil {
		return content, err
	}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.File_Content,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return content, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to get content of file %s", path)
		return content, fmt.Errorf("failed to get file content")
	}

	var info model.FileInfo
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &info)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content: %v", err)
		return content, fmt.Errorf("json err: %v", err)
	}

	return info.Content, nil
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
		return errors.New("failed to init repo")
	}

	// 判断是否存在 /local/default/default.nftable
	defaultConfPath := filepath.Join(repoPath, "default/default.nftable")
	exist, err := s.fileExist(hostID, defaultConfPath)
	if err != nil {
		LOG.Error("Failed to check %s", defaultConfPath)
		return errors.New("failed to check default conf")
	}
	// 不存在则初始化
	if !exist {
		LOG.Info("default.nftable not exists in host %d", hostID)
		var content string
		// 获取 /etc/nftables.conf 内容
		var sysExist bool
		detail, err := s.fileContent(hostID, "/etc/nftables.conf")
		if err != nil {
			LOG.Error("Failed to get /etc/nftables.conf")
			// 获取失败，以模板内容初始化
			content = string(templateConf)
			sysExist = false
		} else {
			// 获取成功，以/etc/nftables.conf的内容初始化
			content = detail
			sysExist = true
		}
		// 检查content 是否包含默认规则
		content, err = s.checkConfContent(content)
		if err != nil {
			LOG.Error("Failed to check conf content: %v", err)
			return err
		}

		gitCreate := model.GitCreate{
			HostID:       hostID,
			RepoPath:     repoPath,
			RelativePath: "default/default.nftable",
			Content:      content,
		}
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
			LOG.Error("failed to create default.nftable")
			return errors.New("failed to create default conf")
		}

		// /etc/nftables.conf不存在，创建它
		if !sysExist {
			createFile := model.FileCreate{
				Source:  "/etc/nftables.conf",
				Content: gitCreate.Content,
			}
			err := s.createFile(uint64(hostID), createFile)
			if err != nil {
				LOG.Error("Failed to create /etc/nftables.conf: %v", err)
				return errors.New("failed to create /etc/nftables.conf")
			}
		}
	}

	return nil
}

func (s *NFTable) handleHostID(reqType string, hostID uint64) (uint, error) {
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

func (s *NFTable) needSync(reqType string, hostID uint64) (bool, error) {
	if reqType == "global" {
		defaultHost, err := s.hostRepo.Get(s.hostRepo.WithByDefault())
		if err != nil {
			return false, err
		}
		if hostID != uint64(defaultHost.ID) {
			return true, nil
		}
	}
	return false, nil
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
		return errors.New("failed to edit file")
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
		return errors.New("failed to create file")
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
		return errors.New("failed to delete file")
	}

	return nil
}

const safeguardRule = "tcp dport { 22, 9918, 9919 } accept"

// 统一转换为 inet idb-filter 表（简化版本）
func (s *NFTable) convertToInetIdbFilter(content string) (string, error) {
	lines := strings.Split(content, "\n")
	var output []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 匹配任何包含 "table" 和 "filter" 的行，替换为标准表名
		if strings.HasPrefix(trimmed, "table ") && strings.Contains(trimmed, "filter") {
			// 简单替换为标准表定义
			output = append(output, "table inet idb-filter {")
			continue
		}

		output = append(output, line)
	}

	return strings.Join(output, "\n"), nil
}

// 确保包含必要的默认规则
func (s *NFTable) ensureDefaultRules(content string) string {
	lines := strings.Split(content, "\n")
	var output []string
	var chainContent []string
	insideChain := false
	indent := "    " // 默认缩进

	defaultRules := map[string]bool{
		"iif lo accept":                       true,
		"iif \"lo\" accept":                   true,
		"iifname \"lo\" accept":               true,
		"iif docker0 accept":                  true,
		"iif \"docker0\" accept":              true,
		"iifname \"docker0\" accept":          true,
		"iif docker_gwbridge accept":          true,
		"iif \"docker_gwbridge\" accept":      true,
		"iifname \"docker_gwbridge\" accept":  true,
		"iifname \"br-+\" accept":             true,
		"iifname \"veth+\" accept":            true,
		"ct state established,related accept": true,
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		currentIndent := getIndent(line)

		// 检测input链开始
		if strings.HasPrefix(trimmed, "chain input") {
			insideChain = true
			output = append(output, line)
			continue
		}

		// 在input链内部
		if insideChain {
			// 检查是否是链结束
			if trimmed == "}" {
				// 固定在链内容最前面添加默认规则
				defaultOrder := []string{
					"iif lo accept",
					"iif docker0 accept",
					"iif docker_gwbridge accept",
					"iifname \"br-+\" accept",
					"iifname \"veth+\" accept",
					"ct state established,related accept",
				}
				for _, r := range defaultOrder {
					output = append(output, indent+r)
				}

				// 添加缓存的链内容
				output = append(output, chainContent...)
				// 添加链结束标记
				output = append(output, line)
				insideChain = false
				continue
			}

			// 跳过已存在的默认规则，避免重复
			if defaultRules[trimmed] {
				continue
			}

			// 缓存其他规则内容
			chainContent = append(chainContent, line)
			// 保存缩进格式
			if currentIndent != "" && indent == "    " {
				indent = currentIndent
			}
		} else {
			// 不在链内，直接添加
			output = append(output, line)
		}
	}

	return strings.Join(output, "\n")
}

func (s *NFTable) checkConfContent(content string) (string, error) {
	LOG.Info("check content: %s", content)
	safeContent := content
	var err error

	// Step 1: 统一转换为 inet idb-filter 表
	safeContent, err = s.convertToInetIdbFilter(safeContent)
	if err != nil {
		LOG.Error("failed to convert to inet idb-filter: %v", err)
		return "", err
	}

	// Step 2: 确保包含必要的默认规则
	safeContent = s.ensureDefaultRules(safeContent)

	// Step 3: 更新安全端口规则
	safePorts := []int{22, 9918, 9919}
	for _, port := range safePorts {
		safeContent, err = updatePortRuleInConfContent(
			safeContent,
			model.PortRule{
				Protocol:  "",
				PortStart: port,
				PortEnd:   port,
				Rules: []model.RuleItem{
					{
						Type:   "default",
						Action: "accept",
					},
				},
			},
		)
		if err != nil {
			LOG.Error("failed to check content for port %d", port)
			return "", err
		}
	}

	// Step 4: 清理 flush ruleset
	safeContent = removeFlushRuleset(safeContent)
	return safeContent, nil
}

// 清理掉规则中的 flush ruleset 行
func removeFlushRuleset(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 跳过 flush ruleset 行
		if strings.HasPrefix(trimmed, "flush ruleset") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

func (s *NFTable) status(hostID uint64) (*model.NftablesStatus, error) {
	var result model.NftablesStatus

	// 命令行查询nftables是否安装，并返回 installed/not installed
	command := "command -v nft >/dev/null 2>&1 && echo \"installed\" || echo \"not installed\""
	commandResult, err := s.sendCommand(uint(hostID), command)
	if err != nil {
		LOG.Error("Failed to check install status")
		return &result, errors.New("failed to check install status")
	}
	LOG.Info("Install status result: %s", commandResult.Result)
	result.Status = strings.TrimSpace(commandResult.Result)

	// 脚本检测防火墙
	logPath := filepath.Join("/tmp", "iDB_nftable_detect.log")
	scriptPath := fmt.Sprintf("/tmp/iDB_nftable_detect_%d.sh", time.Now().Unix())
	createFile := model.FileCreate{
		Source:  scriptPath,
		Content: string(detectFirewall),
	}
	err = s.createFile(hostID, createFile)
	if err != nil {
		LOG.Error("Failed to create detect shell file")
		return &result, err
	}
	// 执行脚本，获得结果
	scriptResult := model.ScriptResult{
		Start: time.Now(),
		End:   time.Now(),
		Out:   "",
		Err:   "",
	}

	scriptExec := model.ScriptExec{
		ScriptPath: scriptPath,
		LogPath:    logPath,
		Remove:     true,
	}

	data, err := utils.ToJSONString(scriptExec)
	if err != nil {
		return &result, err
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
		LOG.Error("Failed to run detect script")
		return &result, errors.New("failed to run detect script")
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return &result, errors.New("failed to run detect script")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &scriptResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to filetree: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}
	LOG.Info("Detect result: %v", scriptResult)
	result.Active = strings.TrimSpace(scriptResult.Out)

	return &result, nil
}

func (s *NFTable) install(hostID uint64) error {
	// 将安装脚本内容保存到一个临时文件
	logPath := filepath.Join("/tmp", "iDB_nftable_install.log")
	scriptPath := fmt.Sprintf("/tmp/iDB_nftable_%d.sh", time.Now().Unix())
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
		Remove:     true,
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
		return errors.New("failed to run install script")
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return errors.New("failed to get filetree")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		LOG.Error("Error unmarshaling data to filetree: %v", err)
		return fmt.Errorf("json err: %v", err)
	}
	LOG.Info("Install result: %v", result)
	if result.Out == "Failed" {
		return errors.New("Install failed")
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
		return errors.New("invalid option")
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
		return errors.New("invalid option")
	}

	// 将切换脚本内容保存到一个临时文件
	logPath := filepath.Join("/tmp", "iDB_nftable_switch.log")
	scriptPath := fmt.Sprintf("/tmp/iDB_nftable_switch_%d.sh", time.Now().Unix())

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
		Remove:     true,
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
		return errors.New("failed to run switch script")
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return errors.New("failed to run switch script")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		LOG.Error("Error unmarshaling data to filetree: %v", err)
		return fmt.Errorf("json err: %v", err)
	}
	LOG.Info("Switch result: %v", result)
	out := strings.TrimSpace(result.Out)
	if !strings.HasSuffix(out, "Success") {
		return errors.New("switch failed")
	}

	// 切换到nftables后，刷一次配置
	if req.Option == "nftables" {
		raw, err := s.getConfRaw(uint(hostID))
		if err != nil {
			LOG.Error("Failed to get conf")
			return err
		}
		err = s.setConfRaw(uint(hostID), model.ConfRaw{Content: raw.Content})
		if err != nil {
			LOG.Error("Failed to set conf")
			return err
		}
	}

	return nil
}

func (s *NFTable) getCategories(hostID uint64, req model.QueryGitFile) (*model.PageResult, error) {
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
		return &pageResult, errors.New("failed to get conf list")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf list: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *NFTable) createCategory(hostID uint64, req model.CreateGitCategory) error {
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
		return errors.New("failed to get create conf file")
	}

	return nil
}

func (s *NFTable) updateCategory(hostID uint64, req model.UpdateGitCategory) error {
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
		return errors.New("failed to update conf file")
	}

	return nil
}

func (s *NFTable) deleteCategory(hostID uint64, req model.DeleteGitCategory) error {
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
		return errors.New("failed to delete conf file")
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

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return &pageResult, err
	}

	gitQuery := model.GitQuery{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: req.Category,
		Extension:    ".nftable", //筛选.nftable
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
		return &pageResult, errors.New("failed to get conf list")
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

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	gitCreate := model.GitCreate{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
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
		return errors.New("failed to create conf file")
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
		return nil, errors.New("failed to get conf detail")
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
		newRelativePath = filepath.Join(newCategory, newName+".nftable")
	} else {
		newRelativePath = newName + ".nftable"
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
		return errors.New("failed to update conf file")
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

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	gitDelete := model.GitDelete{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
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
		return errors.New("failed to delete conf file")
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
		return errors.New("failed to restore conf file")
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
		return &pageResult, errors.New("failed to get conf logs")
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
		return "", errors.New("failed to get conf diff")
	}

	return actionResponse.Data.Action.Data, nil
}

func (s *NFTable) syncGlobal(hostID uint) error {
	LOG.Info("Start syncing global nftabls for host %d", hostID)

	defaultHost, err := s.hostRepo.Get(s.hostRepo.WithByDefault())
	if err != nil {
		LOG.Error("Failed to get default host: %v", err)
		return err
	}
	if hostID == defaultHost.ID {
		LOG.Error("Attempting to sync global nftabls on default host (ID: %d)", hostID)
		return errors.New("can't sync global nftabls in default host")
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
	remoteUrl := fmt.Sprintf("%s://%s:%d/api/v1/git/nftabls/global", scheme, host, settingInfo.BindPort)
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
		return errors.New("failed to sync global nftabls")
	}

	LOG.Info("Successfully synced global nftabls for host %d", hostID)
	return nil
}

func (s *NFTable) confActivate(hostID uint64, req model.ServiceActivate) error {

	// 先看是否需要同步
	needSync, err := s.needSync(req.Type, hostID)
	if err != nil {
		LOG.Error("Failed to check if sync is needed: %v", err)
		return err
	}
	// 执行同步
	if needSync {
		if err := s.syncGlobal(uint(hostID)); err != nil {
			LOG.Error("Failed to sync global services: %v", err)
			return err
		}
	}

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
	// 检查repo
	err = s.checkRepo(uint(hostID), repoPath)
	if err != nil {
		return err
	}

	// 获取脚本详情
	gitGetFile := model.GitGetFile{
		HostID:       uint(hostID),
		RepoPath:     repoPath,
		RelativePath: relativePath,
	}
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
		return errors.New("failed to get conf detail")
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
		err = s.updateFile(uint64(hostID), editFile)
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
			return errors.New("test failed")
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
			return errors.New("enable failed")
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
		err = s.updateFile(uint64(hostID), editFile)
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
			return errors.New("test failed")
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
			return errors.New("enable failed")
		}

	default:
		return errors.New("unsupported action")
	}

	return nil
}

func (s *NFTable) getProcessStatus(hostID uint) (*model.PageResult, error) {
	var result model.PageResult

	// 获取监听端口信息
	command := "ss -ltnpH"
	commandResult, err := s.sendCommand(hostID, command)
	if err != nil {
		LOG.Error("Failed to get port listening info")
		return &result, errors.New("get port listening info failed")
	}
	portInfos := strings.TrimSpace(commandResult.Result)
	LOG.Info("Port listening info: %s", portInfos)
	if portInfos == "" {
		LOG.Error("No port listening info")
		return &result, errors.New("no port listening info")
	}

	// 获取本机 IP 地址
	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.SysInfo_Network,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		LOG.Error("Failed to send SysInfo_Network action: %v", err)
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to get network info")
		return &result, fmt.Errorf("failed to get network info")
	}

	var network model.NetworkInfo
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &network)
	if err != nil {
		LOG.Error("Error unmarshaling data to network info: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	var ips []string
	for _, ni := range network.Networks {
		for _, addr := range ni.Address {
			ips = append(ips, addr.Ip)
		}
	}
	// command = "ip -o addr show | awk '{print $4}' | cut -d/ -f1"
	// commandResult, err = s.sendCommand(hostID, command)
	// if err != nil {
	// 	LOG.Error("Failed to get IP info")
	// 	return &result, errors.New("get ip info failed")
	// }
	// ipInfos := strings.TrimSpace(commandResult.Result)
	// LOG.Info("IP info: %s", ipInfos)
	// if ipInfos == "" {
	// 	LOG.Error("No IP info")
	// 	return &result, errors.New("no ip info")
	// }

	// var ips []string
	// for _, ip := range strings.Split(ipInfos, "\n") {
	// 	ip = strings.TrimSpace(ip)
	// 	if ip != "" {
	// 		ips = append(ips, ip)
	// 	}
	// }

	// 解析 nft ruleset
	command = "nft list ruleset"
	commandResult, err = s.sendCommand(hostID, command)
	if err != nil {
		return &result, errors.New("get ruleset failed")
	}
	rulesetText := strings.TrimSpace(commandResult.Result)
	if rulesetText == "" {
		return &result, errors.New("empty ruleset")
	}

	statusList := []model.ProcessStatus{}
	scanner := bufio.NewScanner(bytes.NewReader([]byte(portInfos)))
	re := regexp.MustCompile(`users:\(\("([^"]+)",pid=(\d+),.*?\)\)`)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		addrPort := fields[3]
		ip, portStr := parseAddrPort(addrPort)
		if ip == "" || portStr == "" {
			continue
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		procMatch := re.FindStringSubmatch(line)
		if len(procMatch) != 3 {
			continue
		}
		processName := procMatch[1]
		pid, _ := strconv.Atoi(procMatch[2])

		var addresses []string
		if ip == "*" || ip == "0.0.0.0" || ip == "::" || ip == "" {
			addresses = ips
		} else {
			addresses = []string{ip}
		}

		accesses := []model.PortAccessStatus{}
		for _, addr := range addresses {
			verdict := matchPortVerdictFromText(rulesetText, port)
			accesses = append(accesses, model.PortAccessStatus{
				Address: addr,
				Status:  verdictToStatus(verdict),
			})
		}

		statusList = append(statusList, model.ProcessStatus{
			Process: processName,
			Pid:     pid,
			Port:    port,
			Access:  refineAccessStatus(accesses),
		})
	}

	result.Total = int64(len(statusList))
	result.Items = statusList
	return &result, nil
}

// matchPortVerdictFromText 根据 nft ruleset 文本，按顺序解析所有表和input链的规则
func matchPortVerdictFromText(ruleset string, port int) string {
	lines := strings.Split(ruleset, "\n")
	portStr := strconv.Itoa(port)

	insideChain := false
	defaultPolicy := "accept"

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "chain input") {
			insideChain = true
			continue
		}

		if insideChain {
			if line == "}" {
				insideChain = false
				continue
			}

			if strings.Contains(line, "hook input") {
				if strings.Contains(line, "policy") {
					policyParts := strings.Split(line, "policy")
					if len(policyParts) > 1 {
						policy := strings.TrimSpace(strings.TrimSuffix(policyParts[1], ";"))
						defaultPolicy = strings.ToLower(policy)
					}
				} else {
					defaultPolicy = "accept"
				}
			}

			if strings.Contains(line, "dport") {
				re := regexp.MustCompile(`dport\s+(\{[^}]+\}|[0-9]+)`)
				matches := re.FindStringSubmatch(line)
				if len(matches) < 2 {
					continue
				}
				portPart := matches[1]

				var ports []string
				if strings.HasPrefix(portPart, "{") {
					portList := strings.Trim(portPart, "{} ")
					for _, p := range strings.Split(portList, ",") {
						ports = append(ports, strings.TrimSpace(p))
					}
				} else {
					ports = []string{portPart}
				}

				for _, p := range ports {
					if p == portStr {
						if strings.HasSuffix(line, " accept") {
							return "accept"
						} else if strings.HasSuffix(line, " drop") || strings.HasSuffix(line, " reject") {
							return "reject"
						}
					}
				}
			}
		}
	}

	return defaultPolicy
}

func verdictToStatus(verdict string) string {
	switch verdict {
	case "accept":
		return "accepted"
	case "drop", "reject":
		return "rejected"
	default:
		return "unknown"
	}
}

func refineAccessStatus(accesses []model.PortAccessStatus) []model.PortAccessStatus {
	allAccepted := true
	localOnly := true
	someRejected := false

	for _, a := range accesses {
		if a.Status != "accepted" {
			allAccepted = false
		}
		if a.Status == "rejected" {
			someRejected = true
		}
		if !strings.HasPrefix(a.Address, "127.") && a.Address != "::1" {
			localOnly = false
		}
	}

	switch {
	case allAccepted && localOnly:
		for i := range accesses {
			accesses[i].Status = "local-only"
		}
	case allAccepted:
		for i := range accesses {
			accesses[i].Status = "fully-accepted"
		}
	case someRejected:
		for i := range accesses {
			if accesses[i].Status == "accepted" {
				accesses[i].Status = "restricted"
			}
		}
	}
	return accesses
}

// parseAddrPort 解析 IP:PORT 字符串，兼容 IPv6、带中括号、%interface 等情况
func parseAddrPort(addrPort string) (string, string) {
	// IPv6 带中括号：[::1]%lo:80 或 [::1]:80
	if strings.HasPrefix(addrPort, "[") {
		endIdx := strings.Index(addrPort, "]")
		if endIdx == -1 {
			return "", ""
		}
		ip := addrPort[1:endIdx]
		// 去除 % 接口信息
		if percent := strings.Index(ip, "%"); percent != -1 {
			ip = ip[:percent]
		}
		rest := addrPort[endIdx+1:]
		if strings.HasPrefix(rest, ":") {
			return ip, rest[1:]
		}
		return "", ""
	}

	// IPv4 或不带中括号的 IPv6
	parts := strings.Split(addrPort, ":")
	if len(parts) < 2 {
		return "", ""
	}
	port := parts[len(parts)-1]
	ip := strings.Join(parts[:len(parts)-1], ":")
	if percent := strings.Index(ip, "%"); percent != -1 {
		ip = ip[:percent]
	}
	return ip, port
}

func (s *NFTable) getBaseRules(hostID uint) (*model.BaseRules, error) {
	var result model.BaseRules

	// 获取 /etc/nftables.conf 内容
	detail, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return &result, err
	}
	scanner := bufio.NewScanner(strings.NewReader(detail))
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	inputPolicy := parseInputPolicy(lines)
	result.InputPolicy = inputPolicy
	return &result, nil
}

func parseInputPolicy(lines []string) string {
	insideInput := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "chain input") {
			insideInput = true
			continue
		}

		if insideInput {
			if trimmed == "}" {
				insideInput = false
				continue
			}

			if strings.Contains(trimmed, "hook input") {
				if strings.Contains(trimmed, "policy") {
					policyParts := strings.Split(trimmed, "policy")
					if len(policyParts) > 1 {
						policy := strings.TrimSpace(strings.TrimSuffix(policyParts[1], ";"))
						return strings.ToLower(policy)
					}
				} else {
					return "accept"
				}
			}
		}
	}

	return ""
}

func (s *NFTable) setBaseRules(hostID uint, req model.BaseRules) error {
	// 获取 /etc/nftables.conf 内容
	confContent, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return fmt.Errorf("failed to get conf detail %v", err)
	}

	// 设置 input policy
	newConfContent, err := setInputPolicy(confContent, req.InputPolicy)
	if err != nil {
		LOG.Error("Failed to set input policy")
		return fmt.Errorf("failed to set input policy %v", err)
	}

	// 更新并激活
	return s.updateThenActivate(hostID, newConfContent)
}

func setInputPolicy(confContent string, newPolicy string) (string, error) {
	lines := strings.Split(confContent, "\n")
	var output []string
	insideInput := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 检测进入 input chain 块
		if strings.HasPrefix(trimmed, "chain input") {
			insideInput = true
			output = append(output, line)
			continue
		}

		if insideInput {
			// 结束 input chain
			if trimmed == "}" {
				insideInput = false
			}

			// 尝试匹配 policy 行
			if strings.Contains(trimmed, "hook input") {
				indent := getIndent(line)

				if strings.Contains(trimmed, "policy") {
					// 已有 policy → 替换
					beforePolicy := strings.Split(trimmed, "policy")[0]
					newLine := indent + strings.TrimSpace(beforePolicy) + " policy " + newPolicy + ";"
					output = append(output, newLine)
				} else {
					// 没有 policy → 添加
					newLine := indent + trimmed + " policy " + newPolicy + ";"
					output = append(output, newLine)
				}
				continue
			}
		}

		// 默认保留原始行
		output = append(output, line)
	}

	return strings.Join(output, "\n"), nil
}

func (s *NFTable) getPorts(hostID uint) (*model.PageResult, error) {
	var result model.PageResult

	// 获取 /etc/nftables.conf 内容
	detail, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return &result, err
	}

	scanner := bufio.NewScanner(strings.NewReader(detail))
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	rules := parseNftRules(lines)

	result.Total = int64(len(rules))
	result.Items = rules

	return &result, nil
}

func parseNftRules(lines []string) []model.PortRule {
	rulesByRange := map[string]*model.PortRule{}

	for _, line := range lines {
		ruleItems, err := parseRuleLine(line)
		if err != nil || len(ruleItems) == 0 {
			continue
		}

		for _, ri := range ruleItems {
			key := fmt.Sprintf("%d-%d", ri.PortStart, ri.PortEnd)
			if _, exists := rulesByRange[key]; !exists {
				rulesByRange[key] = &model.PortRule{
					Protocol:  "tcp",
					PortStart: ri.PortStart,
					PortEnd:   ri.PortEnd,
					Rules:     []model.RuleItem{},
				}
			}
			rulesByRange[key].Rules = append(rulesByRange[key].Rules, ri.Rule)
		}
	}

	result := []model.PortRule{}
	for _, r := range rulesByRange {
		result = append(result, *r)
	}
	return result
}

// RuleParseResult 用于 parseRuleLine 返回多端口规则
type RuleParseResult struct {
	PortStart int
	PortEnd   int
	Rule      model.RuleItem
}

func parseRuleLine(line string) ([]RuleParseResult, error) {
	if !strings.HasPrefix(line, "tcp dport") {
		return nil, nil
	}

	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, errors.New("invalid rule format")
	}

	action := extractAction(line)
	var ruleType string
	var rate string
	var count int
	if strings.Contains(line, "limit rate") {
		ruleType = model.RuleRateLimit
		rate = extractRate(line)
	} else if strings.Contains(line, "ct count") && strings.Contains(line, "over") {
		ruleType = model.RuleConcurrentLimit
		count = extractCount(line)
	} else {
		ruleType = model.RuleDefault
	}

	ri := model.RuleItem{
		Type:   ruleType,
		Rate:   rate,
		Count:  count,
		Action: action,
	}

	results := []RuleParseResult{}

	// 解析端口表达式
	portExpr := parts[2]
	if strings.HasPrefix(portExpr, "{") && strings.HasSuffix(portExpr, "}") {
		// 多端口集合 {8080,8081,8082} 或 {8080-8085,8090}
		portExpr = strings.Trim(portExpr, "{}")
		portParts := strings.Split(portExpr, ",")
		for _, p := range portParts {
			p = strings.TrimSpace(p)
			if strings.Contains(p, "-") {
				rangeParts := strings.Split(p, "-")
				if len(rangeParts) != 2 {
					continue
				}
				start, err1 := strconv.Atoi(rangeParts[0])
				end, err2 := strconv.Atoi(rangeParts[1])
				if err1 == nil && err2 == nil {
					results = append(results, RuleParseResult{PortStart: start, PortEnd: end, Rule: ri})
				}
			} else {
				port, err := strconv.Atoi(p)
				if err == nil {
					results = append(results, RuleParseResult{PortStart: port, PortEnd: port, Rule: ri})
				}
			}
		}
	} else if strings.Contains(portExpr, "-") {
		// 端口段 8080-8090
		ports := strings.Split(portExpr, "-")
		if len(ports) == 2 {
			start, err1 := strconv.Atoi(ports[0])
			end, err2 := strconv.Atoi(ports[1])
			if err1 == nil && err2 == nil {
				results = append(results, RuleParseResult{PortStart: start, PortEnd: end, Rule: ri})
			}
		}
	} else {
		// 单端口
		port, err := strconv.Atoi(portExpr)
		if err == nil {
			results = append(results, RuleParseResult{PortStart: port, PortEnd: port, Rule: ri})
		}
	}

	return results, nil
}

func extractRate(line string) string {
	re := regexp.MustCompile(`limit rate (\S+)`)
	match := re.FindStringSubmatch(line)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func extractCount(line string) int {
	re := regexp.MustCompile(`over (\d+)`)
	match := re.FindStringSubmatch(line)
	if len(match) > 1 {
		count, _ := strconv.Atoi(match[1])
		return count
	}
	return 0
}

func extractAction(line string) string {
	if strings.HasSuffix(line, " accept") {
		return "accept"
	} else if strings.HasSuffix(line, " drop") {
		return "drop"
	} else if strings.HasSuffix(line, " reject") {
		return "reject"
	}
	return ""
}

func (s *NFTable) setPortRules(hostID uint, req model.SetPortRule) error {
	// 获取 /etc/nftables.conf 内容
	confContent, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return fmt.Errorf("failed to get conf detail %v", err)
	}

	newConfContent, err := updatePortRuleInConfContent(
		confContent,
		model.PortRule{
			Protocol:  "tcp",
			PortStart: req.PortStart,
			PortEnd:   req.PortEnd,
			Rules:     req.Rules,
		},
	)
	if err != nil {
		LOG.Error("Failed to update conf content")
		return fmt.Errorf("failed to update conf content %v", err)
	}

	// 更新并激活
	return s.updateThenActivate(hostID, newConfContent)
}

func updatePortRuleInConfContent(confContent string, newRule model.PortRule) (string, error) {
	lines := strings.Split(confContent, "\n")
	var output []string
	insideChain := false

	// 生成新的规则行
	newLines := generateNftRules([]model.PortRule{newRule})

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "chain input") {
			insideChain = true
			output = append(output, line)
			continue
		}

		if insideChain {
			// 1. 集合规则处理：发现 { } 就拆分
			if strings.Contains(trimmed, "tcp dport {") {
				parsed, _ := parseRuleLine(trimmed)
				if len(parsed) > 0 {
					// 转换成 PortRule，再生成单端口/端口段行
					var splitRules []model.PortRule
					for _, pr := range parsed {
						splitRules = append(splitRules, model.PortRule{
							Protocol:  "tcp",
							PortStart: pr.PortStart,
							PortEnd:   pr.PortEnd,
							Rules:     []model.RuleItem{pr.Rule},
						})
					}
					for _, nl := range generateNftRules(splitRules) {
						output = append(output, "        "+nl)
					}
				}
				continue
			}

			// 2. 正常替换逻辑：删除与 newRule 相同端口范围的旧行
			portPattern := fmt.Sprintf(
				`tcp dport\s+(\{[^}]*\}|%d|%d-%d)(\s|$)`,
				newRule.PortStart, newRule.PortStart, newRule.PortEnd,
			)
			matched, _ := regexp.MatchString(portPattern, trimmed)
			if matched {
				continue
			}

			// 3. 在 chain input 结束时，插入新的规则
			if trimmed == "}" {
				for _, nl := range newLines {
					output = append(output, "        "+nl)
				}
				output = append(output, line)
				insideChain = false
				continue
			}
		}

		output = append(output, line)
	}

	return strings.Join(output, "\n"), nil
}

func generateNftRules(rules []model.PortRule) []string {
	var output []string
	for _, portRule := range rules {
		var portExpr string
		if portRule.PortStart == portRule.PortEnd {
			portExpr = fmt.Sprintf("%d", portRule.PortStart)
		} else {
			portExpr = fmt.Sprintf("%d-%d", portRule.PortStart, portRule.PortEnd)
		}

		for _, rule := range portRule.Rules {
			var line string
			switch rule.Type {
			case model.RuleRateLimit:
				line = fmt.Sprintf("tcp dport %s ip saddr limit rate %s %s", portExpr, rule.Rate, rule.Action)
			case model.RuleConcurrentLimit:
				line = fmt.Sprintf("tcp dport %s ct count ip saddr over %d %s", portExpr, rule.Count, rule.Action)
			case model.RuleDefault:
				line = fmt.Sprintf("tcp dport %s %s", portExpr, rule.Action)
			}
			output = append(output, line)
		}
	}
	return output
}

func (s *NFTable) deletePortRules(hostID uint, portStart, portEnd uint) error {
	// 获取 /etc/nftables.conf 内容
	confContent, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return fmt.Errorf("failed to get conf detail %v", err)
	}

	// 删除指定端口或端口段的规则
	newConfContent, err := deletePortRuleInConf(confContent, portStart, portEnd)
	if err != nil {
		LOG.Error("Failed to update conf content")
		return fmt.Errorf("failed to update conf content %v", err)
	}

	// 更新并激活
	return s.updateThenActivate(hostID, newConfContent)
}

func deletePortRuleInConf(confContent string, portStart, portEnd uint) (string, error) {
	lines := strings.Split(confContent, "\n")
	var output []string
	insideChain := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "chain input") {
			insideChain = true
			output = append(output, line)
			continue
		}

		if insideChain {
			var portPattern string
			if portEnd == 0 || portEnd == portStart {
				// 单端口
				portPattern = fmt.Sprintf(`tcp dport %d(\s|$)`, portStart)
			} else {
				// 端口段
				portPattern = fmt.Sprintf(`tcp dport %d-%d(\s|$)`, portStart, portEnd)
			}

			matched, _ := regexp.MatchString(portPattern, trimmed)
			if matched {
				continue
			}

			if trimmed == "}" {
				output = append(output, line)
				insideChain = false
				continue
			}
		}

		output = append(output, line)
	}

	return strings.Join(output, "\n"), nil
}

func (s *NFTable) getIpBlacklist(hostID uint) (*model.PageResult, error) {
	var result model.PageResult
	// 获取 /etc/nftables.conf 内容
	detail, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return &result, err
	}
	scanner := bufio.NewScanner(strings.NewReader(detail))
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	blacklist := parseNftBlacklist(lines)
	result.Total = int64(len(blacklist))
	result.Items = blacklist
	return &result, nil
}

func parseNftBlacklist(lines []string) []string {
	var blacklist []string
	insideInputChain := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "chain input") {
			insideInputChain = true
			continue
		}
		if trimmed == "}" && insideInputChain {
			insideInputChain = false
			continue
		}
		if insideInputChain &&
			strings.HasPrefix(trimmed, "ip saddr") &&
			strings.Contains(trimmed, "drop") &&
			!strings.Contains(trimmed, "tcp dport") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				blacklist = append(blacklist, parts[2])
			}
		}
	}
	return blacklist
}

func (s *NFTable) addIPToBlacklist(hostID uint, req model.IPRequest) error {
	// 获取 /etc/nftables.conf 内容
	confContent, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return fmt.Errorf("failed to get conf detail %v", err)
	}

	// 添加 IP 到黑名单
	newConfContent, err := addBlacklistIP(confContent, req.IP)
	if err != nil {
		LOG.Error("Failed to update conf content")
		return fmt.Errorf("failed to update conf content %v", err)
	}

	// 更新并激活
	return s.updateThenActivate(hostID, newConfContent)
}

func addBlacklistIP(confContent string, ip string) (string, error) {
	lines := strings.Split(confContent, "\n")
	var output []string
	insideChain := false
	inserted := false
	ruleLine := fmt.Sprintf("ip saddr %s drop", ip)

	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		// 是否已经存在该规则
		if trimmed == ruleLine {
			return "", errors.New("ip already exists in blacklist")
		}

		output = append(output, lines[i])

		// 进入 input chain 区块
		if strings.HasPrefix(trimmed, "chain input") {
			insideChain = true
			continue
		}

		// 准备插入前，检查是否在 input chain 的末尾
		if insideChain && trimmed == "}" {
			// 插入规则前，删除最后一个 "}"，并插入规则及花括号
			output = output[:len(output)-1]
			output = append(output, "        "+ruleLine)
			output = append(output, "    }")
			inserted = true
			insideChain = false // 插入完毕，重置状态
		}
	}

	if !inserted {
		return "", errors.New("input chain not found")
	}

	return strings.Join(output, "\n"), nil
}

func (s *NFTable) deleteIPFromBlacklist(hostID uint, req model.IPRequest) error {
	// 获取 /etc/nftables.conf 内容
	confContent, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return fmt.Errorf("failed to get conf detail %v", err)
	}

	// 删除 IP 黑名单
	newConfContent, err := removeBlacklistIP(confContent, req.IP)
	if err != nil {
		LOG.Error("Failed to update conf content")
		return fmt.Errorf("failed to update conf content %v", err)
	}

	// 更新并激活
	return s.updateThenActivate(hostID, newConfContent)
}

func removeBlacklistIP(confContent string, ip string) (string, error) {
	lines := strings.Split(confContent, "\n")
	var output []string
	insideChain := false
	ruleLine := fmt.Sprintf("ip saddr %s drop", ip)
	removed := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "chain input") {
			insideChain = true
			output = append(output, line)
			continue
		}

		if insideChain {
			if trimmed == ruleLine {
				// 跳过这一行，实现删除
				removed = true
				continue
			}
			if trimmed == "}" {
				insideChain = false
			}
		}

		output = append(output, line)
	}

	if !removed {
		return "", errors.New("blacklist rule not found")
	}

	return strings.Join(output, "\n"), nil
}

func (s *NFTable) getPingStatus(hostID uint) (*model.PingStatus, error) {
	var result model.PingStatus
	// 获取 /etc/nftables.conf 内容
	detail, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return &result, err
	}
	scanner := bufio.NewScanner(strings.NewReader(detail))
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	allowed := parsePingStatus(lines)
	result.Allowed = allowed
	return &result, nil
}

func parsePingStatus(lines []string) bool {
	insideInput := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "chain input") {
			insideInput = true
			continue
		}

		if insideInput {
			if trimmed == "}" {
				insideInput = false
				continue
			}

			if strings.Contains(trimmed, "icmp type echo-request") &&
				(strings.Contains(trimmed, "drop") || strings.Contains(trimmed, "reject")) {
				return false
			}
		}
	}

	return true
}

func (s *NFTable) setPingStatus(hostID uint, allowed bool) error {
	// 获取 /etc/nftables.conf 内容
	confContent, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return fmt.Errorf("failed to get conf detail %v", err)
	}

	// 设置 ping
	newConfContent, err := setPingStatus(confContent, allowed)
	if err != nil {
		LOG.Error("Failed to update conf content")
		return fmt.Errorf("failed to update conf content %v", err)
	}

	// 更新并激活
	return s.updateThenActivate(hostID, newConfContent)
}

func setPingStatus(confContent string, allowed bool) (string, error) {
	lines := strings.Split(confContent, "\n")
	var output []string
	insideInput := false
	ruleExists := false

	// 使用正则匹配 drop ping 行
	pingDropPattern := regexp.MustCompile(`(?i)\bicmp type echo-request\b.*\b(drop|reject)\b`)

	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		// 进入 chain input 区块
		if strings.HasPrefix(trimmed, "chain input") {
			insideInput = true
			output = append(output, lines[i])
			continue
		}

		if insideInput {
			if trimmed == "}" {
				// 退出 input 区块
				if !allowed && !ruleExists {
					// 获取上一行的缩进（如果有的话）
					indent := detectIndent(lines, i)
					output = append(output, indent+"ip protocol icmp icmp type echo-request drop")
				}
				insideInput = false
				output = append(output, lines[i])
				continue
			}

			if pingDropPattern.MatchString(trimmed) {
				ruleExists = true
				if allowed {
					// 跳过该行以删除
					continue
				}
			}
		}

		// 默认加入当前行
		output = append(output, lines[i])
	}

	return strings.Join(output, "\n"), nil
}

// 辅助函数：尽量从前一行推测缩进
func detectIndent(lines []string, currentIndex int) string {
	if currentIndex > 0 {
		line := lines[currentIndex-1]
		for i, ch := range line {
			if ch != ' ' && ch != '\t' {
				return line[:i]
			}
		}
	}
	// 默认 8 空格
	return "        "
}

// 提取原行的缩进（空格或 tab）
func getIndent(line string) string {
	for i, ch := range line {
		if ch != ' ' && ch != '\t' {
			return line[:i]
		}
	}
	return ""
}

func (s *NFTable) getConfRaw(hostID uint) (*model.ConfRaw, error) {
	var result model.ConfRaw
	// 获取 /etc/nftables.conf 内容
	detail, err := s.fileContent(hostID, "/etc/nftables.conf")
	if err != nil {
		LOG.Error("Failed to get conf detail")
		return &result, err
	}
	result.Content = detail
	return &result, nil
}

func (s *NFTable) setConfRaw(hostID uint, req model.ConfRaw) error {
	return s.updateThenActivate(hostID, req.Content)
}

func (s *NFTable) updateThenActivate(hostID uint, newConfContent string) error {
	// 检查content
	safeContent, err := s.checkConfContent(newConfContent)
	if err != nil {
		return err
	}

	// Step1: 覆盖 /etc/nftables.conf内容
	editFile := model.FileEdit{
		Source:  "/etc/nftables.conf",
		Content: safeContent,
	}
	err = s.updateFile(uint64(hostID), editFile)
	if err != nil {
		LOG.Error("Failed to update nftables conf")
		return fmt.Errorf("failed to update nftables conf %v", err)
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
		return errors.New("test failed")
	}

	// 生效前，清理idb-filter表
	command = "nft delete table inet idb-filter"
	commandResult, err = s.sendCommand(uint(hostID), command)
	if err != nil {
		LOG.Error("Failed to delete table idb-filter")
		return err
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
		return errors.New("enable failed")
	}

	// Step2: 更新 /local/default/default.nftable
	repoPath := filepath.Join(s.pluginConf.Items.WorkDir, "local")
	relativePath := "default/default.nftable"

	// 检查repo
	err = s.checkRepo(hostID, repoPath)
	if err != nil {
		return err
	}

	// 更新
	gitUpdate := model.GitUpdate{
		HostID:          hostID,
		RepoPath:        repoPath,
		RelativePath:    relativePath,
		NewRelativePath: "",
		Dir:             false,
		Content:         safeContent,
	}
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
		LOG.Error("failed to send action Git_Update")
		return fmt.Errorf("failed to send action Git_Update %v", err)
	}
	if !actionResponse.Data.Action.Result {
		LOG.Error("failed to update conf file")
		return errors.New("failed to update conf file")
	}

	return nil
}
