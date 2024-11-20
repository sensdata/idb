package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/sensdata/idb/core/utils/common"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/websocket"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type ContainerService struct{}

type IContainerService interface {
	Inspect(req model.Inspect) (string, error)
	Prune(req model.Prune) (model.PruneResult, error)

	ContainerQuery(req model.QueryContainer) (*model.PageResult, error)
	ContainerList() ([]string, error)
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

func NewIContainerService() IContainerService {
	return &ContainerService{}
}

func (u *ContainerService) Inspect(req model.Inspect) (string, error) {
	client, err := NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	var inspectInfo interface{}
	switch req.Type {
	case "container":
		inspectInfo, err = client.ContainerInspect(context.Background(), req.ID)
	case "image":
		inspectInfo, _, err = client.ImageInspectWithRaw(context.Background(), req.ID)
	case "network":
		inspectInfo, err = client.NetworkInspect(context.TODO(), req.ID, network.InspectOptions{})
	case "volume":
		inspectInfo, err = client.VolumeInspect(context.TODO(), req.ID)
	}
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(inspectInfo)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (u *ContainerService) Prune(req model.Prune) (model.PruneResult, error) {
	report := model.PruneResult{}
	client, err := NewDockerClient()
	if err != nil {
		return report, err
	}
	defer client.Close()
	pruneFilters := filters.NewArgs()
	if req.WithTagAll {
		pruneFilters.Add("dangling", "false")
		if req.PruneType != "image" {
			pruneFilters.Add("until", "24h")
		}
	}
	switch req.PruneType {
	case "container":
		rep, err := client.ContainersPrune(context.Background(), pruneFilters)
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.ContainersDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	case "image":
		rep, err := client.ImagesPrune(context.Background(), pruneFilters)
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.ImagesDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	case "network":
		rep, err := client.NetworksPrune(context.Background(), pruneFilters)
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.NetworksDeleted)
	case "volume":
		versions, err := client.ServerVersion(context.Background())
		if err != nil {
			return report, err
		}
		if common.ComparePanelVersion(versions.APIVersion, "1.42") {
			pruneFilters.Add("all", "true")
		}
		rep, err := client.VolumesPrune(context.Background(), pruneFilters)
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.VolumesDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	case "buildcache":
		rep, err := client.BuildCachePrune(context.Background(), types.BuildCachePruneOptions{})
		if err != nil {
			return report, err
		}
		report.DeletedNumber = len(rep.CachesDeleted)
		report.SpaceReclaimed = int(rep.SpaceReclaimed)
	}
	return report, nil
}

func (u *ContainerService) ContainerQuery(req model.QueryContainer) (*model.PageResult, error) {
	var (
		result  model.PageResult
		records []types.Container
		list    []types.Container
	)
	client, err := NewDockerClient()
	if err != nil {
		return &result, err
	}
	defer client.Close()
	options := container.ListOptions{
		All: true,
	}
	if len(req.Filters) != 0 {
		options.Filters = filters.NewArgs()
		options.Filters.Add("label", req.Filters)
	}
	containers, err := client.ContainerList(context.Background(), options)
	if err != nil {
		return &result, err
	}
	if req.ExcludeAppStore {
		for _, item := range containers {
			if created, ok := item.Labels[composeCreatedBy]; ok && created == "Apps" {
				continue
			}
			list = append(list, item)
		}
	} else {
		list = containers
	}

	if len(req.Name) != 0 {
		length, count := len(list), 0
		for count < length {
			if !strings.Contains(list[count].Names[0][1:], req.Name) {
				list = append(list[:count], list[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	if req.State != "all" {
		length, count := len(list), 0
		for count < length {
			if list[count].State != req.State {
				list = append(list[:count], list[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	switch req.OrderBy {
	case "name":
		sort.Slice(list, func(i, j int) bool {
			if req.Order == constant.OrderAsc {
				return list[i].Names[0][1:] < list[j].Names[0][1:]
			}
			return list[i].Names[0][1:] > list[j].Names[0][1:]
		})
	case "state":
		sort.Slice(list, func(i, j int) bool {
			if req.Order == constant.OrderAsc {
				return list[i].State < list[j].State
			}
			return list[i].State > list[j].State
		})
	default:
		sort.Slice(list, func(i, j int) bool {
			if req.Order == constant.OrderAsc {
				return list[i].Created < list[j].Created
			}
			return list[i].Created > list[j].Created
		})
	}

	total, start, end := len(list), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]types.Container, 0)
	} else {
		if end >= total {
			end = total
		}
		records = list[start:end]
	}

	backDatas := make([]model.ContainerInfo, len(records))
	for i := 0; i < len(records); i++ {
		item := records[i]
		IsFromCompose := false
		if _, ok := item.Labels[composeProjectLabel]; ok {
			IsFromCompose = true
		}
		IsFromApp := false
		if created, ok := item.Labels[composeCreatedBy]; ok && created == "Apps" {
			IsFromApp = true
		}

		ports := loadContainerPort(item.Ports)
		info := model.ContainerInfo{
			ContainerID:   item.ID,
			CreateTime:    time.Unix(item.Created, 0).Format("2006-01-02 15:04:05"),
			Name:          item.Names[0][1:],
			ImageId:       strings.Split(item.ImageID, ":")[1],
			ImageName:     item.Image,
			State:         item.State,
			RunTime:       item.Status,
			Ports:         ports,
			IsFromApp:     IsFromApp,
			IsFromCompose: IsFromCompose,
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
		backDatas[i] = info
		if item.NetworkSettings != nil && len(item.NetworkSettings.Networks) > 0 {
			networks := make([]string, 0, len(item.NetworkSettings.Networks))
			for key := range item.NetworkSettings.Networks {
				networks = append(networks, item.NetworkSettings.Networks[key].IPAddress)
			}
			sort.Strings(networks)
			backDatas[i].Network = networks
		}
	}

	result.Total = int64(total)
	result.Items = backDatas

	return &result, nil
}

func (u *ContainerService) ContainerList() ([]string, error) {
	client, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	containers, err := client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	var datas []string
	for _, container := range containers {
		for _, name := range container.Names {
			if len(name) != 0 {
				datas = append(datas, strings.TrimPrefix(name, "/"))
			}
		}
	}

	return datas, nil
}

func (u *ContainerService) ContainerCreate(req model.ContainerOperate) error {
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := context.Background()
	newContainer, _ := client.ContainerInspect(ctx, req.Name)
	if newContainer.ContainerJSONBase != nil {
		return errors.New(constant.ErrContainerName)
	}

	if !checkImageExist(client, req.Image) || req.ForcePull {
		if err := pullImages(ctx, client, req.Image); err != nil {
			if !req.ForcePull {
				return err
			}
			global.LOG.Error("force pull image %s failed, err: %v", req.Image, err)
		}
	}
	imageInfo, _, err := client.ImageInspectWithRaw(ctx, req.Image)
	if err != nil {
		return err
	}
	if len(req.Entrypoint) == 0 {
		req.Entrypoint = imageInfo.Config.Entrypoint
	}
	if len(req.Cmd) == 0 {
		req.Cmd = imageInfo.Config.Cmd
	}
	config, hostConf, networkConf, err := loadConfigInfo(true, req, nil)
	if err != nil {
		return err
	}
	global.LOG.Info("new container info %s has been made, now start to create", req.Name)
	con, err := client.ContainerCreate(ctx, config, hostConf, networkConf, &v1.Platform{}, req.Name)
	if err != nil {
		_ = client.ContainerRemove(ctx, req.Name, container.RemoveOptions{RemoveVolumes: true, Force: true})
		return err
	}
	global.LOG.Info("create container %s successful! now check if the container is started and delete the container information if it is not.", req.Name)
	if err := client.ContainerStart(ctx, con.ID, container.StartOptions{}); err != nil {
		_ = client.ContainerRemove(ctx, req.Name, container.RemoveOptions{RemoveVolumes: true, Force: true})
		return fmt.Errorf("create successful but start failed, err: %v", err)
	}
	return nil
}

func (u *ContainerService) ContainerUpdate(req model.ContainerOperate) error {
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := context.Background()
	newContainer, _ := client.ContainerInspect(ctx, req.Name)
	if newContainer.ContainerJSONBase != nil && newContainer.ID != req.ContainerID {
		return errors.New(constant.ErrContainerName)
	}

	oldContainer, err := client.ContainerInspect(ctx, req.ContainerID)
	if err != nil {
		return err
	}
	if !checkImageExist(client, req.Image) || req.ForcePull {
		if err := pullImages(ctx, client, req.Image); err != nil {
			if !req.ForcePull {
				return err
			}
			return fmt.Errorf("pull image %s failed, err: %v", req.Image, err)
		}
	}

	if err := client.ContainerRemove(ctx, req.ContainerID, container.RemoveOptions{Force: true}); err != nil {
		return err
	}

	config, hostConf, networkConf, err := loadConfigInfo(false, req, &oldContainer)
	if err != nil {
		reCreateAfterUpdate(req.Name, client, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings)
		return err
	}

	global.LOG.Info("new container info %s has been update, now start to recreate", req.Name)
	con, err := client.ContainerCreate(ctx, config, hostConf, networkConf, &v1.Platform{}, req.Name)
	if err != nil {
		reCreateAfterUpdate(req.Name, client, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings)
		return fmt.Errorf("update container failed, err: %v", err)
	}
	global.LOG.Info("update container %s successful! now check if the container is started.", req.Name)
	if err := client.ContainerStart(ctx, con.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("update successful but start failed, err: %v", err)
	}

	return nil
}

func (u *ContainerService) ContainerUpgrade(req model.ContainerUpgrade) error {
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := context.Background()
	oldContainer, err := client.ContainerInspect(ctx, req.Name)
	if err != nil {
		return err
	}
	if !checkImageExist(client, req.Image) || req.ForcePull {
		if err := pullImages(ctx, client, req.Image); err != nil {
			if !req.ForcePull {
				return err
			}
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
	if err := client.ContainerRemove(ctx, req.Name, container.RemoveOptions{Force: true}); err != nil {
		return err
	}

	global.LOG.Info("new container info %s has been update, now start to recreate", req.Name)
	con, err := client.ContainerCreate(ctx, config, hostConf, &networkConf, &v1.Platform{}, req.Name)
	if err != nil {
		reCreateAfterUpdate(req.Name, client, oldContainer.Config, oldContainer.HostConfig, oldContainer.NetworkSettings)
		return fmt.Errorf("upgrade container failed, err: %v", err)
	}
	global.LOG.Info("upgrade container %s successful! now check if the container is started.", req.Name)
	if err := client.ContainerStart(ctx, con.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("upgrade successful but start failed, err: %v", err)
	}

	return nil
}

func (u *ContainerService) ContainerInfo(containerID string) (*model.ContainerOperate, error) {
	client, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	ctx := context.Background()
	oldContainer, err := client.ContainerInspect(ctx, containerID)
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
		data.Labels = append(data.Labels, fmt.Sprintf("%s=%s", key, val))
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
	data.Volumes = loadVolumeBinds(oldContainer.Mounts)

	return &data, nil
}

func (u *ContainerService) ContainerResourceUsage() ([]model.ContainerResourceUsage, error) {
	client, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	list, err := client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	var datas []model.ContainerResourceUsage
	var wg sync.WaitGroup
	wg.Add(len(list))
	for i := 0; i < len(list); i++ {
		go func(item types.Container) {
			datas = append(datas, loadCpuAndMem(client, item.ID))
			wg.Done()
		}(list[i])
	}
	wg.Wait()
	return datas, nil
}

func (u *ContainerService) ContainerResourceLimit() (*model.ContainerResourceLimit, error) {
	cpuCounts, err := cpu.Counts(true)
	if err != nil {
		return nil, fmt.Errorf("load cpu limit failed, err: %v", err)
	}
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("load memory limit failed, err: %v", err)
	}

	data := model.ContainerResourceLimit{
		CPU:    cpuCounts,
		Memory: memoryInfo.Total,
	}
	return &data, nil
}

func (u *ContainerService) ContainerStats(id string) (*model.ContainerStats, error) {
	client, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	res, err := client.ContainerStats(context.TODO(), id, false)
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
	var stats *types.StatsJSON
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, err
	}
	var data model.ContainerStats
	data.CPUPercent = calculateCPUPercentUnix(stats)
	data.IORead, data.IOWrite = calculateBlockIO(stats.BlkioStats)
	data.Memory = float64(stats.MemoryStats.Usage) / 1024 / 1024
	if cache, ok := stats.MemoryStats.Stats["cache"]; ok {
		data.Cache = float64(cache) / 1024 / 1024
	}
	data.NetworkRX, data.NetworkTX = calculateNetwork(stats.Networks)
	data.ShotTime = stats.Read
	return &data, nil
}

func (u *ContainerService) ContainerRename(req model.Rename) error {
	ctx := context.Background()
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()

	newContainer, _ := client.ContainerInspect(ctx, req.NewName)
	if newContainer.ContainerJSONBase != nil {
		return errors.New(constant.ErrContainerName)
	}
	return client.ContainerRename(ctx, req.Name, req.NewName)
}

func (u *ContainerService) ContainerLogClean(containerID string) error {
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := context.Background()
	containerItem, err := client.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}
	if err := client.ContainerStop(ctx, containerItem.ID, container.StopOptions{}); err != nil {
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

	if err := client.ContainerStart(ctx, containerItem.ID, container.StartOptions{}); err != nil {
		return err
	}
	return nil
}

func (u *ContainerService) ContainerOperation(req model.ContainerOperation) error {
	var err error
	ctx := context.Background()
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	for _, item := range req.Names {
		global.LOG.Info("start container %s operation %s", item, req.Operation)
		switch req.Operation {
		case constant.ContainerOpStart:
			err = client.ContainerStart(ctx, item, container.StartOptions{})
		case constant.ContainerOpStop:
			err = client.ContainerStop(ctx, item, container.StopOptions{})
		case constant.ContainerOpRestart:
			err = client.ContainerRestart(ctx, item, container.StopOptions{})
		case constant.ContainerOpKill:
			err = client.ContainerKill(ctx, item, "SIGKILL")
		case constant.ContainerOpPause:
			err = client.ContainerPause(ctx, item)
		case constant.ContainerOpUnpause:
			err = client.ContainerUnpause(ctx, item)
		case constant.ContainerOpRemove:
			err = client.ContainerRemove(ctx, item, container.RemoveOptions{RemoveVolumes: true, Force: true})
		}
	}
	return err
}

func (u *ContainerService) ContainerLogs(wsConn *websocket.Conn, containerType, container, since, tail string, follow bool) error {
	defer func() { wsConn.Close() }()
	if utils.CheckIllegal(container, since, tail) {
		return errors.New(constant.ErrCmdIllegal)
	}
	commandName := "docker"
	commandArg := []string{"logs", container}
	if containerType == "compose" {
		commandName = "docker-compose"
		commandArg = []string{"-f", container, "logs"}
	}
	if tail != "0" {
		commandArg = append(commandArg, "--tail")
		commandArg = append(commandArg, tail)
	}
	if since != "all" {
		commandArg = append(commandArg, "--since")
		commandArg = append(commandArg, since)
	}
	if follow {
		commandArg = append(commandArg, "-f")
	}
	if !follow {
		cmd := exec.Command(commandName, commandArg...)
		cmd.Stderr = cmd.Stdout
		stdout, _ := cmd.CombinedOutput()
		if !utf8.Valid(stdout) {
			return errors.New("invalid utf8")
		}
		if err := wsConn.WriteMessage(websocket.TextMessage, stdout); err != nil {
			global.LOG.Error("send message with log to ws failed, err: %v", err)
		}
		return nil
	}

	cmd := exec.Command(commandName, commandArg...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		return err
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		return err
	}
	exitCh := make(chan struct{})
	go func() {
		_, wsData, _ := wsConn.ReadMessage()
		if string(wsData) == "close conn" {
			_ = cmd.Process.Signal(syscall.SIGTERM)
			exitCh <- struct{}{}
		}
	}()

	go func() {
		buffer := make([]byte, 1024)
		for {
			select {
			case <-exitCh:
				return
			default:
				n, err := stdout.Read(buffer)
				if err != nil {
					if err == io.EOF {
						return
					}
					global.LOG.Error("read bytes from log failed, err: %v", err)
					return
				}
				if !utf8.Valid(buffer[:n]) {
					continue
				}
				if err = wsConn.WriteMessage(websocket.TextMessage, buffer[:n]); err != nil {
					global.LOG.Error("send message with log to ws failed, err: %v", err)
					return
				}
			}
		}
	}()
	_ = cmd.Wait()
	return nil
}

func stringsToMap(list []string) map[string]string {
	var labelMap = make(map[string]string)
	for _, label := range list {
		if strings.Contains(label, "=") {
			sps := strings.SplitN(label, "=", 2)
			labelMap[sps[0]] = sps[1]
		}
	}
	return labelMap
}

func calculateCPUPercentUnix(stats *types.StatsJSON) float64 {
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
func calculateMemPercentUnix(memStats types.MemoryStats) float64 {
	memPercent := 0.0
	memUsage := float64(memStats.Usage)
	memLimit := float64(memStats.Limit)
	if memUsage > 0.0 && memLimit > 0.0 {
		memPercent = (memUsage / memLimit) * 100.0
	}
	return memPercent
}
func calculateBlockIO(blkio types.BlkioStats) (blkRead float64, blkWrite float64) {
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
func calculateNetwork(network map[string]types.NetworkStats) (float64, float64) {
	var rx, tx float64

	for _, v := range network {
		rx += float64(v.RxBytes) / 1024
		tx += float64(v.TxBytes) / 1024
	}
	return rx, tx
}

func checkImageExist(client *client.Client, imageName string) bool {
	images, err := client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		fmt.Println(err)
		return false
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName || tag == imageName+":latest" {
				return true
			}
		}
	}
	return false
}

func pullImages(ctx context.Context, client *client.Client, imageName string) error {
	options := image.PullOptions{}
	authConfig := registry.AuthConfig{
		Username: "username", // TODO: 推送的用户名,
		Password: "password", // TODO: 推送的密码
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	options.RegistryAuth = authStr
	out, err := client.ImagePull(ctx, imageName, options)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(io.Discard, out)
	if err != nil {
		return err
	}
	return nil
}

func loadCpuAndMem(client *client.Client, container string) model.ContainerResourceUsage {
	data := model.ContainerResourceUsage{
		ContainerID: container,
	}
	res, err := client.ContainerStats(context.Background(), container, false)
	if err != nil {
		return data
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return data
	}
	var stats *types.StatsJSON
	if err := json.Unmarshal(body, &stats); err != nil {
		return data
	}

	data.CPUTotalUsage = stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage
	data.SystemUsage = stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage
	data.CPUPercent = calculateCPUPercentUnix(stats)
	data.PercpuUsage = len(stats.CPUStats.CPUUsage.PercpuUsage)

	data.MemoryCache = stats.MemoryStats.Stats["cache"]
	data.MemoryUsage = stats.MemoryStats.Usage
	data.MemoryLimit = stats.MemoryStats.Limit

	data.MemoryPercent = calculateMemPercentUnix(stats.MemoryStats)
	return data
}

func checkPortStats(ports []model.PortHelper) (nat.PortMap, error) {
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

func loadConfigInfo(isCreate bool, req model.ContainerOperate, oldContainer *types.ContainerJSON) (*container.Config, *container.HostConfig, *network.NetworkingConfig, error) {
	var config container.Config
	var hostConf container.HostConfig
	if !isCreate {
		config = *oldContainer.Config
		hostConf = *oldContainer.HostConfig
	}
	var networkConf network.NetworkingConfig

	portMap, err := checkPortStats(req.ExposedPorts)
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
	config.Labels = stringsToMap(req.Labels)
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
			return nil, nil, nil, fmt.Errorf("Please set up the network")
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

func reCreateAfterUpdate(name string, client *client.Client, config *container.Config, hostConf *container.HostConfig, networkConf *types.NetworkSettings) {
	ctx := context.Background()

	var oldNetworkConf network.NetworkingConfig
	if networkConf != nil {
		for networkKey := range networkConf.Networks {
			oldNetworkConf.EndpointsConfig = map[string]*network.EndpointSettings{networkKey: {}}
			break
		}
	}

	oldContainer, err := client.ContainerCreate(ctx, config, hostConf, &oldNetworkConf, &v1.Platform{}, name)
	if err != nil {
		global.LOG.Error("recreate after container update failed, err: %v", err)
		return
	}
	if err := client.ContainerStart(ctx, oldContainer.ID, container.StartOptions{}); err != nil {
		global.LOG.Error("restart after container update failed, err: %v", err)
	}
	global.LOG.Error("recreate after container update successful")
}

func loadVolumeBinds(binds []types.MountPoint) []model.VolumeHelper {
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

func loadContainerPort(ports []types.Port) []string {
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
	list1 := simplifyPort(ipv4Ports)
	list2 := simplifyPort(ipv6Ports)
	return append(list1, list2...)
}
func simplifyPort(ports []types.Port) []string {
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
