package client

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils/common"
)

func (c DockerClient) VolumePage(req model.SearchPageInfo) (*model.PageResult, error) {
	var result model.PageResult
	list, err := c.cli.VolumeList(context.TODO(), volume.ListOptions{})
	if err != nil {
		return &result, err
	}
	if len(req.Info) != 0 {
		length, count := len(list.Volumes), 0
		for count < length {
			if !strings.Contains(list.Volumes[count].Name, req.Info) {
				list.Volumes = append(list.Volumes[:count], list.Volumes[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	var (
		data    []model.Volume
		records []*volume.Volume
	)
	sort.Slice(list.Volumes, func(i, j int) bool {
		return list.Volumes[i].CreatedAt > list.Volumes[j].CreatedAt
	})
	total, start, end := len(list.Volumes), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]*volume.Volume, 0)
	} else {
		if end >= total {
			end = total
		}
		records = list.Volumes[start:end]
	}

	nyc, _ := time.LoadLocation(common.LoadTimeZone())
	for _, item := range records {
		var labels []model.KeyValue
		for key, val := range item.Labels {
			labels = append(labels, model.KeyValue{Key: key, Value: val})
		}
		var createTime time.Time
		if strings.Contains(item.CreatedAt, "Z") {
			createTime, _ = time.ParseInLocation("2006-01-02T15:04:05Z", item.CreatedAt, nyc)
		} else if strings.Contains(item.CreatedAt, "+") {
			createTime, _ = time.ParseInLocation("2006-01-02T15:04:05+08:00", item.CreatedAt, nyc)
		} else {
			createTime, _ = time.ParseInLocation("2006-01-02T15:04:05", item.CreatedAt, nyc)
		}
		data = append(data, model.Volume{
			CreatedAt:  createTime,
			Name:       item.Name,
			Driver:     item.Driver,
			Mountpoint: item.Mountpoint,
			Labels:     labels,
		})
	}

	result.Total = int64(total)
	result.Items = data
	return &result, nil
}

func (c DockerClient) VolumeList() (*model.PageResult, error) {
	var result model.PageResult
	list, err := c.cli.VolumeList(context.TODO(), volume.ListOptions{})
	if err != nil {
		return &result, err
	}
	var datas []model.Options
	for _, item := range list.Volumes {
		datas = append(datas, model.Options{
			Option: item.Name,
		})
	}
	sort.Slice(datas, func(i, j int) bool {
		return datas[i].Option < datas[j].Option
	})
	result.Total = int64(len(datas))
	result.Items = datas
	return &result, nil
}

func (c DockerClient) VolumeDelete(req model.BatchDelete) error {
	for _, id := range req.Names {
		if err := c.cli.VolumeRemove(context.TODO(), id, true); err != nil {
			if strings.Contains(err.Error(), "volume is in use") {
				return errors.New(constant.ErrInUsed)
			}
			return err
		}
	}
	return nil
}

func (c DockerClient) VolumeCreate(req model.VolumeCreate) error {
	arg := filters.NewArgs()
	arg.Add("name", req.Name)
	vos, _ := c.cli.VolumeList(context.TODO(), volume.ListOptions{Filters: arg})
	if len(vos.Volumes) != 0 {
		for _, v := range vos.Volumes {
			if v.Name == req.Name {
				return constant.ErrRecordExist
			}
		}
	}

	labelsMap := make(map[string]string)
	for _, kv := range req.Labels {
		labelsMap[kv.Key] = kv.Value
	}
	options := volume.CreateOptions{
		Name:       req.Name,
		Driver:     req.Driver,
		DriverOpts: common.StringsToMap(req.Options),
		Labels:     labelsMap,
	}
	if _, err := c.cli.VolumeCreate(context.TODO(), options); err != nil {
		return err
	}
	return nil
}
