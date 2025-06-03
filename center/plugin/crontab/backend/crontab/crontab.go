package crontab

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

	var ruleOption model.FormField
	for _, field := range standardFormFields {
		switch field.Key {
		case "Rule":
			ruleOption = field
		default:
		}
	}

	if len(confBytes) == 0 {
		LOG.Error("Invalid conf file bytes")
		return serviceForm, fmt.Errorf("invalid conf bytes")
	}

	// 按行解析 serviceBytes
	lines := strings.Split(string(confBytes), "\n")
	for _, line := range lines {
		rule := strings.TrimSpace(line)
		// 跳过空行和注释
		if rule == "" || strings.HasPrefix(rule, "#") {
			continue
		}

		// 匹配规则行
		if ruleOption.Validation != nil && utils.MatchPattern(rule, ruleOption.Validation.Pattern) {
			var ruleFormField model.ServiceFormField
			if err := copier.Copy(&ruleFormField, &ruleOption); err != nil {
				LOG.Error("Failed to copy ruleOption field")
				continue
			}
			ruleFormField.Value = rule
			serviceForm.Fields = append(serviceForm.Fields, ruleFormField)
		}
	}

	return serviceForm, nil
}

func buildConfContentWithForm(keyValues []model.KeyValue, standardFormFields []model.FormField) (string, error) {
	var newLines []string

	// 规则变量
	var ruleOption *model.FormField
	for _, field := range standardFormFields {
		switch field.Key {
		case "Rule":
			ruleOption = &field
		default:
		}
	}
	if ruleOption == nil {
		LOG.Error("No standard rule option found")
		return "", constant.ErrInternalServer
	}

	// 检查每个提交的项
	for _, kv := range keyValues {
		// 校验key
		if kv.Key != ruleOption.Key {
			LOG.Error("Invalid form data submited with unsupport key %s", kv.Key)
			return "", fmt.Errorf("invalid key in form")
		}

		// 校验value
		rule := strings.TrimSpace(kv.Value)
		// 匹配规则行
		if ruleOption.Validation != nil && utils.MatchPattern(rule, ruleOption.Validation.Pattern) {
			// 是规则行
			newLines = append(newLines, rule)
		}
	}

	// 将 newLines 转换成单个字符串
	newContent := strings.Join(newLines, "\n")
	return newContent, nil
}

func (s *CronTab) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
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

func (s *CronTab) sendCommand(hostId uint, command string) (*model.CommandResult, error) {
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

func (s *CronTab) checkRepo(hostID uint, repoPath string) error {
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

func (s *CronTab) handleHostID(reqType string, hostID uint64) (uint, error) {
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

func (s *CronTab) needSync(reqType string, hostID uint64) (bool, error) {
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

func (s *CronTab) createFile(hostID uint64, op model.FileCreate) error {
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

func (s *CronTab) deleteFile(hostID uint64, op model.FileDelete) error {
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

func (s *CronTab) getCategories(hostID uint64, req model.QueryGitFile) (*model.PageResult, error) {
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

func (s *CronTab) createCategory(hostID uint64, req model.CreateGitCategory) error {
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

func (s *CronTab) updateCategory(hostID uint64, req model.UpdateGitCategory) error {
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
		return fmt.Errorf("failed to update conf file")
	}

	return nil
}

func (s *CronTab) deleteCategory(hostID uint64, req model.DeleteGitCategory) error {
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

func (s *CronTab) getConfList(hostID uint64, req model.QueryGitFile) (*model.PageResult, error) {
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
		Extension:    ".crontab", //筛选.crontab
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

func (s *CronTab) getForm(hostID uint64, req model.GetGitFileDetail) (*model.ServiceForm, error) {
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
		relativePath = filepath.Join(req.Category, req.Name+".crontab")
	} else {
		relativePath = req.Name + ".crontab"
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

func (s *CronTab) createForm(hostID uint64, req model.CreateServiceForm) error {
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

	// TODO: 目前执行的是使用提交项组织内容后，进行全量覆盖
	newContent, err := buildConfContentWithForm(req.Form, s.form.Fields)
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
		relativePath = filepath.Join(req.Category, req.Name+".crontab")
	} else {
		relativePath = req.Name + ".crontab"
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

func (s *CronTab) updateForm(hostID uint64, req model.UpdateServiceForm) error {
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
		relativePath = filepath.Join(req.Category, req.Name+".crontab")
	} else {
		relativePath = req.Name + ".crontab"
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
		newRelativePath = filepath.Join(newCategory, newName+".crontab")
	} else {
		newRelativePath = newName + ".crontab"
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

	// TODO: 目前执行的是使用提交项组织内容后，进行全量覆盖
	newContent, err := buildConfContentWithForm(req.Form, s.form.Fields)
	if err != nil {
		LOG.Error("Failed to replace conf content: %v", err)
		return constant.ErrInternalServer
	}

	// 更新
	gitUpdate := model.GitUpdate{
		HostID:          hid,
		RepoPath:        repoPath,
		RelativePath:    relativePath,
		NewRelativePath: newRelativePath,
		Dir:             false,
		Content:         newContent,
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

func (s *CronTab) create(hostID uint64, req model.CreateGitFile, extension string) error {
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

func (s *CronTab) getContent(hostID uint64, req model.GetGitFileDetail) (*model.GitFile, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".crontab")
	} else {
		relativePath = req.Name + ".crontab"
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

func (s *CronTab) update(hostID uint64, req model.UpdateGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".crontab")
	} else {
		relativePath = req.Name + ".crontab"
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
		newRelativePath = filepath.Join(newCategory, newName+".crontab")
	} else {
		newRelativePath = newName + ".crontab"
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
		return fmt.Errorf("failed to update conf file")
	}

	return nil
}

func (s *CronTab) delete(hostID uint64, req model.DeleteGitFile, extension string) error {
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

func (s *CronTab) restore(hostID uint64, req model.RestoreGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".crontab")
	} else {
		relativePath = req.Name + ".crontab"
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

func (s *CronTab) getConfLog(hostID uint64, req model.GitFileLog) (*model.PageResult, error) {
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
		relativePath = filepath.Join(req.Category, req.Name+".crontab")
	} else {
		relativePath = req.Name + ".crontab"
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

func (s *CronTab) getConfDiff(hostID uint64, req model.GitFileDiff) (string, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".crontab")
	} else {
		relativePath = req.Name + ".crontab"
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

func (s *CronTab) syncGlobal(hostID uint) error {
	LOG.Info("Start syncing global crontabs for host %d", hostID)

	defaultHost, err := s.hostRepo.Get(s.hostRepo.WithByDefault())
	if err != nil {
		LOG.Error("Failed to get default host: %v", err)
		return err
	}
	if hostID == defaultHost.ID {
		LOG.Error("Attempting to sync global crontabs on default host (ID: %d)", hostID)
		return fmt.Errorf("can't sync global crontabs in default host")
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
	remoteUrl := fmt.Sprintf("%s://%s:%d/api/v1/git/crontab/global", scheme, host, settingInfo.BindPort)
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
		return fmt.Errorf("failed to sync global crontabs")
	}

	LOG.Info("Successfully synced global crontabs for host %d", hostID)
	return nil
}

func (s *CronTab) confActivate(hostID uint64, req model.ServiceActivate) error {

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
		relativePath = filepath.Join(req.Category, req.Name+".crontab")
	} else {
		relativePath = req.Name + ".crontab"
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
		LOG.Error("Action error: %v", err)
		return err
	}

	if !actionResponse.Data.Action.Result {
		LOG.Error("Action failed")
		return fmt.Errorf("failed to get conf detail")
	}

	var gitFile model.GitFile
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &gitFile)
	if err != nil {
		LOG.Error("Error unmarshaling data to conf detail: %v", err)
		return fmt.Errorf("json err: %v", err)
	}

	// conf file path
	confPath := filepath.Join(repoPath, relativePath)
	// conf file name
	confName := filepath.Base(confPath)
	// conf link file path
	confLinkName := fmt.Sprintf("iDB_%s", strings.TrimSuffix(confName, ".crontab")) // 链接命名使用 iDB 前缀，且不带后缀
	confLinkPath := filepath.Join("/etc/cron.d", confLinkName)

	switch req.Action {
	case "activate":
		// 创建文件
		createFile := model.FileCreate{
			Source:  confLinkPath,
			Content: gitFile.Content,
		}
		err := s.createFile(uint64(hostID), createFile)
		if err != nil {
			LOG.Error("Failed to create conf file")
			return err
		}

		// 添加.linked标记文件到仓库
		createGitFile := model.CreateGitFile{
			Type:     req.Type,
			Category: req.Category,
			Name:     req.Name,
			Content:  "",
		}
		err = s.create(uint64(hostID), createGitFile, ".linked")
		if err != nil {
			LOG.Error("Failed to create linked file")
			return err
		}

	case "deactivate":
		// 删除文件
		deleteFile := model.FileDelete{
			Path: confLinkPath,
		}
		err := s.deleteFile(uint64(hostID), deleteFile)
		if err != nil {
			LOG.Error("Failed to delete conf file")
			return err
		}

		// 从仓库中删除.linked标记文件
		deleteGitFile := model.DeleteGitFile{
			Type:     req.Type,
			Category: req.Category,
			Name:     req.Name,
		}
		err = s.delete(uint64(hostID), deleteGitFile, ".linked")
		if err != nil {
			LOG.Error("Failed to delete linked file")
			return err
		}

	default:
		return errors.New("unsupported action")
	}

	return nil
}

func (s *CronTab) confOperate(hostID uint64, req model.CrontabOperate) (*model.ServiceOperateResult, error) {
	var result model.ServiceOperateResult

	// 先看是否需要同步
	needSync, err := s.needSync(req.Type, hostID)
	if err != nil {
		LOG.Error("Failed to check if sync is needed: %v", err)
		return &result, err
	}
	// 执行同步
	if needSync {
		if err := s.syncGlobal(uint(hostID)); err != nil {
			LOG.Error("Failed to sync global services: %v", err)
			return &result, err
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
		relativePath = filepath.Join(req.Category, req.Name+".crontab")
	} else {
		relativePath = req.Name + ".crontab"
	}

	// 检查repo
	err = s.checkRepo(uint(hostID), repoPath)
	if err != nil {
		return &result, err
	}

	// 执行操作
	switch req.Operation {
	case "execute":
		// 获取脚本内容
		gitGetFile := model.GitGetFile{
			HostID:       uint(hostID),
			RepoPath:     repoPath,
			RelativePath: relativePath,
		}
		data, err := utils.ToJSONString(gitGetFile)
		if err != nil {
			return &result, err
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
			LOG.Error("Failed to send action: %v", err)
			return &result, err
		}
		if !actionResponse.Data.Action.Result {
			LOG.Error("Action failed")
			return &result, fmt.Errorf("failed to get conf detail")
		}
		var gitFile model.GitFile
		err = utils.FromJSONString(actionResponse.Data.Action.Data, &gitFile)
		if err != nil {
			LOG.Error("Error unmarshaling data to conf detail: %v", err)
			return &result, fmt.Errorf("json err: %v", err)
		}
		// 从脚本提取命令
		commandLines := extractCrontabCommands(gitFile.Content)
		if len(commandLines) == 0 {
			return &result, fmt.Errorf("no commands found in the crontab file")
		}
		// 执行命令
		command := commandLines[0]
		commandResult, err := s.sendCommand(uint(hostID), command)
		if err != nil {
			LOG.Error("Failed to execute conf")
			return &result, err
		}
		result.Result = commandResult.Result
	}
	return &result, nil
}

func extractCrontabCommands(content string) []string {
	var commands []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if strings.HasPrefix(fields[0], "@") && len(fields) > 1 {
			// 兼容 @reboot、@daily 等
			commands = append(commands, strings.Join(fields[1:], " "))
		} else if len(fields) > 5 {
			// 标准时间表达式
			commands = append(commands, strings.Join(fields[5:], " "))
		}
	}
	return commands
}
