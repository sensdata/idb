package client

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
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils/common"
)

func (c DockerClient) ImagePage(req model.SearchPageInfo) (*model.PageResult, error) {
	var (
		result    model.PageResult
		list      []image.Summary
		records   []model.Image
		backDatas []model.Image
	)
	list, err := c.cli.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return &result, err
	}
	containers, _ := c.cli.ContainerList(context.Background(), container.ListOptions{All: true})
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
		size := c.formatFileSize(image.Size)
		records = append(records, model.Image{
			ID:        image.ID,
			Tags:      image.RepoTags,
			IsUsed:    c.checkUsed(image.ID, containers),
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

func (c DockerClient) ImageList() (*model.PageResult, error) {
	var (
		result    model.PageResult
		list      []image.Summary
		backDatas []model.Options
	)
	list, err := c.cli.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return &result, err
	}
	for _, image := range list {
		for _, tag := range image.RepoTags {
			backDatas = append(backDatas, model.Options{
				Option: tag,
			})
		}
	}
	result.Total = int64(len(backDatas))
	result.Items = backDatas
	return &result, nil
}

func (c DockerClient) ImageBuild(req model.ImageBuild, dataDir string, logDir string, logger *log.Log) (*model.ImageOperationResult, error) {
	var result model.ImageOperationResult
	fileName := "Dockerfile"
	if req.From == "edit" {
		dir := fmt.Sprintf("%s/docker/build/%s", dataDir, strings.ReplaceAll(req.Name, ":", "_"))
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return &result, err
			}
		}

		pathItem := fmt.Sprintf("%s/Dockerfile", dir)
		file, err := os.OpenFile(pathItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return &result, err
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
		return &result, err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: fileName,
		Tags:       []string{req.Name},
		Remove:     true,
		Labels:     common.StringsToMap(req.Tags),
	}

	dockerLogDir := path.Join(logDir, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return &result, err
		}
	}
	logItem := fmt.Sprintf("%s/image_build_%s_%s.log", dockerLogDir, strings.ReplaceAll(req.Name, ":", "_"), time.Now().Format("20060102150405"))
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return &result, err
	}
	go func() {
		defer file.Close()
		defer tar.Close()
		res, err := c.cli.ImageBuild(context.Background(), tar, opts)
		if err != nil {
			logger.Error("build image %s failed, err: %v", req.Name, err)
			_, _ = file.WriteString("image build failed!")
			return
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Error("build image %s failed, err: %v", req.Name, err)
			_, _ = file.WriteString(fmt.Sprintf("build image %s failed, err: %v", req.Name, err))
			_, _ = file.WriteString("image build failed!")
			return
		}

		if strings.Contains(string(body), "errorDetail") || strings.Contains(string(body), "error:") {
			logger.Error("build image %s failed", req.Name)
			_, _ = file.Write(body)
			_, _ = file.WriteString("image build failed!")
			return
		}
		logger.Info("build image %s successful!", req.Name)
		_, _ = file.Write(body)
		_, _ = file.WriteString("image build successful!")
	}()
	result.Result = path.Base(logItem)
	return &result, nil
}

func (c DockerClient) ImagePull(req model.ImagePull, logDir string, logger *log.Log) (*model.ImageOperationResult, error) {
	var result model.ImageOperationResult
	dockerLogDir := path.Join(logDir, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return &result, err
		}
	}
	imageItemName := strings.ReplaceAll(path.Base(req.ImageName), ":", "_")
	logItem := fmt.Sprintf("%s/image_pull_%s_%s.log", dockerLogDir, imageItemName, time.Now().Format("20060102150405"))
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return &result, err
	}

	go func() {
		defer file.Close()
		out, err := c.cli.ImagePull(context.TODO(), req.ImageName, image.PullOptions{})
		if err != nil {
			logger.Error("image %s pull failed, err: %v", req.ImageName, err)
			return
		}
		defer out.Close()
		logger.Info("pull image %s successful!", req.ImageName)
		_, _ = io.Copy(file, out)
	}()

	result.Result = path.Base(logItem)
	return &result, nil

}

func (c DockerClient) ImageLoad(req model.ImageLoad) (*model.ImageOperationResult, error) {
	var result model.ImageOperationResult
	file, err := os.Open(req.Path)
	if err != nil {
		global.LOG.Error("Faield to open file %v", err)
		return &result, err
	}
	defer file.Close()
	res, err := c.cli.ImageLoad(context.TODO(), file, true)
	if err != nil {
		global.LOG.Error("Faield to load image file %v", err)
		return &result, err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		global.LOG.Error("Faield to read image file %v", err)
		return &result, err
	}
	global.LOG.Info("image load result: %s", content)
	result.Result = string(content)
	if strings.Contains(string(content), "Error") {
		return &result, errors.New(string(content))
	}
	return &result, nil
}

func (c DockerClient) ImageSave(req model.ImageSave) error {
	out, err := c.cli.ImageSave(context.TODO(), []string{req.TagName})
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

func (c DockerClient) ImageTag(req model.ImageTag) error {
	if err := c.cli.ImageTag(context.TODO(), req.SourceID, req.TargetName); err != nil {
		return err
	}
	return nil
}

func (c DockerClient) ImagePush(req model.ImagePush, logDir string, logger *log.Log) (*model.ImageOperationResult, error) {
	var result model.ImageOperationResult
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
		return &result, err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	options.RegistryAuth = authStr
	newName := fmt.Sprintf("%s/%s", "docker.io", req.Name)
	if newName != req.TagName {
		if err := c.cli.ImageTag(context.TODO(), req.TagName, newName); err != nil {
			return &result, err
		}
	}

	dockerLogDir := logDir + "/docker_logs"
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return &result, err
		}
	}
	imageItemName := strings.ReplaceAll(path.Base(req.Name), ":", "_")
	logItem := fmt.Sprintf("%s/image_push_%s_%s.log", dockerLogDir, imageItemName, time.Now().Format("20060102150405"))
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return &result, err
	}
	go func() {
		defer file.Close()
		out, err := c.cli.ImagePush(context.TODO(), newName, options)
		if err != nil {
			logger.Error("image %s push failed, err: %v", req.TagName, err)
			_, _ = file.WriteString("image push failed!")
			return
		}
		defer out.Close()
		logger.Info("push image %s successful!", req.Name)
		_, _ = io.Copy(file, out)
		_, _ = file.WriteString("image push successful!")
	}()

	result.Result = path.Base(logItem)
	return &result, nil
}

func (c DockerClient) ImageRemove(req model.BatchDelete) error {
	for _, id := range req.Names {
		if _, err := c.cli.ImageRemove(context.TODO(), id, image.RemoveOptions{Force: req.Force, PruneChildren: true}); err != nil {
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
