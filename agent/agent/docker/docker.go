package docker

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sensdata/idb/agent/agent/docker/client"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
)

type DockerService struct{}

type IDockerService interface {
	DockerStatus() string
	DockerConf() *model.DaemonJsonConf
	DockerUpdateConf(req model.KeyValue) error
	DockerUpdateConfByFile(info model.DaemonJsonUpdateByFile) error
	DockerUpdateLogOption(req model.LogOption) error
	DockerUpdateIpv6Option(req model.Ipv6Option) error
	DockerOperation(req model.DockerOperation) error

	Inspect(req model.Inspect) (string, error)
	Prune(req model.Prune) (*model.PruneResult, error)

	ContainerQuery(req model.QueryContainer) (*model.PageResult, error)
	ContainerNames() ([]string, error)
	ContainerCreate(req model.ContainerOperate) error
	ContainerUpdate(req model.ContainerOperate) error
	ContainerUpgrade(req model.ContainerUpgrade) error
	ContainerInfo(containerID string) (*model.ContainerOperate, error)
	ContainerResourceUsage() ([]model.ContainerResourceUsage, error)
	ContainerResourceLimit() (*model.ContainerResourceLimit, error)
	ContainerStats(id string) (*model.ContainerStats, error)
	ContainerRename(req model.Rename) error
	ContainerLogClean(containerID string) error
	ContainerOperation(req model.ContainerOperation) error
	ContainerLogs(wsConn *websocket.Conn, containerType, container, since, tail string, follow bool) error

	// ComposePage(req model.SearchPageInfo) (int64, interface{}, error)
	// ComposeCreate(req model.ComposeCreate) (string, error)
	// ComposeOperation(req model.ComposeOperation) error
	// ComposeTest(req model.ComposeCreate) (bool, error)
	// ComposeUpdate(req model.ComposeUpdate) error

	ImagePage(req model.SearchPageInfo) (*model.PageResult, error)
	ImageList() ([]model.Options, error)
	ImageBuild(req model.ImageBuild) (string, error)
	ImagePull(req model.ImagePull) (string, error)
	ImageLoad(req model.ImageLoad) error
	ImageSave(req model.ImageSave) error
	ImagePush(req model.ImagePush) (string, error)
	ImageRemove(req model.BatchDelete) error
	ImageTag(req model.ImageTag) error

	VolumePage(req model.SearchPageInfo) (*model.PageResult, error)
	VolumeList() ([]model.Options, error)
	VolumeDelete(req model.BatchDelete) error
	VolumeCreate(req model.VolumeCreate) error

	NetworkPage(req model.SearchPageInfo) (*model.PageResult, error)
	NetworkList() ([]model.Options, error)
	NetworkDelete(req model.BatchDelete) error
	NetworkCreate(req model.NetworkCreate) error
}

func NewIDockerService() IDockerService {
	return &DockerService{}
}

func (s *DockerService) DockerStatus() string {
	client, err := client.NewClient()
	if err != nil {
		return constant.Stopped
	}
	defer client.Close()
	return client.LoadDockerStatus()
}

func (s *DockerService) DockerConf() *model.DaemonJsonConf {
	client, err := client.NewClient()
	if err != nil {
		return nil
	}
	defer client.Close()
	return client.LoadDockerConf()
}

func (s *DockerService) DockerUpdateConf(req model.KeyValue) error {
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

	stdout, err := shell.ExecuteCommand("systemctl restart docker")
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func (s *DockerService) DockerUpdateConfByFile(req model.DaemonJsonUpdateByFile) error {
	if len(req.File) == 0 {
		_ = os.Remove(constant.DaemonJsonPath)
		return nil
	}
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			return err
		}
		_, _ = os.Create(constant.DaemonJsonPath)
	}
	file, err := os.OpenFile(constant.DaemonJsonPath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()

	stdout, err := shell.ExecuteCommand("systemctl restart docker")
	if err != nil {
		return errors.New(string(stdout))
	}
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

	stdout, err := shell.ExecuteCommand("systemctl restart docker")
	if err != nil {
		return errors.New(string(stdout))
	}
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

	stdout, err := shell.ExecuteCommand("systemctl restart docker")
	if err != nil {
		return errors.New(string(stdout))
	}
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

func (s *DockerService) Inspect(req model.Inspect) (string, error) {
	client, err := client.NewClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	inspectInfo, err := client.Inspect(req)
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(inspectInfo)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
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
