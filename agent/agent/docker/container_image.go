package docker

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/archive"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

func (u *ContainerService) ImagePage(req model.SearchPageInfo) (*model.PageResult, error) {
	var (
		result    model.PageResult
		list      []image.Summary
		records   []model.Image
		backDatas []model.Image
	)
	client, err := NewDockerClient()
	if err != nil {
		return &result, err
	}
	defer client.Close()
	list, err = client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return &result, err
	}
	containers, _ := client.ContainerList(context.Background(), container.ListOptions{All: true})
	if len(req.Info) != 0 {
		length, count := len(list), 0
		for count < length {
			hasTag := false
			for _, tag := range list[count].RepoTags {
				if strings.Contains(tag, req.Info) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				list = append(list[:count], list[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}

	for _, image := range list {
		size := formatFileSize(image.Size)
		records = append(records, model.Image{
			ID:        image.ID,
			Tags:      image.RepoTags,
			IsUsed:    checkUsed(image.ID, containers),
			CreatedAt: time.Unix(image.Created, 0),
			Size:      size,
		})
	}
	total, start, end := len(records), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		backDatas = make([]model.Image, 0)
	} else {
		if end >= total {
			end = total
		}
		backDatas = records[start:end]
	}

	result.Total = int64(total)
	result.Items = backDatas

	return &result, nil
}

func (u *ContainerService) ImageList() ([]model.Options, error) {
	var (
		list      []image.Summary
		backDatas []model.Options
	)
	client, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	list, err = client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, image := range list {
		for _, tag := range image.RepoTags {
			backDatas = append(backDatas, model.Options{
				Option: tag,
			})
		}
	}
	return backDatas, nil
}

func (u *ContainerService) ImageBuild(req model.ImageBuild) (string, error) {
	client, err := NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	fileName := "Dockerfile"
	if req.From == "edit" {
		dir := fmt.Sprintf("%s/docker/build/%s", constant.AgentDataDir, strings.ReplaceAll(req.Name, ":", "_"))
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return "", err
			}
		}

		pathItem := fmt.Sprintf("%s/Dockerfile", dir)
		file, err := os.OpenFile(pathItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return "", err
		}
		defer file.Close()
		write := bufio.NewWriter(file)
		_, _ = write.WriteString(string(req.Dockerfile))
		write.Flush()
		req.Dockerfile = dir
	} else {
		fileName = path.Base(req.Dockerfile)
		req.Dockerfile = path.Dir(req.Dockerfile)
	}
	tar, err := archive.TarWithOptions(req.Dockerfile+"/", &archive.TarOptions{})
	if err != nil {
		return "", err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: fileName,
		Tags:       []string{req.Name},
		Remove:     true,
		Labels:     stringsToMap(req.Tags),
	}

	dockerLogDir := path.Join(constant.AgentLogDir, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return "", err
		}
	}
	logItem := fmt.Sprintf("%s/image_build_%s_%s.log", dockerLogDir, strings.ReplaceAll(req.Name, ":", "_"), time.Now().Format("20060102150405"))
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	go func() {
		defer file.Close()
		defer tar.Close()
		res, err := client.ImageBuild(context.Background(), tar, opts)
		if err != nil {
			global.LOG.Error("build image %s failed, err: %v", req.Name, err)
			_, _ = file.WriteString("image build failed!")
			return
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			global.LOG.Error("build image %s failed, err: %v", req.Name, err)
			_, _ = file.WriteString(fmt.Sprintf("build image %s failed, err: %v", req.Name, err))
			_, _ = file.WriteString("image build failed!")
			return
		}

		if strings.Contains(string(body), "errorDetail") || strings.Contains(string(body), "error:") {
			global.LOG.Error("build image %s failed", req.Name)
			_, _ = file.Write(body)
			_, _ = file.WriteString("image build failed!")
			return
		}
		global.LOG.Info("build image %s successful!", req.Name)
		_, _ = file.Write(body)
		_, _ = file.WriteString("image build successful!")
	}()

	return path.Base(logItem), nil
}

func (u *ContainerService) ImagePull(req model.ImagePull) (string, error) {
	client, err := NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	dockerLogDir := path.Join(constant.AgentLogDir, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return "", err
		}
	}
	imageItemName := strings.ReplaceAll(path.Base(req.ImageName), ":", "_")
	logItem := fmt.Sprintf("%s/image_pull_%s_%s.log", dockerLogDir, imageItemName, time.Now().Format("20060102150405"))
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}

	go func() {
		defer file.Close()
		out, err := client.ImagePull(context.TODO(), req.ImageName, image.PullOptions{})
		if err != nil {
			global.LOG.Error("image %s pull failed, err: %v", req.ImageName, err)
			return
		}
		defer out.Close()
		global.LOG.Info("pull image %s successful!", req.ImageName)
		_, _ = io.Copy(file, out)
	}()
	return path.Base(logItem), nil

}

func (u *ContainerService) ImageLoad(req model.ImageLoad) error {
	file, err := os.Open(req.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	res, err := client.ImageLoad(context.TODO(), file, true)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), "Error") {
		return errors.New(string(content))
	}
	return nil
}

func (u *ContainerService) ImageSave(req model.ImageSave) error {
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()

	out, err := client.ImageSave(context.TODO(), []string{req.TagName})
	if err != nil {
		return err
	}
	defer out.Close()
	file, err := os.OpenFile(fmt.Sprintf("%s/%s.tar", req.Path, req.Name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = io.Copy(file, out); err != nil {
		return err
	}
	return nil
}

func (u *ContainerService) ImageTag(req model.ImageTag) error {
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.ImageTag(context.TODO(), req.SourceID, req.TargetName); err != nil {
		return err
	}
	return nil
}

func (u *ContainerService) ImagePush(req model.ImagePush) (string, error) {
	client, err := NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	// repo, err := imageRepoRepo.Get(commonRepo.WithByID(req.RepoID))
	// if err != nil {
	// 	return "", err
	// }
	options := image.PushOptions{All: true}
	authConfig := registry.AuthConfig{
		Username: "username", // TODO: 推送的用户名,
		Password: "password", // TODO: 推送的密码
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	options.RegistryAuth = authStr
	newName := fmt.Sprintf("%s/%s", "docker.io", req.Name)
	if newName != req.TagName {
		if err := client.ImageTag(context.TODO(), req.TagName, newName); err != nil {
			return "", err
		}
	}

	dockerLogDir := constant.AgentLogDir + "/docker_logs"
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return "", err
		}
	}
	imageItemName := strings.ReplaceAll(path.Base(req.Name), ":", "_")
	logItem := fmt.Sprintf("%s/image_push_%s_%s.log", dockerLogDir, imageItemName, time.Now().Format("20060102150405"))
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	go func() {
		defer file.Close()
		out, err := client.ImagePush(context.TODO(), newName, options)
		if err != nil {
			global.LOG.Error("image %s push failed, err: %v", req.TagName, err)
			_, _ = file.WriteString("image push failed!")
			return
		}
		defer out.Close()
		global.LOG.Info("push image %s successful!", req.Name)
		_, _ = io.Copy(file, out)
		_, _ = file.WriteString("image push successful!")
	}()

	return path.Base(logItem), nil
}

func (u *ContainerService) ImageRemove(req model.BatchDelete) error {
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	for _, id := range req.Names {
		if _, err := client.ImageRemove(context.TODO(), id, image.RemoveOptions{Force: req.Force, PruneChildren: true}); err != nil {
			if strings.Contains(err.Error(), "image is being used") || strings.Contains(err.Error(), "is using") {
				if strings.Contains(id, "sha256:") {
					return errors.New(constant.ErrObjectInUsed)
				}
				return errors.New(constant.ErrInUsed)
			}
			return err
		}
	}
	return nil
}

func formatFileSize(fileSize int64) (size string) {
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

func checkUsed(imageID string, containers []types.Container) bool {
	for _, container := range containers {
		if container.ImageID == imageID {
			return true
		}
	}
	return false
}
