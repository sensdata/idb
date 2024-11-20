package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
)

type DockerService struct{}

type IDockerService interface {
	UpdateConf(req model.KeyValue) error
	UpdateLogOption(req model.LogOption) error
	UpdateIpv6Option(req model.Ipv6Option) error
	UpdateConfByFile(info model.DaemonJsonUpdateByFile) error
	LoadDockerStatus() string
	LoadDockerConf() *model.DaemonJsonConf
	OperateDocker(req model.DockerOperation) error
}

func NewIDockerService() IDockerService {
	return &DockerService{}
}

type daemonJsonItem struct {
	Status       string    `json:"status"`
	Mirrors      []string  `json:"registry-mirrors"`
	Registries   []string  `json:"insecure-registries"`
	LiveRestore  bool      `json:"live-restore"`
	Ipv6         bool      `json:"ipv6"`
	FixedCidrV6  string    `json:"fixed-cidr-v6"`
	Ip6Tables    bool      `json:"ip6tables"`
	Experimental bool      `json:"experimental"`
	IPTables     bool      `json:"iptables"`
	ExecOpts     []string  `json:"exec-opts"`
	LogOption    logOption `json:"log-opts"`
}

type logOption struct {
	LogMaxSize string `json:"max-size"`
	LogMaxFile string `json:"max-file"`
}

func (u *DockerService) LoadDockerStatus() string {
	client, err := NewDockerClient()
	if err != nil {
		return constant.Stopped
	}
	defer client.Close()
	if _, err := client.Ping(context.Background()); err != nil {
		return constant.Stopped
	}

	return constant.StatusRunning
}

func (u *DockerService) LoadDockerConf() *model.DaemonJsonConf {
	ctx := context.Background()
	var data model.DaemonJsonConf
	data.IPTables = true
	data.Status = constant.StatusRunning
	data.Version = "-"
	client, err := NewDockerClient()
	if err != nil {
		data.Status = constant.Stopped
	} else {
		defer client.Close()
		if _, err := client.Ping(ctx); err != nil {
			data.Status = constant.Stopped
		}
		itemVersion, err := client.ServerVersion(ctx)
		if err == nil {
			data.Version = itemVersion.Version
		}
	}
	data.IsSwarm = false
	stdout2, _ := shell.ExecuteCommand("docker info  | grep Swarm")
	if string(stdout2) == " Swarm: active\n" {
		data.IsSwarm = true
	}
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil {
		return &data
	}
	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return &data
	}
	var conf daemonJsonItem
	daemonMap := make(map[string]interface{})
	if err := json.Unmarshal(file, &daemonMap); err != nil {
		return &data
	}
	arr, err := json.Marshal(daemonMap)
	if err != nil {
		return &data
	}
	if err := json.Unmarshal(arr, &conf); err != nil {
		fmt.Println(err)
		return &data
	}
	if _, ok := daemonMap["iptables"]; !ok {
		conf.IPTables = true
	}
	data.CgroupDriver = "cgroupfs"
	for _, opt := range conf.ExecOpts {
		if strings.HasPrefix(opt, "native.cgroupdriver=") {
			data.CgroupDriver = strings.ReplaceAll(opt, "native.cgroupdriver=", "")
			break
		}
	}
	data.Ipv6 = conf.Ipv6
	data.FixedCidrV6 = conf.FixedCidrV6
	data.Ip6Tables = conf.Ip6Tables
	data.Experimental = conf.Experimental
	data.LogMaxSize = conf.LogOption.LogMaxSize
	data.LogMaxFile = conf.LogOption.LogMaxFile
	data.Mirrors = conf.Mirrors
	data.Registries = conf.Registries
	data.IPTables = conf.IPTables
	data.LiveRestore = conf.LiveRestore
	return &data
}

func (u *DockerService) UpdateConf(req model.KeyValue) error {
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

func (u *DockerService) UpdateLogOption(req model.LogOption) error {
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

	changeLogOption(daemonMap, req.LogMaxFile, req.LogMaxSize)
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

func (u *DockerService) UpdateIpv6Option(req model.Ipv6Option) error {
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

func (u *DockerService) UpdateConfByFile(req model.DaemonJsonUpdateByFile) error {
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

func (u *DockerService) OperateDocker(req model.DockerOperation) error {
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

func changeLogOption(daemonMap map[string]interface{}, logMaxFile, logMaxSize string) {
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
