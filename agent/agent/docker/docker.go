package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sensdata/idb/agent/agent/docker/client"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
)

type DockerService struct{}

type IDockerService interface {
	DockerStatus() (*model.DockerStatus, error)
	DockerConf() (*model.DaemonJsonConf, error)
	DockerUpdateConf(req model.KeyValue) error
	DockerUpdateConfByFile(req model.DaemonJsonUpdateRaw) error
	DockerUpdateLogOption(req model.LogOption) error
	DockerUpdateIpv6Option(req model.Ipv6Option) error
	DockerOperation(req model.DockerOperation) error

	Inspect(req model.Inspect) (*model.InspectResult, error)
	Prune(req model.Prune) (*model.PruneResult, error)

	ContainerQuery(req model.QueryContainer) (*model.PageResult, error)
	ContainerNames() (*model.PageResult, error)
	ContainerCreate(req model.ContainerOperate) error
	ContainerUpdate(req model.ContainerOperate) error
	ContainerUpgrade(req model.ContainerUpgrade) error
	ContainerInfo(containerID string) (*model.ContainerOperate, error)
	ContainerResourceUsage() (*model.PageResult, error)
	ContainerResourceLimit() (*model.ContainerResourceLimit, error)
	ContainerStats(id string) (*model.ContainerStats, error)
	ContainerRename(req model.Rename) error
	ContainerLogClean(containerID string) error
	ContainerOperation(req model.ContainerOperation) error
	ContainerLogs(req model.FileContentPartReq) (*model.FileContentPartRsp, error)

	ComposePage(req model.QueryCompose) (*model.PageResult, error)
	ComposeTest(req model.ComposeCreate) (*model.ComposeTestResult, error)
	ComposeCreate(req model.ComposeCreate) (*model.ComposeCreateResult, error)
	ComposeRemove(req model.ComposeRemove) error
	ComposeOperation(req model.ComposeOperation) error
	ComposeDetail(req model.ComposeDetailReq) (*model.ComposeDetailRsp, error)
	ComposeUpdate(req model.ComposeUpdate) error

	ImagePage(req model.SearchPageInfo) (*model.PageResult, error)
	ImageList() (*model.PageResult, error)
	ImageBuild(req model.ImageBuild) (*model.ImageOperationResult, error)
	ImagePull(req model.ImagePull) (*model.ImageOperationResult, error)
	ImageLoad(req model.ImageLoad) (*model.ImageOperationResult, error)
	ImageSave(req model.ImageSave) error
	ImagePush(req model.ImagePush) (*model.ImageOperationResult, error)
	ImageRemove(req model.BatchDelete) error
	ImageTag(req model.ImageTag) error

	VolumePage(req model.SearchPageInfo) (*model.PageResult, error)
	VolumeList() (*model.PageResult, error)
	VolumeDelete(req model.BatchDelete) error
	VolumeCreate(req model.VolumeCreate) error

	NetworkPage(req model.SearchPageInfo) (*model.PageResult, error)
	NetworkList() (*model.PageResult, error)
	NetworkDelete(req model.BatchDelete) error
	NetworkCreate(req model.NetworkCreate) error
}

func NewIDockerService() IDockerService {
	return &DockerService{}
}

func (s *DockerService) DockerStatus() (*model.DockerStatus, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.DockerStatus{Status: constant.Stopped}, err
	}
	defer client.Close()
	return client.LoadDockerStatus()
}

func (s *DockerService) DockerConf() (*model.DaemonJsonConf, error) {
	client, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.LoadDockerConf()
}

func (s *DockerService) DockerUpdateConf(req model.KeyValue) error {
	global.LOG.Info("upd conf start")
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			global.LOG.Error("mkdir %v err: %v", constant.DaemonJsonPath, err)
			return err
		}
		if err = os.WriteFile(constant.DaemonJsonPath, []byte("{}"), 0640); err != nil {
			global.LOG.Error("create %v err: %v", constant.DaemonJsonPath, err)
			return err
		}
	}

	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		global.LOG.Error("read %v err: %v", constant.DaemonJsonPath, err)
		return err
	}
	daemonMap := make(map[string]interface{})
	if err := json.Unmarshal(file, &daemonMap); err != nil {
		global.LOG.Error("unmarshal %v err: %v", constant.DaemonJsonPath, err)
		return fmt.Errorf("invalid JSON in %s; please check and fix the file: %w", constant.DaemonJsonPath, err)
	}

	switch req.Key {
	case "Registries":
		req.Value = strings.TrimSuffix(req.Value, ",")
		if len(req.Value) == 0 {
			delete(daemonMap, "insecure-registries")
		} else {
			daemonMap["insecure-registries"] = strings.Split(req.Value, ",")
		}
	case "Mirrors":
		req.Value = strings.TrimSuffix(req.Value, ",")
		if len(req.Value) == 0 {
			delete(daemonMap, "registry-mirrors")
		} else {
			daemonMap["registry-mirrors"] = strings.Split(req.Value, ",")
		}
	case "Ipv6":
		if req.Value == "disable" {
			delete(daemonMap, "ipv6")
			delete(daemonMap, "fixed-cidr-v6")
			delete(daemonMap, "ip6tables")
			delete(daemonMap, "experimental")
		}
	case "LogOption":
		if req.Value == "disable" {
			delete(daemonMap, "log-opts")
		}
	case "LiveRestore":
		if req.Value == "disable" {
			delete(daemonMap, "live-restore")
		} else {
			daemonMap["live-restore"] = true
		}
	case "IPtables":
		if req.Value == "enable" {
			delete(daemonMap, "iptables")
		} else {
			daemonMap["iptables"] = false
		}
	case "Driver":
		if opts, ok := daemonMap["exec-opts"]; ok {
			if optsValue, isArray := opts.([]interface{}); isArray {
				for i := 0; i < len(optsValue); i++ {
					if opt, isStr := optsValue[i].(string); isStr {
						if strings.HasPrefix(opt, "native.cgroupdriver=") {
							optsValue[i] = "native.cgroupdriver=" + req.Value
							break
						}
					}
				}
			}
		} else {
			if req.Value == "systemd" {
				daemonMap["exec-opts"] = []string{"native.cgroupdriver=systemd"}
			}
		}
	}
	if len(daemonMap) == 0 {
		global.LOG.Info("no conf fields, won't update conf")
		return nil
	}
	global.LOG.Info("upd conf")
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	global.LOG.Info("new conf %v", string(newJson))
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}
	global.LOG.Info("restart docker")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.LOG.Error("panic in docker restart goroutine: %v", r)
			}
		}()
		stdout, err := shell.ExecuteCommand("systemctl restart docker")
		if err != nil {
			global.LOG.Error("restart docker: %v", stdout)
		}
	}()
	return nil
}

func (s *DockerService) DockerUpdateConfByFile(req model.DaemonJsonUpdateRaw) error {
	global.LOG.Info("upd conf raw start")

	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			return err
		}
		if err = os.WriteFile(constant.DaemonJsonPath, []byte("{}"), 0640); err != nil {
			global.LOG.Error("create %v err: %v", constant.DaemonJsonPath, err)
			return err
		}
	}

	if len(req.Content) == 0 {
		global.LOG.Info("no content, won't update conf")
		return fmt.Errorf("invalid content")
	}

	daemonMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(strings.TrimSpace(req.Content)), &daemonMap); err != nil {
		global.LOG.Error("failed to unmarshal conf raw content %v", err)
		return err
	}

	if len(daemonMap) == 0 {
		global.LOG.Info("no conf fields, won't update conf")
		return nil
	}

	global.LOG.Info("upd conf raw")
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	global.LOG.Info("new conf raw %v", string(newJson))
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}
	global.LOG.Info("restart docker")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.LOG.Error("panic in docker restart goroutine: %v", r)
			}
		}()
		stdout, err := shell.ExecuteCommand("systemctl restart docker")
		if err != nil {
			global.LOG.Error("restart docker: %v", stdout)
		}
	}()
	return nil
}

func (s *DockerService) DockerUpdateLogOption(req model.LogOption) error {
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			return err
		}
		_, _ = os.Create(constant.DaemonJsonPath)
	}

	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	daemonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &daemonMap)

	s.changeLogOption(daemonMap, req.LogMaxFile, req.LogMaxSize)
	if len(daemonMap) == 0 {
		_ = os.Remove(constant.DaemonJsonPath)
		return nil
	}
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}

	go func() {
		stdout, err := shell.ExecuteCommand("systemctl restart docker")
		if err != nil {
			global.LOG.Error(string(stdout))
		}
	}()
	return nil
}

func (s *DockerService) DockerUpdateIpv6Option(req model.Ipv6Option) error {
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			return err
		}
		_, _ = os.Create(constant.DaemonJsonPath)
	}

	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	daemonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &daemonMap)

	daemonMap["ipv6"] = true
	daemonMap["fixed-cidr-v6"] = req.FixedCidrV6
	if req.Ip6Tables {
		daemonMap["ip6tables"] = req.Ip6Tables
	}
	if req.Experimental {
		daemonMap["experimental"] = req.Experimental
	}
	if len(daemonMap) == 0 {
		_ = os.Remove(constant.DaemonJsonPath)
		return nil
	}
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}

	go func() {
		stdout, err := shell.ExecuteCommand("systemctl restart docker")
		if err != nil {
			global.LOG.Error(string(stdout))
		}
	}()
	return nil
}

func (s *DockerService) DockerOperation(req model.DockerOperation) error {
	service := "docker"
	if req.Operation == "stop" {
		service = "docker.socket"
	}
	stdout, err := shell.ExecuteCommand(fmt.Sprintf("systemctl %s %s ", req.Operation, service))
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func (s *DockerService) Inspect(req model.Inspect) (*model.InspectResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.InspectResult{Type: req.Type, ID: req.ID}, err
	}
	defer client.Close()
	return client.Inspect(req)
}

func (s *DockerService) Prune(req model.Prune) (*model.PruneResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PruneResult{}, err
	}
	defer client.Close()
	return client.Prune(req)
}

func (s *DockerService) changeLogOption(daemonMap map[string]interface{}, logMaxFile, logMaxSize string) {
	if opts, ok := daemonMap["log-opts"]; ok {
		if len(logMaxFile) != 0 || len(logMaxSize) != 0 {
			daemonMap["log-driver"] = "json-file"
		}
		optsMap, isMap := opts.(map[string]interface{})
		if isMap {
			if len(logMaxFile) != 0 {
				optsMap["max-file"] = logMaxFile
			} else {
				delete(optsMap, "max-file")
			}
			if len(logMaxSize) != 0 {
				optsMap["max-size"] = logMaxSize
			} else {
				delete(optsMap, "max-size")
			}
			if len(optsMap) == 0 {
				delete(daemonMap, "log-opts")
			}
		} else {
			optsMap := make(map[string]interface{})
			if len(logMaxFile) != 0 {
				optsMap["max-file"] = logMaxFile
			}
			if len(logMaxSize) != 0 {
				optsMap["max-size"] = logMaxSize
			}
			if len(optsMap) != 0 {
				daemonMap["log-opts"] = optsMap
			}
		}
	} else {
		if len(logMaxFile) != 0 || len(logMaxSize) != 0 {
			daemonMap["log-driver"] = "json-file"
		}
		optsMap := make(map[string]interface{})
		if len(logMaxFile) != 0 {
			optsMap["max-file"] = logMaxFile
		}
		if len(logMaxSize) != 0 {
			optsMap["max-size"] = logMaxSize
		}
		if len(optsMap) != 0 {
			daemonMap["log-opts"] = optsMap
		}
	}
}
