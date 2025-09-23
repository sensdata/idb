package docker

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *DockerMan) dockerInstallStatus(hostID uint64) (*model.DockerInstallStatus, error) {
	var status model.DockerInstallStatus

	// 命令行查询docker安装状态
	command := "command -v docker >/dev/null 2>&1 && echo \"installed\" || echo \"not installed\""
	commandResult, err := s.sendCommand(uint(hostID), command)
	if err != nil {
		LOG.Error("Failed to check install status")
		return &status, errors.New("failed to check install status")
	}
	LOG.Info("Install status result: %s", commandResult.Result)
	status.Status = strings.TrimSpace(commandResult.Result)
	return &status, nil
}

func (s *DockerMan) dockerInstall(hostID uint64) error {
	// 将安装脚本内容保存到一个临时文件
	logPath := filepath.Join("/tmp", "iDB_docker_install.log")
	scriptPath := fmt.Sprintf("/tmp/iDB_docker_install_%d.sh", time.Now().Unix())
	createFile := model.FileCreate{
		Source:  scriptPath,
		Content: string(installDockerShell),
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

	return nil
}

func (s *DockerMan) dockerStatus(hostID uint64) (*model.DockerStatus, error) {
	var status model.DockerStatus

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Status,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &status, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &status, fmt.Errorf("failed to get docker status")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &status)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to docker status: %v", err)
		return &status, fmt.Errorf("json err: %v", err)
	}

	return &status, nil
}

func (s *DockerMan) dockerConf(hostID uint64) (*model.DaemonJsonConf, error) {
	var conf model.DaemonJsonConf

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Conf,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &conf, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &conf, fmt.Errorf("failed to get docker conf")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &conf)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to docker conf: %v", err)
		return &conf, fmt.Errorf("json err: %v", err)
	}

	return &conf, nil
}

func (s *DockerMan) dockerConfRaw(hostID uint64) (*model.DaemonJsonUpdateRaw, error) {
	var raw model.DaemonJsonUpdateRaw

	req := model.FileContentReq{
		Path:   "/etc/docker/daemon.json",
		Expand: true,
	}
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &raw, err
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
		return &raw, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &raw, fmt.Errorf("failed to get file content")
	}

	var info model.FileInfo
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &info)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content: %v", err)
		return &raw, fmt.Errorf("json err: %v", err)
	}

	raw.Content = info.Content
	return &raw, nil
}

func (s *DockerMan) dockerUpdateConf(hostID uint64, req model.KeyValue) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Upd_Conf,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update docker conf")
	}

	return nil
}

func (s *DockerMan) dockerUpdateConfByFile(hostID uint64, req model.DaemonJsonUpdateRaw) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Upd_Conf_File,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update docker conf by file")
	}

	return nil
}

func (s *DockerMan) dockerUpdateLogOption(hostID uint64, req model.LogOption) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Upd_Log,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update log option")
	}

	return nil
}

func (s *DockerMan) dockerUpdateIpv6Option(hostID uint64, req model.Ipv6Option) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Upd_Ipv6,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to update ipv6 option")
	}

	return nil
}

func (s *DockerMan) dockerOperation(hostID uint64, req model.DockerOperation) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Operation,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to execute operation")
	}

	return nil
}

func (s *DockerMan) inspect(hostID uint64, req model.Inspect) (*model.InspectResult, error) {
	result := model.InspectResult{
		Type: req.Type,
		ID:   req.ID,
	}

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Inspect,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to inspect")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to inspect result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) prune(hostID uint64, req model.Prune) (*model.PruneResult, error) {
	var result model.PruneResult

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Prune,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to prune")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to prune result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}
