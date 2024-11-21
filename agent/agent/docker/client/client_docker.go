package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
)

func (c DockerClient) LoadDockerStatus() string {
	if _, err := c.cli.Ping(context.Background()); err != nil {
		return constant.Stopped
	}

	return constant.StatusRunning
}

func (c DockerClient) LoadDockerConf() *model.DaemonJsonConf {
	ctx := context.Background()
	var data model.DaemonJsonConf
	data.IPTables = true
	data.Status = constant.StatusRunning
	data.Version = "-"
	if _, err := c.cli.Ping(ctx); err != nil {
		data.Status = constant.Stopped
	}
	itemVersion, err := c.cli.ServerVersion(ctx)
	if err == nil {
		data.Version = itemVersion.Version
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
