package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils/common"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type DockerClient struct {
	cli *client.Client
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

func NewClient() (DockerClient, error) {
	return NewClientForHost("unix:///var/run/docker.sock")
}

func NewClientForHost(host string) (DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(host), client.WithAPIVersionNegotiation())
	if err != nil {
		return DockerClient{}, err
	}

	return DockerClient{
		cli: cli,
	}, nil
}

func (c DockerClient) Close() {
	_ = c.cli.Close()
}

func (c DockerClient) Inspect(req model.Inspect) (*model.InspectResult, error) {
	result := model.InspectResult{
		Type: req.Type,
		ID:   req.ID,
	}

	var inspectInfo interface{}
	var err error
	switch req.Type {
	case "container":
		inspectInfo, err = c.cli.ContainerInspect(context.Background(), req.ID)
	case "image":
		inspectInfo, _, err = c.cli.ImageInspectWithRaw(context.Background(), req.ID)
	case "network":
		inspectInfo, err = c.cli.NetworkInspect(context.TODO(), req.ID, network.InspectOptions{})
	case "volume":
		inspectInfo, err = c.cli.VolumeInspect(context.TODO(), req.ID)
	}
	if err != nil {
		return &result, err
	}

	bytes, err := json.Marshal(inspectInfo)
	if err != nil {
		return &result, err
	}
	result.Info = string(bytes)

	return &result, nil
}

func (c DockerClient) Prune(req model.Prune) (*model.PruneResult, error) {
	report := model.PruneResult{}
	pruneFilters := filters.NewArgs()
	if req.WithTagAll {
		pruneFilters.Add("dangling", "false")
		if req.PruneType != "image" {
			pruneFilters.Add("until", "24h")
		}
	}
	switch req.PruneType {
	case "container":
		rep, err := c.cli.ContainersPrune(context.Background(), pruneFilters)
		if err != nil {
			return &report, err
		}
		report.DeletedNumber = len(rep.ContainersDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	case "image":
		rep, err := c.cli.ImagesPrune(context.Background(), pruneFilters)
		if err != nil {
			return &report, err
		}
		report.DeletedNumber = len(rep.ImagesDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	case "network":
		rep, err := c.cli.NetworksPrune(context.Background(), pruneFilters)
		if err != nil {
			return &report, err
		}
		report.DeletedNumber = len(rep.NetworksDeleted)
	case "volume":
		versions, err := c.cli.ServerVersion(context.Background())
		if err != nil {
			return &report, err
		}
		if common.ComparePanelVersion(versions.APIVersion, "1.42") {
			pruneFilters.Add("all", "true")
		}
		rep, err := c.cli.VolumesPrune(context.Background(), pruneFilters)
		if err != nil {
			return &report, err
		}
		report.DeletedNumber = len(rep.VolumesDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	case "buildcache":
		rep, err := c.cli.BuildCachePrune(context.Background(), types.BuildCachePruneOptions{})
		if err != nil {
			return &report, err
		}
		report.DeletedNumber = len(rep.CachesDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	}
	return &report, nil
}

func (c DockerClient) DeleteImage(imageID string) error {
	if _, err := c.cli.ImageRemove(context.Background(), imageID, image.RemoveOptions{Force: true}); err != nil {
		return err
	}
	return nil
}

func (c DockerClient) PullImage(ctx context.Context, imageName string) error {
	// TODO: 增加authConfig
	options := image.PullOptions{}
	// authConfig := registry.AuthConfig{
	// 	Username: "username", // TODO: 推送的用户名,
	// 	Password: "password", // TODO: 推送的密码
	// }
	// encodedJSON, err := json.Marshal(authConfig)
	// if err != nil {
	// 	return err
	// }
	// authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	// options.RegistryAuth = authStr
	if _, err := c.cli.ImagePull(ctx, imageName, options); err != nil {
		return err
	}
	return nil
}

func (c DockerClient) GetImageIDByName(imageName string) (string, error) {
	filter := filters.NewArgs()
	filter.Add("reference", imageName)
	list, err := c.cli.ImageList(context.Background(), image.ListOptions{
		Filters: filter,
	})
	if err != nil {
		return "", err
	}
	if len(list) > 0 {
		return list[0].ID, nil
	}
	return "", nil
}

func (c DockerClient) CheckImageExist(imageName string, useExactMatch bool) (bool, error) {
	filter := filters.NewArgs()
	if useExactMatch {
		// 精确匹配 imageName
		filter.Add("reference", imageName)
	} else {
		// 通配匹配 imageName 及其所有标签
		filter.Add("reference", imageName)
		filter.Add("reference", imageName+":*")
	}

	list, err := c.cli.ImageList(context.Background(), image.ListOptions{
		Filters: filter,
	})
	if err != nil {
		return false, err
	}
	return len(list) > 0, nil
}

func (c DockerClient) NetworkExist(name string) bool {
	var options types.NetworkListOptions
	options.Filters = filters.NewArgs(filters.Arg("name", name))
	networks, err := c.cli.NetworkList(context.Background(), options)
	if err != nil {
		return false
	}
	return len(networks) > 0
}

func (c DockerClient) CreateNetwork(name string) error {
	_, err := c.cli.NetworkCreate(context.Background(), name, types.NetworkCreate{
		Driver: "bridge",
	})
	return err
}

func (c DockerClient) reCreateAfterUpdate(ctx context.Context, name string, config *container.Config, hostConf *container.HostConfig, networkConf *types.NetworkSettings, logger *log.Log) {
	var oldNetworkConf network.NetworkingConfig
	if networkConf != nil {
		for networkKey := range networkConf.Networks {
			oldNetworkConf.EndpointsConfig = map[string]*network.EndpointSettings{networkKey: {}}
			break
		}
	}

	oldContainer, err := c.cli.ContainerCreate(ctx, config, hostConf, &oldNetworkConf, &v1.Platform{}, name)
	if err != nil {
		logger.Error("recreate after container update failed, err: %v", err)
		return
	}
	if err := c.cli.ContainerStart(ctx, oldContainer.ID, container.StartOptions{}); err != nil {
		logger.Error("restart after container update failed, err: %v", err)
	}
	logger.Error("recreate after container update successful")
}

func (c DockerClient) calculateBlockIO(blkio types.BlkioStats) (blkRead float64, blkWrite float64) {
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = (blkRead + float64(bioEntry.Value)) / 1024 / 1024
		case "write":
			blkWrite = (blkWrite + float64(bioEntry.Value)) / 1024 / 1024
		}
	}
	return
}

func (c DockerClient) calculateMemPercentUnix(memStats types.MemoryStats) float64 {
	memPercent := 0.0
	memUsage := float64(memStats.Usage)
	memLimit := float64(memStats.Limit)
	if memUsage > 0.0 && memLimit > 0.0 {
		memPercent = (memUsage / memLimit) * 100.0
	}
	return memPercent
}

func (c DockerClient) calculateCPUPercentUnix(stats *types.StatsJSON) float64 {
	cpuPercent := 0.0
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * 100.0
		if len(stats.CPUStats.CPUUsage.PercpuUsage) != 0 {
			cpuPercent = cpuPercent * float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		}
	}
	return cpuPercent
}

func (c DockerClient) calculateNetwork(network map[string]types.NetworkStats) (float64, float64) {
	var rx, tx float64

	for _, v := range network {
		rx += float64(v.RxBytes) / 1024
		tx += float64(v.TxBytes) / 1024
	}
	return rx, tx
}

func (c DockerClient) checkPortStats(ports []model.PortHelper) (nat.PortMap, error) {
	portMap := make(nat.PortMap)
	if len(ports) == 0 {
		return portMap, nil
	}
	for _, port := range ports {
		if strings.Contains(port.ContainerPort, "-") {
			if !strings.Contains(port.HostPort, "-") {
				return portMap, errors.New(constant.ErrPortRules)
			}
			hostStart, _ := strconv.Atoi(strings.Split(port.HostPort, "-")[0])
			hostEnd, _ := strconv.Atoi(strings.Split(port.HostPort, "-")[1])
			containerStart, _ := strconv.Atoi(strings.Split(port.ContainerPort, "-")[0])
			containerEnd, _ := strconv.Atoi(strings.Split(port.ContainerPort, "-")[1])
			if (hostEnd-hostStart) <= 0 || (containerEnd-containerStart) <= 0 {
				return portMap, errors.New(constant.ErrPortRules)
			}
			if (containerEnd - containerStart) != (hostEnd - hostStart) {
				return portMap, errors.New(constant.ErrPortRules)
			}
			for i := 0; i <= hostEnd-hostStart; i++ {
				bindItem := nat.PortBinding{HostPort: strconv.Itoa(hostStart + i), HostIP: port.HostIP}
				portMap[nat.Port(fmt.Sprintf("%d/%s", containerStart+i, port.Protocol))] = []nat.PortBinding{bindItem}
			}
			for i := hostStart; i <= hostEnd; i++ {
				if common.ScanPort(i) {
					return portMap, errors.New(constant.ErrPortInUsed)
				}
			}
		} else {
			portItem := 0
			if strings.Contains(port.HostPort, "-") {
				portItem, _ = strconv.Atoi(strings.Split(port.HostPort, "-")[0])
			} else {
				portItem, _ = strconv.Atoi(port.HostPort)
			}
			if common.ScanPort(portItem) {
				return portMap, errors.New(constant.ErrPortInUsed)
			}
			bindItem := nat.PortBinding{HostPort: strconv.Itoa(portItem), HostIP: port.HostIP}
			portMap[nat.Port(fmt.Sprintf("%s/%s", port.ContainerPort, port.Protocol))] = []nat.PortBinding{bindItem}
		}
	}
	return portMap, nil
}

func (c DockerClient) loadVolumeBinds(binds []types.MountPoint) []model.VolumeHelper {
	var datas []model.VolumeHelper
	for _, bind := range binds {
		var volumeItem model.VolumeHelper
		volumeItem.Type = string(bind.Type)
		if bind.Type == "volume" {
			volumeItem.SourceDir = bind.Name
		} else {
			volumeItem.SourceDir = bind.Source
		}
		volumeItem.ContainerDir = bind.Destination
		volumeItem.Mode = "ro"
		if bind.RW {
			volumeItem.Mode = "rw"
		}
		datas = append(datas, volumeItem)
	}
	return datas
}

func (c DockerClient) loadContainerPort(ports []types.Port) []string {
	var (
		ipv4Ports []types.Port
		ipv6Ports []types.Port
	)
	for _, port := range ports {
		if strings.Contains(port.IP, ":") {
			ipv6Ports = append(ipv6Ports, port)
		} else {
			ipv4Ports = append(ipv4Ports, port)
		}
	}
	list1 := c.simplifyPort(ipv4Ports)
	list2 := c.simplifyPort(ipv6Ports)
	return append(list1, list2...)
}

func (c DockerClient) simplifyPort(ports []types.Port) []string {
	var datas []string
	if len(ports) == 0 {
		return datas
	}
	if len(ports) == 1 {
		ip := ""
		if len(ports[0].IP) != 0 {
			ip = ports[0].IP + ":"
		}
		itemPortStr := fmt.Sprintf("%s%v/%s", ip, ports[0].PrivatePort, ports[0].Type)
		if ports[0].PublicPort != 0 {
			itemPortStr = fmt.Sprintf("%s%v->%v/%s", ip, ports[0].PublicPort, ports[0].PrivatePort, ports[0].Type)
		}
		datas = append(datas, itemPortStr)
		return datas
	}

	sort.Slice(ports, func(i, j int) bool {
		return ports[i].PrivatePort < ports[j].PrivatePort
	})
	start := ports[0]

	for i := 1; i < len(ports); i++ {
		if ports[i].PrivatePort != ports[i-1].PrivatePort+1 || ports[i].IP != ports[i-1].IP || ports[i].PublicPort != ports[i-1].PublicPort+1 || ports[i].Type != ports[i-1].Type {
			if ports[i-1].PrivatePort == start.PrivatePort {
				itemPortStr := fmt.Sprintf("%s:%v/%s", start.IP, start.PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v->%v/%s", start.IP, start.PublicPort, start.PrivatePort, start.Type)
				}
				if len(start.IP) == 0 {
					itemPortStr = strings.TrimPrefix(itemPortStr, ":")
				}
				datas = append(datas, itemPortStr)
			} else {
				itemPortStr := fmt.Sprintf("%s:%v-%v/%s", start.IP, start.PrivatePort, ports[i-1].PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v-%v->%v-%v/%s", start.IP, start.PublicPort, ports[i-1].PublicPort, start.PrivatePort, ports[i-1].PrivatePort, start.Type)
				}
				if len(start.IP) == 0 {
					itemPortStr = strings.TrimPrefix(itemPortStr, ":")
				}
				datas = append(datas, itemPortStr)
			}
			start = ports[i]
		}
		if i == len(ports)-1 {
			if ports[i].PrivatePort == start.PrivatePort {
				itemPortStr := fmt.Sprintf("%s:%v/%s", start.IP, start.PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v->%v/%s", start.IP, start.PublicPort, start.PrivatePort, start.Type)
				}
				if len(start.IP) == 0 {
					itemPortStr = strings.TrimPrefix(itemPortStr, ":")
				}
				datas = append(datas, itemPortStr)
			} else {
				itemPortStr := fmt.Sprintf("%s:%v-%v/%s", start.IP, start.PrivatePort, ports[i].PrivatePort, start.Type)
				if start.PublicPort != 0 {
					itemPortStr = fmt.Sprintf("%s:%v-%v->%v-%v/%s", start.IP, start.PublicPort, ports[i].PublicPort, start.PrivatePort, ports[i].PrivatePort, start.Type)
				}
				if len(start.IP) == 0 {
					itemPortStr = strings.TrimPrefix(itemPortStr, ":")
				}
				datas = append(datas, itemPortStr)
			}
		}
	}
	return datas
}

func (c DockerClient) loadCpuAndMem(containerID string) model.ContainerResourceUsage {
	data := model.ContainerResourceUsage{
		ContainerID: containerID,
	}
	res, err := c.cli.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		return data
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return data
	}
	var stats *container.StatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return data
	}

	data.CPUTotalUsage = stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage
	data.SystemUsage = stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage
	data.CPUPercent = c.calculateCPUPercentUnix(stats)
	data.PercpuUsage = len(stats.CPUStats.CPUUsage.PercpuUsage)

	data.MemoryCache = stats.MemoryStats.Stats["cache"]
	data.MemoryUsage = stats.MemoryStats.Usage
	data.MemoryLimit = stats.MemoryStats.Limit

	data.MemoryPercent = c.calculateMemPercentUnix(stats.MemoryStats)
	return data
}

func (c DockerClient) loadConfigInfo(isCreate bool, req model.ContainerOperate, oldContainer *types.ContainerJSON) (*container.Config, *container.HostConfig, *network.NetworkingConfig, error) {
	var config container.Config
	var hostConf container.HostConfig
	if !isCreate {
		config = *oldContainer.Config
		hostConf = *oldContainer.HostConfig
	}
	var networkConf network.NetworkingConfig

	portMap, err := c.checkPortStats(req.ExposedPorts)
	if err != nil {
		return nil, nil, nil, err
	}
	exposed := make(nat.PortSet)
	for port := range portMap {
		exposed[port] = struct{}{}
	}
	config.Image = req.Image
	config.Cmd = req.Cmd
	config.Entrypoint = req.Entrypoint
	config.Env = req.Env
	config.Labels = req.Labels
	config.ExposedPorts = exposed
	config.OpenStdin = req.OpenStdin
	config.Tty = req.Tty

	if len(req.Network) != 0 {
		switch req.Network {
		case "host", "none", "bridge":
			hostConf.NetworkMode = container.NetworkMode(req.Network)
		}
		if req.Ipv4 != "" || req.Ipv6 != "" {
			networkConf.EndpointsConfig = map[string]*network.EndpointSettings{
				req.Network: {
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: req.Ipv4,
						IPv6Address: req.Ipv6,
					},
				}}
		} else {
			networkConf.EndpointsConfig = map[string]*network.EndpointSettings{req.Network: {}}
		}
	} else {
		if req.Ipv4 != "" || req.Ipv6 != "" {
			return nil, nil, nil, fmt.Errorf("please set up the network")
		}
		networkConf = network.NetworkingConfig{}
	}

	hostConf.Privileged = req.Privileged
	hostConf.AutoRemove = req.AutoRemove
	hostConf.CPUShares = req.CPUShares
	hostConf.PublishAllPorts = req.PublishAllPorts
	hostConf.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(req.RestartPolicy)}
	if req.RestartPolicy == "on-failure" {
		hostConf.RestartPolicy.MaximumRetryCount = 5
	}
	hostConf.NanoCPUs = int64(req.NanoCPUs * 1000000000)
	hostConf.Memory = int64(req.Memory * 1024 * 1024)
	hostConf.MemorySwap = 0
	hostConf.PortBindings = portMap
	hostConf.Binds = []string{}
	hostConf.Mounts = []mount.Mount{}
	config.Volumes = make(map[string]struct{})
	for _, volume := range req.Volumes {
		if volume.Type == "volume" {
			hostConf.Mounts = append(hostConf.Mounts, mount.Mount{
				Type:   mount.Type(volume.Type),
				Source: volume.SourceDir,
				Target: volume.ContainerDir,
			})
			config.Volumes[volume.ContainerDir] = struct{}{}
		} else {
			hostConf.Binds = append(hostConf.Binds, fmt.Sprintf("%s:%s:%s", volume.SourceDir, volume.ContainerDir, volume.Mode))
		}
	}
	return &config, &hostConf, &networkConf, nil
}

func (c DockerClient) formatFileSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}

func (c DockerClient) checkUsed(imageID string, containers []types.Container) bool {
	for _, container := range containers {
		if container.ImageID == imageID {
			return true
		}
	}
	return false
}
