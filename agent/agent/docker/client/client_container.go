package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
)

func (c DockerClient) ContainerQuery(req model.QueryContainer) (*model.PageResult, error) {
	var (
		result  model.PageResult
		records []types.Container
	)

	// 构建过滤器
	options := container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(), // 初始化过滤器
	}
	if len(req.Info) > 0 {
		options.Filters.Add("name", req.Info)
	}
	if req.State != "all" {
		options.Filters.Add("status", req.State)
	}

	// 获取容器列表
	containers, err := c.cli.ContainerList(context.Background(), options)
	if err != nil {
		return &result, err
	}

	// 排序
	switch req.OrderBy {
	case "name":
		sort.Slice(containers, func(i, j int) bool {
			if req.Order == constant.OrderAsc {
				return containers[i].Names[0][1:] < containers[j].Names[0][1:]
			}
			return containers[i].Names[0][1:] > containers[j].Names[0][1:]
		})
	case "state":
		sort.Slice(containers, func(i, j int) bool {
			if req.Order == constant.OrderAsc {
				return containers[i].State < containers[j].State
			}
			return containers[i].State > containers[j].State
		})
	default:
		sort.Slice(containers, func(i, j int) bool {
			if req.Order == constant.OrderAsc {
				return containers[i].Created < containers[j].Created
			}
			return containers[i].Created > containers[j].Created
		})
	}

	// 分页
	total := len(containers)
	start, end := (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]types.Container, 0)
	} else {
		if end > total {
			end = total
		}
		records = containers[start:end]
	}

	backDatas := make([]model.ContainerInfo, len(records))
	for i := 0; i < len(records); i++ {
		item := records[i]
		var from string
		if lable, ok := item.Labels[constant.IDBType]; ok {
			from = lable
		}
		ports := c.loadContainerPort(item.Ports)
		info := model.ContainerInfo{
			ContainerID: item.ID,
			CreateTime:  time.Unix(item.Created, 0).Format("2006-01-02 15:04:05"),
			Name:        item.Names[0][1:],
			ImageId:     strings.Split(item.ImageID, ":")[1],
			ImageName:   item.Image,
			State:       item.State,
			RunTime:     item.Status,
			Ports:       ports,
			From:        from,
		}
		// install, _ := appInstallRepo.GetFirst(appInstallRepo.WithContainerName(info.Name))
		// if install.ID > 0 {
		// 	info.AppInstallName = install.Name
		// 	info.AppName = install.App.Name
		// 	websites, _ := websiteRepo.GetBy(websiteRepo.WithAppInstallId(install.ID))
		// 	for _, website := range websites {
		// 		info.Websites = append(info.Websites, website.PrimaryDomain)
		// 	}
		// }
		if item.NetworkSettings != nil && len(item.NetworkSettings.Networks) > 0 {
			networks := make([]string, 0, len(item.NetworkSettings.Networks))
			for key := range item.NetworkSettings.Networks {
				networks = append(networks, item.NetworkSettings.Networks[key].IPAddress)
			}
			sort.Strings(networks)
			info.Network = networks
		}
		backDatas[i] = info
	}

	result.Total = int64(total)
	result.Items = backDatas
	return &result, nil
}

func (c DockerClient) ContainerNames() (*model.PageResult, error) {
	var result model.PageResult
	containers, err := c.cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return &result, err
	}

	var datas []string
	for _, container := range containers {
		for _, name := range container.Names {
			datas = append(datas, strings.TrimPrefix(name, "/"))
		}
	}

	sort.Strings(datas) // 按字母顺序排序

	result.Total = int64(len(datas))
	result.Items = datas
	return &result, nil
}

func (c DockerClient) ContainerCreate(req model.ContainerOperate, logger *log.Log) error {
	ctx := context.Background()
	newContainer, _ := c.cli.ContainerInspect(ctx, req.Name)
	if newContainer.ContainerJSONBase != nil {
		return errors.New(constant.ErrContainerName)
	}

	exist, err := c.CheckImageExist(req.Image, true)
	if err != nil {
		return err
	}
	if !exist || req.ForcePull {
		if err := c.PullImage(ctx, req.Image); err != nil {
			if !req.ForcePull {
				return err
			}
			logger.Error("force pull image %s failed, err: %v", req.Image, err)
		}
	}
	imageInfo, _, err := c.cli.ImageInspectWithRaw(ctx, req.Image)
	if err != nil {
		return err
	}
	if len(req.Entrypoint) == 0 {
		req.Entrypoint = imageInfo.Config.Entrypoint
	}
	if len(req.Cmd) == 0 {
		req.Cmd = imageInfo.Config.Cmd
	}
	config, hostConf, networkConf, err := c.loadConfigInfo(true, req, nil)
	if err != nil {
		return err
	}
	logger.Info("new container info %s has been made, now start to create", req.Name)
	con, err := c.cli.ContainerCreate(ctx, config, hostConf, networkConf, &v1.Platform{}, req.Name)
	if err != nil {
		_ = c.cli.ContainerRemove(ctx, req.Name, container.RemoveOptions{RemoveVolumes: true, Force: true})
		return err
	}
	logger.Info("create container %s successful! now check if the container is started and delete the container information if it is not.", req.Name)
	if err := c.cli.ContainerStart(ctx, con.ID, container.StartOptions{}); err != nil {
		_ = c.cli.ContainerRemove(ctx, req.Name, container.RemoveOptions{RemoveVolumes: true, Force: true})
		return fmt.Errorf("create successful but start failed, err: %v", err)
	}
	return nil
}

func (c DockerClient) ContainerUpdate(req model.ContainerOperate, logger *log.Log) error {
	ctx := context.Background()
	newContainer, _ := c.cli.ContainerInspect(ctx, req.Name)
	if newContainer.ContainerJSONBase != nil && newContainer.ID != req.ContainerID {
		return errors.New(constant.ErrContainerName)
	}

	oldContainer, err := c.cli.ContainerInspect(ctx, req.ContainerID)
	if err != nil {
		return err
	}
	exist, err := c.CheckImageExist(req.Image, true)
	if err != nil {
		return err
	}
	if !exist || req.ForcePull {
		if err := c.PullImage(ctx, req.Image); err != nil {
			if !req.ForcePull {
				return err
			}
			logger.Error("force pull image %s failed, err: %v", req.Image, err)
			return fmt.Errorf("pull image %s failed, err: %v", req.Image, err)
		}
	}

	if err := c.cli.ContainerRemove(ctx, req.ContainerID, container.RemoveOptions{Force: true}); err != nil {
		return err
	}

	config, hostConf, networkConf, err := c.loadConfigInfo(false, req, &oldContainer)
	if err != nil {
		c.reCreateAfterUpdate(ctx, req.Name, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings, logger)
		return err
	}

	logger.Info("new container info %s has been update, now start to recreate", req.Name)
	con, err := c.cli.ContainerCreate(ctx, config, hostConf, networkConf, &v1.Platform{}, req.Name)
	if err != nil {
		c.reCreateAfterUpdate(ctx, req.Name, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings, logger)
		return fmt.Errorf("update container failed, err: %v", err)
	}
	logger.Info("update container %s successful! now check if the container is started.", req.Name)
	if err := c.cli.ContainerStart(ctx, con.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("update successful but start failed, err: %v", err)
	}

	return nil
}

func (c DockerClient) ContainerUpgrade(req model.ContainerUpgrade, logger *log.Log) error {
	ctx := context.Background()
	oldContainer, err := c.cli.ContainerInspect(ctx, req.Name)
	if err != nil {
		return err
	}
	exist, err := c.CheckImageExist(req.Image, true)
	if err != nil {
		return err
	}
	if !exist || req.ForcePull {
		if err := c.PullImage(ctx, req.Image); err != nil {
			if !req.ForcePull {
				return err
			}
			logger.Error("force pull image %s failed, err: %v", req.Image, err)
			return fmt.Errorf("pull image %s failed, err: %v", req.Image, err)
		}
	}
	config := oldContainer.Config
	config.Image = req.Image
	hostConf := oldContainer.HostConfig
	var networkConf network.NetworkingConfig
	if oldContainer.NetworkSettings != nil {
		for networkKey := range oldContainer.NetworkSettings.Networks {
			networkConf.EndpointsConfig = map[string]*network.EndpointSettings{networkKey: {}}
			break
		}
	}
	if err := c.cli.ContainerRemove(ctx, req.Name, container.RemoveOptions{Force: true}); err != nil {
		return err
	}

	logger.Info("new container info %s has been update, now start to recreate", req.Name)
	con, err := c.cli.ContainerCreate(ctx, config, hostConf, &networkConf, &v1.Platform{}, req.Name)
	if err != nil {
		c.reCreateAfterUpdate(ctx, req.Name, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings, logger)
		return fmt.Errorf("upgrade container failed, err: %v", err)
	}
	logger.Info("upgrade container %s successful! now check if the container is started.", req.Name)
	if err := c.cli.ContainerStart(ctx, con.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("upgrade successful but start failed, err: %v", err)
	}

	return nil
}

func (c DockerClient) ContainerInfo(containerID string) (*model.ContainerOperate, error) {
	ctx := context.Background()
	oldContainer, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}

	var data model.ContainerOperate
	data.ContainerID = oldContainer.ID
	data.Name = strings.ReplaceAll(oldContainer.Name, "/", "")
	data.Image = oldContainer.Config.Image
	if oldContainer.NetworkSettings != nil {
		for network := range oldContainer.NetworkSettings.Networks {
			data.Network = network
			break
		}
	}

	networkSettings := oldContainer.NetworkSettings
	bridgeNetworkSettings := networkSettings.Networks[data.Network]
	if bridgeNetworkSettings.IPAMConfig != nil {
		ipv4Address := bridgeNetworkSettings.IPAMConfig.IPv4Address
		data.Ipv4 = ipv4Address
		ipv6Address := bridgeNetworkSettings.IPAMConfig.IPv6Address
		data.Ipv6 = ipv6Address
	} else {
		data.Ipv4 = bridgeNetworkSettings.IPAddress
	}

	data.Cmd = oldContainer.Config.Cmd
	data.OpenStdin = oldContainer.Config.OpenStdin
	data.Tty = oldContainer.Config.Tty
	data.Entrypoint = oldContainer.Config.Entrypoint
	data.Env = oldContainer.Config.Env
	data.CPUShares = oldContainer.HostConfig.CPUShares
	for key, val := range oldContainer.Config.Labels {
		data.Labels = append(data.Labels, model.KeyValue{Key: key, Value: val})
	}
	for key, val := range oldContainer.HostConfig.PortBindings {
		var itemPort model.PortHelper
		if !strings.Contains(string(key), "/") {
			continue
		}
		itemPort.ContainerPort = strings.Split(string(key), "/")[0]
		itemPort.Protocol = strings.Split(string(key), "/")[1]
		for _, binds := range val {
			itemPort.HostIP = binds.HostIP
			itemPort.HostPort = binds.HostPort
			data.ExposedPorts = append(data.ExposedPorts, itemPort)
		}
	}
	data.AutoRemove = oldContainer.HostConfig.AutoRemove
	data.Privileged = oldContainer.HostConfig.Privileged
	data.PublishAllPorts = oldContainer.HostConfig.PublishAllPorts
	data.RestartPolicy = string(oldContainer.HostConfig.RestartPolicy.Name)
	if oldContainer.HostConfig.NanoCPUs != 0 {
		data.NanoCPUs = float64(oldContainer.HostConfig.NanoCPUs) / 1000000000
	}
	if oldContainer.HostConfig.Memory != 0 {
		data.Memory = float64(oldContainer.HostConfig.Memory) / 1024 / 1024
	}
	data.Volumes = c.loadVolumeBinds(oldContainer.Mounts)

	return &data, nil
}

func (c DockerClient) ContainerResourceUsage() (*model.PageResult, error) {
	var result model.PageResult
	list, err := c.cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return &result, err
	}
	var datas []model.ContainerResourceUsage
	var wg sync.WaitGroup
	wg.Add(len(list))
	for i := 0; i < len(list); i++ {
		go func(item types.Container) {
			datas = append(datas, c.loadCpuAndMem(item.ID))
			wg.Done()
		}(list[i])
	}
	wg.Wait()

	result.Total = int64(len(datas))
	result.Items = datas
	return &result, nil
}

func (c DockerClient) ContainerStats(id string) (*model.ContainerStats, error) {
	res, err := c.cli.ContainerStats(context.TODO(), id, false)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()
		return nil, err
	}
	res.Body.Close()
	var stats *container.StatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, err
	}
	var data model.ContainerStats
	data.CPUPercent = c.calculateCPUPercentUnix(stats)
	data.IORead, data.IOWrite = c.calculateBlockIO(stats.BlkioStats)
	data.Memory = float64(stats.MemoryStats.Usage) / 1024 / 1024
	if cache, ok := stats.MemoryStats.Stats["cache"]; ok {
		data.Cache = float64(cache) / 1024 / 1024
	}
	data.NetworkRX, data.NetworkTX = c.calculateNetwork(stats.Networks)
	data.ShotTime = stats.Read
	return &data, nil
}

func (c DockerClient) ContainerRename(req model.Rename) error {
	ctx := context.Background()
	newContainer, _ := c.cli.ContainerInspect(ctx, req.NewName)
	if newContainer.ContainerJSONBase != nil {
		return errors.New(constant.ErrContainerName)
	}
	return c.cli.ContainerRename(ctx, req.Name, req.NewName)
}

func (c DockerClient) ContainerLogClean(containerID string) error {
	ctx := context.Background()
	containerItem, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}
	if err := c.cli.ContainerStop(ctx, containerItem.ID, container.StopOptions{}); err != nil {
		return err
	}
	file, err := os.OpenFile(containerItem.LogPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = file.Truncate(0); err != nil {
		return err
	}
	_, _ = file.Seek(0, 0)

	files, _ := filepath.Glob(fmt.Sprintf("%s.*", containerItem.LogPath))
	for _, file := range files {
		_ = os.Remove(file)
	}

	if err := c.cli.ContainerStart(ctx, containerItem.ID, container.StartOptions{}); err != nil {
		return err
	}
	return nil
}

func (c DockerClient) ContainerOperation(req model.ContainerOperation) error {
	var err error
	ctx := context.Background()
	for _, item := range req.Names {
		switch req.Operation {
		case constant.ContainerOpStart:
			err = c.cli.ContainerStart(ctx, item, container.StartOptions{})
		case constant.ContainerOpStop:
			err = c.cli.ContainerStop(ctx, item, container.StopOptions{})
		case constant.ContainerOpRestart:
			err = c.cli.ContainerRestart(ctx, item, container.StopOptions{})
		case constant.ContainerOpKill:
			err = c.cli.ContainerKill(ctx, item, "SIGKILL")
		case constant.ContainerOpPause:
			err = c.cli.ContainerPause(ctx, item)
		case constant.ContainerOpUnpause:
			err = c.cli.ContainerUnpause(ctx, item)
		case constant.ContainerOpRemove:
			err = c.cli.ContainerRemove(ctx, item, container.RemoveOptions{RemoveVolumes: true, Force: true})
		}
	}
	return err
}
