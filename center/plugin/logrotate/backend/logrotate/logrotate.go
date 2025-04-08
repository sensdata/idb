package logrotate

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/sensdata/idb/center/core/api/service"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func parseConfBytesToServiceForm(confBytes []byte, standardFormFields []model.FormField) (model.ServiceForm, error) {
	var serviceForm model.ServiceForm

	if len(confBytes) == 0 {
		LOG.Error("Invalid conf file bytes")
		return serviceForm, fmt.Errorf("invalid conf bytes")
	}

	boolOptions := make(map[string]model.FormField)
	var pathOption model.FormField
	var frequencyOption model.FormField
	var rotateCountOption model.FormField
	var createOption model.FormField
	var preRotateOption model.FormField
	var postRotateOption model.FormField
	for _, field := range standardFormFields {
		downCaseKey := strings.ToLower(field.Key)
		switch downCaseKey {
		case "path":
			pathOption = field
		case "frequency":
			frequencyOption = field
		case "count":
			rotateCountOption = field
		case "create":
			createOption = field
		case "prerotate":
			preRotateOption = field
		case "postrotate":
			postRotateOption = field
		default:
			if utils.MatchPattern(
				downCaseKey,
				"^(compress|delaycompress|missingok|notifempty)$",
			) {
				boolOptions[downCaseKey] = field
			}
		}
	}

	// 将 confBytes 转换为字符串
	confContent := string(confBytes)

	// 1. 查找选项起始符和结束符
	startIndex := strings.Index(confContent, "{")
	endIndex := strings.LastIndex(confContent, "}")

	if startIndex == -1 || endIndex == -1 || endIndex <= startIndex {
		return serviceForm, fmt.Errorf("missing option delimiters '{' or '}'")
	}

	// 2. 获取日志文件路径
	logPath := strings.TrimSpace(confContent[:startIndex])
	if logPath == "" {
		return serviceForm, fmt.Errorf("invalid log file path: %s", logPath)
	}
	var pathFormField model.ServiceFormField
	if err := copier.Copy(&pathFormField, &pathOption); err != nil {
		LOG.Error("Failed to copy frequencyOption field")
		return serviceForm, err
	}
	pathFormField.Value = logPath
	serviceForm.Fields = append(serviceForm.Fields, pathFormField)

	// 3. 截取选项内容并按行分割
	options := confContent[startIndex+1 : endIndex]
	lines := strings.Split(options, "\n")

	// 4. 解析每一行选项
	for _, line := range lines {
		option := strings.TrimSpace(line)
		// 跳过空行和注释
		if option == "" || strings.HasPrefix(option, "#") {
			continue
		}

		// 如果是频率设置
		if frequencyOption.Validation != nil && utils.MatchPattern(option, frequencyOption.Validation.Pattern) {
			var serviceFormField model.ServiceFormField
			if err := copier.Copy(&serviceFormField, &frequencyOption); err != nil {
				LOG.Error("Failed to copy frequencyOption field")
				continue
			}
			serviceFormField.Value = option
			serviceForm.Fields = append(serviceForm.Fields, serviceFormField)
			continue
		}

		// 如果是轮转数量
		if rotateCountOption.Validation != nil && utils.MatchPattern(option, rotateCountOption.Validation.Pattern) {
			var serviceFormField model.ServiceFormField
			if err := copier.Copy(&serviceFormField, &rotateCountOption); err != nil {
				LOG.Error("Failed to copy rotateCountOption field")
				continue
			}
			serviceFormField.Value = option
			serviceForm.Fields = append(serviceForm.Fields, serviceFormField)
			continue
		}

		// 如果是创建新日志文件
		if createOption.Validation != nil && utils.MatchPattern(option, createOption.Validation.Pattern) {
			var serviceFormField model.ServiceFormField
			if err := copier.Copy(&serviceFormField, &createOption); err != nil {
				LOG.Error("Failed to copy createOption field")
				continue
			}
			serviceFormField.Value = option
			serviceForm.Fields = append(serviceForm.Fields, serviceFormField)
			continue
		}

		// 如果是bool选项
		for _, boolOption := range boolOptions {
			// 这种选项，用key来承载了写入配置文件的值
			// 文件中已配置的，或者未配置的，都入列
			var serviceFormField model.ServiceFormField
			if err := copier.Copy(&serviceFormField, &boolOption); err != nil {
				LOG.Error("Failed to copy createOption field")
				continue
			}
			// 如果命中，为true
			if option == strings.ToLower(boolOption.Key) {
				serviceFormField.Value = "true"
			} else {
				serviceFormField.Value = "false"
			}
			serviceForm.Fields = append(serviceForm.Fields, serviceFormField)
		}
	}

	// 匹配 prerotate 到第一个 endscript
	preRotateCommand := ""
	prerotatePattern := `prerotate(.*?)endscript`
	prerotateRe := regexp.MustCompile(prerotatePattern)

	prerotateMatch := prerotateRe.FindStringSubmatch(confContent)
	if len(prerotateMatch) > 0 {
		preRotateCommand = strings.TrimSpace(prerotateMatch[1])
		if preRotateCommand == ":" {
			preRotateCommand = ""
		}
	}
	var prerotateFormField model.ServiceFormField
	if err := copier.Copy(&prerotateFormField, &preRotateOption); err != nil {
		LOG.Error("Failed to copy preRotateOption field")
		return serviceForm, err
	}
	prerotateFormField.Value = preRotateCommand
	serviceForm.Fields = append(serviceForm.Fields, prerotateFormField)

	// 匹配 postrotate 到第一个 endscript
	postRotateCommand := ""
	postrotatePattern := `postrotate(.*?)endscript`
	postrotateRe := regexp.MustCompile(postrotatePattern)

	postrotateMatch := postrotateRe.FindStringSubmatch(confContent)
	if len(postrotateMatch) > 0 {
		postRotateCommand = strings.TrimSpace(postrotateMatch[1])
		if postRotateCommand == ":" {
			postRotateCommand = ""
		}
	}
	var postrotateFormField model.ServiceFormField
	if err := copier.Copy(&postrotateFormField, &postRotateOption); err != nil {
		LOG.Error("Failed to copy postRotateOption field")
		return serviceForm, err
	}
	postrotateFormField.Value = postRotateCommand
	serviceForm.Fields = append(serviceForm.Fields, postrotateFormField)

	return serviceForm, nil
}

func replaceValuesInServiceBytes(confBytes []byte, keyValues []model.KeyValue, standardFormFields []model.FormField) (string, error) {
	var newLines []string

	// 构建 keyValues 的查找表
	keyValuesMap := make(map[string]string)
	for _, field := range keyValues {
		keyValuesMap[strings.ToLower(field.Key)] = field.Value
	}

	boolOptions := make(map[string]model.FormField)
	var pathOption model.FormField
	var frequencyOption model.FormField
	var rotateCountOption model.FormField
	var createOption model.FormField
	var preRotateOption model.FormField
	var postRotateOption model.FormField
	for _, field := range standardFormFields {
		downCaseKey := strings.ToLower(field.Key)
		switch downCaseKey {
		case "path":
			pathOption = field
		case "frequency":
			frequencyOption = field
		case "count":
			rotateCountOption = field
		case "create":
			createOption = field
		case "prerotate":
			preRotateOption = field
		case "postrotate":
			postRotateOption = field
		default:
			if utils.MatchPattern(
				downCaseKey,
				"^(compress|delaycompress|missingok|notifempty)$",
			) {
				boolOptions[downCaseKey] = field
			}
		}
	}

	// 将 confBytes 转换为字符串
	confContent := string(confBytes)

	// 1. 查找选项起始符和结束符
	startIndex := strings.Index(confContent, "{")
	endIndex := strings.LastIndex(confContent, "}")

	if startIndex == -1 || endIndex == -1 || endIndex <= startIndex {
		return strings.Join(newLines, "\n"), fmt.Errorf("missing option delimiters '{' or '}'")
	}

	// 2. 获取日志文件路径
	logPath := strings.TrimSpace(confContent[:startIndex])
	if logPath != "" {
		// 检查 key 是否在 keyValuesMap 中
		if newValue, exists := keyValuesMap[strings.ToLower(pathOption.Key)]; exists {
			// 如果 key 存在于 keyValuesMap 中，用新值替换
			newLines = append(newLines, newValue+"{")
		} else {
			// 如果 key 不存在于 keyValuesMap 中，保留原始行
			newLines = append(newLines, logPath+"{")
		}
	}

	// 3. 截取选项内容并按行分割
	options := confContent[startIndex+1 : endIndex]
	lines := strings.Split(options, "\n")

	// 4. 解析每一行选项
	for _, line := range lines {
		option := strings.TrimSpace(line)
		// 跳过空行和注释
		if option == "" || strings.HasPrefix(option, "#") {
			newLines = append(newLines, line)
			continue
		}

		// 如果是频率设置
		if frequencyOption.Validation != nil && utils.MatchPattern(option, frequencyOption.Validation.Pattern) {
			// 检查 key 是否在 keyValuesMap 中
			if newValue, exists := keyValuesMap[strings.ToLower(frequencyOption.Key)]; exists {
				// 如果 key 存在于 keyValuesMap 中，用新值替换
				newLines = append(newLines, strings.ReplaceAll(line, option, newValue))
			} else {
				// 如果 key 不存在于 keyValuesMap 中，保留原始行
				newLines = append(newLines, line)
			}
			continue
		}

		// 如果是轮转数量
		if rotateCountOption.Validation != nil && utils.MatchPattern(option, rotateCountOption.Validation.Pattern) {
			// 检查 key 是否在 keyValuesMap 中
			if newValue, exists := keyValuesMap[strings.ToLower(rotateCountOption.Key)]; exists {
				// 如果 key 存在于 keyValuesMap 中，用新值替换
				newLines = append(newLines, strings.ReplaceAll(line, option, newValue))
			} else {
				// 如果 key 不存在于 keyValuesMap 中，保留原始行
				newLines = append(newLines, line)
			}
			continue
		}

		// 如果是创建新日志文件
		if createOption.Validation != nil && utils.MatchPattern(option, createOption.Validation.Pattern) {
			// 检查 key 是否在 keyValuesMap 中
			if newValue, exists := keyValuesMap[strings.ToLower(createOption.Key)]; exists {
				// 如果 key 存在于 keyValuesMap 中，用新值替换
				newLines = append(newLines, strings.ReplaceAll(line, option, newValue))
			} else {
				// 如果 key 不存在于 keyValuesMap 中，保留原始行
				newLines = append(newLines, line)
			}
			continue
		}

		// 如果是bool选项
		// 这种选项，用key来承载了写入配置文件的值
		// 文件中命中了，入列
		if _, exists := boolOptions[option]; exists {
			// 检查 key 是否在 keyValuesMap 中
			if newValue, exists := keyValuesMap[option]; exists {
				// 如果是true，则保留，否则不添加
				if newValue == "true" {
					newLines = append(newLines, line)
				}
			} else {
				// 如果 key 不存在于 keyValuesMap 中，保留原始行
				newLines = append(newLines, line)
			}
			continue
		}
	}

	// 匹配 prerotate 到第一个 endscript
	prerotatePattern := `prerotate(.*?)endscript`
	prerotateRe := regexp.MustCompile(prerotatePattern)
	prerotateMatch := prerotateRe.FindStringSubmatch(confContent)
	if len(prerotateMatch) > 0 {
		preRotateCommand := strings.TrimSpace(prerotateMatch[1])
		newLines = append(newLines, "prerotate")
		// 检查 key 是否在 keyValuesMap 中
		if newValue, exists := keyValuesMap[strings.ToLower(preRotateOption.Key)]; exists {
			// 如果 key 存在于 keyValuesMap 中，用新值替换
			if newValue == "" {
				newLines = append(newLines, ":")
			} else {
				newLines = append(newLines, newValue)
			}
		} else {
			// 如果 key 不存在于 keyValuesMap 中，保留原始行
			newLines = append(newLines, preRotateCommand)
		}
		newLines = append(newLines, "endscript")
	}

	// 匹配 postrotate 到第一个 endscript
	postrotatePattern := `postrotate(.*?)endscript`
	postrotateRe := regexp.MustCompile(postrotatePattern)
	postrotateMatch := postrotateRe.FindStringSubmatch(confContent)
	if len(postrotateMatch) > 0 {
		postRotateCommand := strings.TrimSpace(postrotateMatch[1])
		newLines = append(newLines, "postrotate")
		// 检查 key 是否在 keyValuesMap 中
		if newValue, exists := keyValuesMap[strings.ToLower(postRotateOption.Key)]; exists {
			// 如果 key 存在于 keyValuesMap 中，用新值替换
			if newValue == "" {
				newLines = append(newLines, ":")
			} else {
				newLines = append(newLines, newValue)
			}
		} else {
			// 如果 key 不存在于 keyValuesMap 中，保留原始行
			newLines = append(newLines, postRotateCommand)
		}
		newLines = append(newLines, "endscript")
	}

	// 结尾
	newLines = append(newLines, "}")
	// 选项行添加空格
	for i := 1; i < len(newLines)-1; i++ {
		newLine := strings.TrimSpace(newLines[i])
		newLines[i] = "    " + newLine // 添加四个空格
	}

	// 将 newLines 转换成单个字符串
	newContent := strings.Join(newLines, "\n")
	return newContent, nil
}

func (s *LogRotate) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
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

func (s *LogRotate) sendCommand(hostId uint, command string) (*model.CommandResult, error) {
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

func (s *LogRotate) checkRepo(hostID uint, repoPath string) error {
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

func (s *LogRotate) handleHostID(reqType string, hostID uint64) (uint, error) {
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

func (s *LogRotate) createFile(hostID uint64, op model.FileCreate) error {
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

func (s *LogRotate) deleteFile(hostID uint64, op model.FileDelete) error {
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
		global.LOG.Error("failed to delete file")
		return fmt.Errorf("failed to delete file")
	}

	return nil
}

func (s *LogRotate) getCategories(hostID uint64, req model.QueryGitFile) (*model.PageResult, error) {
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
		return &pageResult, fmt.Errorf("failed to get conf list")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf list: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *LogRotate) createCategory(hostID uint64, req model.CreateGitCategory) error {
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
		return fmt.Errorf("failed to get create conf file")
	}

	return nil
}

func (s *LogRotate) updateCategory(hostID uint64, req model.UpdateGitCategory) error {
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
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: req.Category,
		Dir:          true,
		Content:      "",
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
		return fmt.Errorf("failed to update conf file")
	}

	return nil
}

func (s *LogRotate) deleteCategory(hostID uint64, req model.DeleteGitCategory) error {
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
		return fmt.Errorf("failed to delete conf file")
	}

	return nil
}

func (s *LogRotate) getConfList(hostID uint64, req model.QueryGitFile) (*model.PageResult, error) {
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
		Extension:    ".logrotate;.linked", //筛选.logrotate.linked
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
		return &pageResult, fmt.Errorf("failed to get conf list")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf list: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *LogRotate) getForm(hostID uint64, req model.GetGitFileDetail) (*model.ServiceForm, error) {
	// If name is empty, return template data
	if req.Name == "" {
		return &s.templateForm, nil
	}

	// Get content
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".logrotate")
	} else {
		relativePath = req.Name + ".logrotate"
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
		LOG.Error("Action error: %v", err)
		return nil, err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("Action failed")
		return nil, fmt.Errorf("failed to get conf detail")
	}

	var gitFile model.GitFile
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &gitFile)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf detail: %v", err)
		return nil, fmt.Errorf("json err: %v", err)
	}

	serviceForm, err := parseConfBytesToServiceForm([]byte(gitFile.Content), s.form.Fields)
	if err != nil {
		LOG.Error("Failed to parse conf content: %v", err)
		return nil, fmt.Errorf("failed to parse conf content")
	}

	return &serviceForm, nil
}

func (s *LogRotate) createForm(hostID uint64, req model.CreateServiceForm) error {
	// 判断提交的form中，有没有不合法的字段
	validKeys := make(map[string]model.FormField)
	for _, field := range s.form.Fields {
		validKeys[field.Key] = field
	}
	for _, item := range req.Form {
		// 检查 key 是否在 validKeys 中
		formField, exists := validKeys[item.Key]
		if exists {
			// 设置了校验规则
			if formField.Validation != nil {
				// 设置了正则匹配，优先正则匹配
				if formField.Validation.Pattern != "" {
					// 使用正则表达式校验
					matched, err := regexp.MatchString(formField.Validation.Pattern, item.Value)
					if err != nil {
						LOG.Error("Invalid regex pattern: %v", err)
						return fmt.Errorf("invalid regex pattern for key %s: %v", item.Key, err)
					}
					if !matched {
						LOG.Error("Value %s does not match the required pattern for key %s", item.Value, item.Key)
						return fmt.Errorf("invalid value for key %s", item.Key)
					}
					// 校验通过
					continue
				}
				// 设置了长度限制
				if formField.Validation.MinLength >= 0 && formField.Validation.MaxLength >= formField.Validation.MinLength {
					if len(item.Value) < formField.Validation.MinLength || len(item.Value) > formField.Validation.MaxLength {
						LOG.Error("Value %s does not has valid length for key %s", item.Value, item.Key)
						return fmt.Errorf("invalid value for key %s", item.Key)
					}
					// 校验通过
					continue
				}
			}
		} else {
			// 不存在，返回错误
			LOG.Error("Invalid form key: %s", item.Key)
			return fmt.Errorf("invalid key: %s", item.Key)
		}
	}

	// conf内容: templateService
	confBytes := templateService
	// 将conf内容中的相关字段替换value
	newContent, err := replaceValuesInServiceBytes(confBytes, req.Form, s.form.Fields)
	if err != nil {
		LOG.Error("Failed to replace conf content: %v", err)
		return constant.ErrInternalServer
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
		relativePath = filepath.Join(req.Category, req.Name+".logrotate")
	} else {
		relativePath = req.Name + ".logrotate"
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
		Content:      newContent,
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
		return fmt.Errorf("failed to get create conf file")
	}

	return nil
}

func (s *LogRotate) updateForm(hostID uint64, req model.UpdateServiceForm) error {
	// 判断提交的form中，有没有不合法的字段
	validKeys := make(map[string]model.FormField)
	for _, field := range s.form.Fields {
		validKeys[field.Key] = field
	}
	for _, item := range req.Form {
		// 检查 key 是否在 validKeys 中
		formField, exists := validKeys[item.Key]
		if exists {
			// 设置了校验规则
			if formField.Validation != nil {
				// 设置了正则匹配，优先正则匹配
				if formField.Validation.Pattern != "" {
					// 使用正则表达式校验
					matched, err := regexp.MatchString(formField.Validation.Pattern, item.Value)
					if err != nil {
						LOG.Error("Invalid regex pattern: %v", err)
						return fmt.Errorf("invalid regex pattern for key %s: %v", item.Key, err)
					}
					if !matched {
						LOG.Error("Value %s does not match the required pattern for key %s", item.Value, item.Key)
						return fmt.Errorf("invalid value for key %s", item.Key)
					}
					// 校验通过
					continue
				}
				// 设置了长度限制
				if formField.Validation.MinLength >= 0 && formField.Validation.MaxLength >= formField.Validation.MinLength {
					if len(item.Value) < formField.Validation.MinLength || len(item.Value) > formField.Validation.MaxLength {
						LOG.Error("Value %s does not has valid length for key %s", item.Value, item.Key)
						return fmt.Errorf("invalid value for key %s", item.Key)
					}
					// 校验通过
					continue
				}
			}
		} else {
			// 不存在，返回错误
			LOG.Error("Invalid form key: %s", item.Key)
			return fmt.Errorf("invalid key: %s", item.Key)
		}
	}

	// Get content
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".logrotate")
	} else {
		relativePath = req.Name + ".logrotate"
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	gitGetFile := model.GitGetFile{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
	}

	// 检查repo
	err = s.checkRepo(gitGetFile.HostID, gitGetFile.RepoPath)
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
		return constant.ErrInternalServer
	}

	var gitFile model.GitFile
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &gitFile)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf detail: %v", err)
		return err
	}

	// conf内容
	confBytes := []byte(gitFile.Content)
	// 将conf内容中的相关字段替换value
	newContent, err := replaceValuesInServiceBytes(confBytes, req.Form, s.form.Fields)
	if err != nil {
		LOG.Error("Failed to replace conf content: %v", err)
		return constant.ErrInternalServer
	}

	// 更新
	gitUpdate := model.GitUpdate{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		Content:      newContent,
	}
	data, err = utils.ToJSONString(gitUpdate)
	if err != nil {
		return err
	}

	actionRequest = model.HostAction{
		HostID: gitUpdate.HostID,
		Action: model.Action{
			Action: model.Git_Update,
			Data:   data,
		},
	}

	actionResponse, err = s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("action failed")
		return fmt.Errorf("failed to update conf file")
	}

	return nil
}

func (s *LogRotate) create(hostID uint64, req model.CreateGitFile, extension string) error {
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
		return fmt.Errorf("failed to get create conf file")
	}

	return nil
}

func (s *LogRotate) getContent(hostID uint64, req model.GetGitFileDetail) (*model.GitFile, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".logrotate")
	} else {
		relativePath = req.Name + ".logrotate"
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

func (s *LogRotate) update(hostID uint64, req model.UpdateGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".logrotate")
	} else {
		relativePath = req.Name + ".logrotate"
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	gitUpdate := model.GitUpdate{
		HostID:       hid,
		RepoPath:     repoPath,
		RelativePath: relativePath,
		Content:      req.Content,
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
		return fmt.Errorf("failed to update conf file")
	}

	return nil
}

func (s *LogRotate) delete(hostID uint64, req model.DeleteGitFile, extension string) error {
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
		return fmt.Errorf("failed to delete conf file")
	}

	return nil
}

func (s *LogRotate) restore(hostID uint64, req model.RestoreGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".logrotate")
	} else {
		relativePath = req.Name + ".logrotate"
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
		return fmt.Errorf("failed to restore conf file")
	}

	return nil
}

func (s *LogRotate) getConfLog(hostID uint64, req model.GitFileLog) (*model.PageResult, error) {
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
		relativePath = filepath.Join(req.Category, req.Name+".logrotate")
	} else {
		relativePath = req.Name + ".logrotate"
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
		return &pageResult, fmt.Errorf("failed to get conf logs")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf logs: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *LogRotate) getConfDiff(hostID uint64, req model.GitFileDiff) (string, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".logrotate")
	} else {
		relativePath = req.Name + ".logrotate"
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
		return "", fmt.Errorf("failed to get conf diff")
	}

	return actionResponse.Data.Action.Data, nil
}

func (s *LogRotate) syncGlobal(hostID uint) error {
	LOG.Info("Start syncing global logrotate for host %d", hostID)

	defaultHost, err := s.hostRepo.Get(s.hostRepo.WithByDefault())
	if err != nil {
		LOG.Error("Failed to get default host: %v", err)
		return err
	}
	if hostID == defaultHost.ID {
		LOG.Error("Attempting to sync global logrotate on default host (ID: %d)", hostID)
		return fmt.Errorf("can't sync global logrotate in default host")
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
	remoteUrl := fmt.Sprintf("%s://%s:%d/api/v1/git/logrotate/global", scheme, host, settingInfo.BindPort)
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
		return fmt.Errorf("failed to sync global logrotate")
	}

	LOG.Info("Successfully synced global logrotate for host %d", hostID)
	return nil
}

func (s *LogRotate) confAction(hostID uint64, req model.ServiceAction) error {

	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".logrotate")
	} else {
		relativePath = req.Name + ".logrotate"
	}

	// global的情况，操作本机
	hid, err := s.handleHostID(req.Type, hostID)
	if err != nil {
		return err
	}

	// 检查repo
	err = s.checkRepo(hid, repoPath)
	if err != nil {
		return err
	}

	// conf file path
	confPath := filepath.Join(repoPath, relativePath)
	// conf file name
	confName := filepath.Base(confPath)
	// conf link file path
	confLinkName := fmt.Sprintf("iDB_%s", strings.TrimSuffix(confName, ".logrotate")) // 链接命名使用 iDB 前缀，且不带后缀
	confLinkPath := filepath.Join("/etc/logrotate.d", confLinkName)

	switch req.Action {
	case "activate":
		// 创建服务链接 /etc/logrotate.d/source -> LinkPath
		createFile := model.FileCreate{
			Source:    confLinkPath,
			IsLink:    true,
			IsSymlink: true,
			LinkPath:  confPath,
		}
		err := s.createFile(uint64(hid), createFile)
		if err != nil {
			LOG.Error("Failed to create symlink")
			return err
		}

		// 添加.linked标记文件到仓库
		createGitFile := model.CreateGitFile{
			Type:     req.Type,
			Category: req.Category,
			Name:     req.Name,
			Content:  "",
		}
		err = s.create(uint64(hid), createGitFile, ".linked")
		if err != nil {
			LOG.Error("Failed to create linked file")
			return err
		}

		// 进行-d测试
		// TODO: 根据测试结果，做进一步的处理
		command := fmt.Sprintf("logrotate -d %s", confLinkPath)
		commandResult, err := s.sendCommand(hid, command)
		if err != nil {
			LOG.Error("Failed to test conf")
			return err
		}
		LOG.Info("Conf test result: %s", commandResult.Result)

	case "deactivate":
		// 删除务链接
		deleteFile := model.FileDelete{
			Path: confLinkPath,
		}
		err := s.deleteFile(uint64(hid), deleteFile)
		if err != nil {
			LOG.Error("Failed to delete symlink")
			return err
		}

		// 从仓库中删除.linked标记文件
		deleteGitFile := model.DeleteGitFile{
			Type:     req.Type,
			Category: req.Category,
			Name:     req.Name,
		}
		err = s.delete(uint64(hid), deleteGitFile, ".linked")
		if err != nil {
			LOG.Error("Failed to delete linked file")
			return err
		}

		// 删除的场景，不做额外测试

	default:
		return errors.New("unsupported action")
	}

	return nil
}
