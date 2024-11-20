package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/model"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type DockerClient struct {
	cli *client.Client
}

func NewClient() (DockerClient, error) {
	var settingItem model.Setting
	_ = global.DB.Where("key = ?", "DockerSockPath").First(&settingItem).Error
	if len(settingItem.Value) == 0 {
		global.LOG.Info("Using default docker host")
		settingItem.Value = "unix:///var/run/docker.sock"
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(settingItem.Value), client.WithAPIVersionNegotiation())
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

func NewDockerClient() (*client.Client, error) {
	var settingItem model.Setting
	_ = global.DB.Where("key = ?", "DockerSockPath").First(&settingItem).Error
	if len(settingItem.Value) == 0 {
		global.LOG.Info("Using default docker host")
		settingItem.Value = "unix:///var/run/docker.sock"
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(settingItem.Value), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (c DockerClient) ListContainersByName(names []string) ([]types.Container, error) {
	var (
		options  container.ListOptions
		namesMap = make(map[string]bool)
		res      []types.Container
	)
	options.All = true
	if len(names) > 0 {
		var array []filters.KeyValuePair
		for _, n := range names {
			namesMap["/"+n] = true
			array = append(array, filters.Arg("name", n))
		}
		options.Filters = filters.NewArgs(array...)
	}
	containers, err := c.cli.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}
	for _, con := range containers {
		if _, ok := namesMap[con.Names[0]]; ok {
			res = append(res, con)
		}
	}
	return res, nil
}

func (c DockerClient) InspectContainer(containerID string) (types.ContainerJSON, error) {
	return c.cli.ContainerInspect(context.Background(), containerID)
}

func (c DockerClient) DeleteImage(imageID string) error {
	if _, err := c.cli.ImageRemove(context.Background(), imageID, image.RemoveOptions{Force: true}); err != nil {
		return err
	}
	return nil
}

func (c DockerClient) PullImage(imageName string, force bool) error {
	if !force {
		exist, err := c.CheckImageExist(imageName)
		if err != nil {
			return err
		}
		if exist {
			return nil
		}
	}
	if _, err := c.cli.ImagePull(context.Background(), imageName, image.PullOptions{}); err != nil {
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

func (c DockerClient) CheckImageExist(imageName string) (bool, error) {
	filter := filters.NewArgs()
	filter.Add("reference", imageName)
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

func CreateDefaultDockerNetwork() error {
	cli, err := NewClient()
	if err != nil {
		global.LOG.Error("init docker client error %s", err.Error())
		return err
	}
	defer cli.Close()
	if !cli.NetworkExist("idb-network") {
		if err := cli.CreateNetwork("idb-network"); err != nil {
			global.LOG.Error("create default docker network  error %s", err.Error())
			return err
		}
	}
	return nil
}

func (c DockerClient) CreateNetwork(name string) error {
	_, err := c.cli.NetworkCreate(context.Background(), name, types.NetworkCreate{
		Driver: "bridge",
	})
	return err
}
