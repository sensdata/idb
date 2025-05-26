package serviceman

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

func parseServiceBytesToServiceForm(serviceBytes []byte, standardFormFields []model.FormField) (model.ServiceForm, error) {
	var serviceForm model.ServiceForm

	if len(serviceBytes) == 0 {
		LOG.Error("Invalid service file bytes")
		return serviceForm, fmt.Errorf("invalid service bytes")
	}

	// 将 standardFormFields 的 key 存入一个集合中以便快速查找
	validKeys := make(map[string]model.FormField)
	for _, field := range standardFormFields {
		validKeys[field.Key] = field
	}

	// 按行解析 serviceBytes
	lines := strings.Split(string(serviceBytes), "\n")
	for _, line := range lines {
		// 跳过空行和注释
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 跳过[Unit]这种行
		if line[0] == '[' && line[len(line)-1] == ']' {
			continue
		}

		// 分割 key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // 如果格式不正确，跳过
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 检查 key 是否在 validKeys 中
		if formField, exists := validKeys[key]; exists {
			var serviceFormField model.ServiceFormField
			if err := copier.Copy(&serviceFormField, &formField); err != nil {
				LOG.Error("Failed to copy formFields to serviceFormField key=%s, value=%s, error=%v", key, value, err)
				continue
			}
			serviceFormField.Value = value // 设置当前值
			serviceForm.Fields = append(serviceForm.Fields, serviceFormField)
		}
	}

	return serviceForm, nil
}

func replaceValuesInServiceBytes(serviceBytes []byte, keyValues []model.KeyValue) (string, error) {
	var newLines []string

	// 构建 keyValues 的查找表
	keyValuesMap := make(map[string]string)
	for _, field := range keyValues {
		keyValuesMap[field.Key] = field.Value
	}

	// 按行解析 serviceBytes
	lines := strings.Split(string(serviceBytes), "\n")
	for _, line := range lines {
		// 跳过空行和注释
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			newLines = append(newLines, line)
			continue
		}

		// 跳过类似 "[Unit]" 的节定义行
		if line[0] == '[' && line[len(line)-1] == ']' {
			newLines = append(newLines, line)
			continue
		}

		// 分割 key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// 如果格式不正确，直接添加原始行
			newLines = append(newLines, line)
			continue
		}

		key := strings.TrimSpace(parts[0])

		// 检查 key 是否在 keyValuesMap 中
		if newValue, exists := keyValuesMap[key]; exists {
			// 如果 key 存在于 keyValuesMap 中，用新值替换
			newLine := fmt.Sprintf("%s=%s", key, newValue)
			newLines = append(newLines, newLine)
		} else {
			// 如果 key 不存在于 keyValuesMap 中，保留原始行
			newLines = append(newLines, line)
		}
	}

	// 将 newLines 转换成单个字符串
	newContent := strings.Join(newLines, "\n")
	return newContent, nil
}

func (s *ServiceMan) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
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

func (s *ServiceMan) sendCommand(hostId uint, command string) (*model.CommandResult, error) {
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

func (s *ServiceMan) checkRepo(hostID uint, repoPath string) error {
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

func (s *ServiceMan) handleHostID(reqType string, hostID uint64) (uint, error) {
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

func (s *ServiceMan) createFile(hostID uint64, op model.FileCreate) error {
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

func (s *ServiceMan) deleteFile(hostID uint64, op model.FileDelete) error {
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

func (s *ServiceMan) getCategories(hostID uint64, req model.QueryGitFile) (*model.PageResult, error) {
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

func (s *ServiceMan) createCategory(hostID uint64, req model.CreateGitCategory) error {
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

func (s *ServiceMan) updateCategory(hostID uint64, req model.UpdateGitCategory) error {
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

func (s *ServiceMan) deleteCategory(hostID uint64, req model.DeleteGitCategory) error {
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

func (s *ServiceMan) getServiceList(hostID uint64, req model.QueryGitFile) (*model.PageResult, error) {
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
		Extension:    ".service", //筛选.service
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
		return &pageResult, fmt.Errorf("failed to get service list")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to service list: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *ServiceMan) getForm(hostID uint64, req model.GetGitFileDetail) (*model.ServiceForm, error) {
	// If name is empty, return template data
	if req.Name == "" {
		return &s.templateServiceForm, nil
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
		relativePath = filepath.Join(req.Category, req.Name+".service")
	} else {
		relativePath = req.Name + ".service"
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
		return nil, fmt.Errorf("failed to get service detail")
	}

	var gitFile model.GitFile
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &gitFile)
	if err != nil {
		LOG.Error("Error unmarshaling data to service detail: %v", err)
		return nil, fmt.Errorf("json err: %v", err)
	}

	serviceForm, err := parseServiceBytesToServiceForm([]byte(gitFile.Content), s.form.Fields)
	if err != nil {
		LOG.Error("Failed to parse service content: %v", err)
		return nil, fmt.Errorf("failed to parse service content")
	}

	return &serviceForm, nil
}

func (s *ServiceMan) createForm(hostID uint64, req model.CreateServiceForm) error {
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
				}
				// 设置了长度限制
				if formField.Validation.MinLength >= 0 && formField.Validation.MaxLength >= formField.Validation.MinLength {
					if len(item.Value) < formField.Validation.MinLength || len(item.Value) > formField.Validation.MaxLength {
						LOG.Error("Value %s does not has valid length for key %s", item.Value, item.Key)
						return fmt.Errorf("invalid value for key %s", item.Key)
					}
				}
			}
		} else {
			// 不存在，返回错误
			LOG.Error("Invalid form key: %s", item.Key)
			return fmt.Errorf("invalid key: %s", item.Key)
		}
	}

	// service内容: templateService
	serviceBytes := templateService
	// 将service内容中的相关字段替换value
	newContent, err := replaceValuesInServiceBytes(serviceBytes, req.Form)
	if err != nil {
		LOG.Error("Failed to replace service content: %v", err)
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
		relativePath = filepath.Join(req.Category, req.Name+".service")
	} else {
		relativePath = req.Name + ".service"
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
		return fmt.Errorf("failed to get create service file")
	}

	return nil
}

func (s *ServiceMan) updateForm(hostID uint64, req model.UpdateServiceForm) error {
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
				}
				// 设置了长度限制
				if formField.Validation.MinLength >= 0 && formField.Validation.MaxLength >= formField.Validation.MinLength {
					if len(item.Value) < formField.Validation.MinLength || len(item.Value) > formField.Validation.MaxLength {
						LOG.Error("Value %s does not has valid length for key %s", item.Value, item.Key)
						return fmt.Errorf("invalid value for key %s", item.Key)
					}
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
		relativePath = filepath.Join(req.Category, req.Name+".service")
	} else {
		relativePath = req.Name + ".service"
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
		newRelativePath = filepath.Join(newCategory, newName+".service")
	} else {
		newRelativePath = newName + ".service"
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
		LOG.Error("Error unmarshaling data to service detail: %v", err)
		return err
	}

	// service内容
	serviceBytes := []byte(gitFile.Content)
	// 将service内容中的相关字段替换value
	newContent, err := replaceValuesInServiceBytes(serviceBytes, req.Form)
	if err != nil {
		LOG.Error("Failed to replace service content: %v", err)
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
		return fmt.Errorf("failed to update service file")
	}

	return nil
}

func (s *ServiceMan) create(hostID uint64, req model.CreateGitFile, extension string) error {
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
		return fmt.Errorf("failed to get create service file")
	}

	return nil
}

func (s *ServiceMan) getContent(hostID uint64, req model.GetGitFileDetail) (*model.GitFile, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".service")
	} else {
		relativePath = req.Name + ".service"
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
		return nil, fmt.Errorf("failed to get service detail")
	}

	var gitFile model.GitFile
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &gitFile)
	if err != nil {
		LOG.Error("Error unmarshaling data to service detail: %v", err)
		return nil, fmt.Errorf("json err: %v", err)
	}

	return &gitFile, nil
}

func (s *ServiceMan) update(hostID uint64, req model.UpdateGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".service")
	} else {
		relativePath = req.Name + ".service"
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
		newRelativePath = filepath.Join(newCategory, newName+".service")
	} else {
		newRelativePath = newName + ".service"
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
		return fmt.Errorf("failed to update service file")
	}

	return nil
}

func (s *ServiceMan) delete(hostID uint64, req model.DeleteGitFile, extension string) error {
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
		return fmt.Errorf("failed to delete service file")
	}

	return nil
}

func (s *ServiceMan) restore(hostID uint64, req model.RestoreGitFile) error {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".service")
	} else {
		relativePath = req.Name + ".service"
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
		return fmt.Errorf("failed to restore service file")
	}

	return nil
}

func (s *ServiceMan) getServiceLog(hostID uint64, req model.GitFileLog) (*model.PageResult, error) {
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
		relativePath = filepath.Join(req.Category, req.Name+".service")
	} else {
		relativePath = req.Name + ".service"
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
		return &pageResult, fmt.Errorf("failed to get service logs")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		LOG.Error("Error unmarshaling data to service logs: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}

func (s *ServiceMan) getServiceDiff(hostID uint64, req model.GitFileDiff) (string, error) {
	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".service")
	} else {
		relativePath = req.Name + ".service"
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
		return "", fmt.Errorf("failed to get service diff")
	}

	return actionResponse.Data.Action.Data, nil
}

func (s *ServiceMan) syncGlobal(hostID uint) error {
	LOG.Info("Start syncing global services for host %d", hostID)

	defaultHost, err := s.hostRepo.Get(s.hostRepo.WithByDefault())
	if err != nil {
		LOG.Error("Failed to get default host: %v", err)
		return err
	}
	if hostID == defaultHost.ID {
		LOG.Error("Attempting to sync global services on default host (ID: %d)", hostID)
		return fmt.Errorf("can't sync global services in default host")
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
	remoteUrl := fmt.Sprintf("%s://%s:%d/api/v1/git/services/global", scheme, host, settingInfo.BindPort)
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
		return fmt.Errorf("failed to sync global services")
	}

	LOG.Info("Successfully synced global services for host %d", hostID)
	return nil
}

func (s *ServiceMan) serviceActivate(hostID uint64, req model.ServiceAction) error {

	var repoPath string
	switch req.Type {
	case "global":
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "global")
	default:
		repoPath = filepath.Join(s.pluginConf.Items.WorkDir, "local")
	}
	var relativePath string
	if req.Category != "" {
		relativePath = filepath.Join(req.Category, req.Name+".service")
	} else {
		relativePath = req.Name + ".service"
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

	// service file path
	servicePath := filepath.Join(repoPath, relativePath)
	// service file name
	serviceName := filepath.Base(servicePath)
	// service link file path
	serviceLinkPath := filepath.Join("/etc/systemd/system", serviceName)

	switch req.Action {
	case "activate":
		// 创建服务链接 /etc/systemd/system/source -> LinkPath
		createFile := model.FileCreate{
			Source:    serviceLinkPath,
			IsLink:    true,
			IsSymlink: true,
			LinkPath:  servicePath,
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

		// systemctl daemon-reload
		command := "systemctl daemon-reload"
		_, err = s.sendCommand(hid, command)
		if err != nil {
			LOG.Error("Failed to reload daemon")
			return err
		}

	case "deactivate":
		// 删除务链接
		deleteFile := model.FileDelete{
			Path: serviceLinkPath,
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

		// systemctl daemon-reload
		command := "systemctl daemon-reload"
		_, err = s.sendCommand(hid, command)
		if err != nil {
			LOG.Error("Failed to reload daemon")
			return err
		}

	default:
		return errors.New("unsupported action")
	}

	return nil
}
