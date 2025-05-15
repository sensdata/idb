package client

import (
	"context"
	"sort"
	"strings"

	"github.com/docker/docker/api/types/network"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils/common"
)

func (c DockerClient) NetworkPage(req model.SearchPageInfo) (*model.PageResult, error) {
	var result model.PageResult
	list, err := c.cli.NetworkList(context.TODO(), network.ListOptions{})
	if err != nil {
		return &result, err
	}
	if len(req.Info) != 0 {
		length, count := len(list), 0
		for count < length {
			if !strings.Contains(list[count].Name, req.Info) {
				list = append(list[:count], list[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	var (
		data    []model.Network
		records []network.Inspect
	)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Created.Before(list[j].Created)
	})
	total, start, end := len(list), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]network.Inspect, 0)
	} else {
		if end >= total {
			end = total
		}
		records = list[start:end]
	}

	for _, item := range records {
		var labels []model.KeyValue
		for key, val := range item.Labels {
			labels = append(labels, model.KeyValue{Key: key, Value: val})
		}
		var ipam network.IPAMConfig
		if len(item.IPAM.Config) > 0 {
			ipam = item.IPAM.Config[0]
		}
		data = append(data, model.Network{
			ID:         item.ID,
			CreatedAt:  item.Created,
			Name:       item.Name,
			Driver:     item.Driver,
			IPAMDriver: item.IPAM.Driver,
			Subnet:     ipam.Subnet,
			Gateway:    ipam.Gateway,
			Attachable: item.Attachable,
			Labels:     labels,
		})
	}

	result.Total = int64(total)
	result.Items = data
	return &result, nil
}

func (c DockerClient) NetworkList() (*model.PageResult, error) {
	var result model.PageResult
	list, err := c.cli.NetworkList(context.TODO(), network.ListOptions{})
	if err != nil {
		return &result, err
	}
	var datas []model.Options
	for _, item := range list {
		datas = append(datas, model.Options{Option: item.Name})
	}
	sort.Slice(datas, func(i, j int) bool {
		return datas[i].Option < datas[j].Option
	})
	result.Total = int64(len(datas))
	result.Items = datas
	return &result, nil
}

func (c DockerClient) NetworkDelete(req model.BatchDelete) error {
	for _, id := range req.Names {
		if err := c.cli.NetworkRemove(context.TODO(), id); err != nil {
			if strings.Contains(err.Error(), "has active endpoints") {
				return errors.New(constant.ErrInUsed)
			}
			return err
		}
	}
	return nil
}
func (c DockerClient) NetworkCreate(req model.NetworkCreate) error {
	var (
		ipams    []network.IPAMConfig
		enableV6 bool
	)
	if req.Ipv4 {
		var itemIpam network.IPAMConfig
		if len(req.AuxAddress) != 0 {
			itemIpam.AuxAddress = make(map[string]string)
		}
		if len(req.Subnet) != 0 {
			itemIpam.Subnet = req.Subnet
		}
		if len(req.Gateway) != 0 {
			itemIpam.Gateway = req.Gateway
		}
		if len(req.IPRange) != 0 {
			itemIpam.IPRange = req.IPRange
		}
		for _, addr := range req.AuxAddress {
			itemIpam.AuxAddress[addr.Key] = addr.Value
		}
		ipams = append(ipams, itemIpam)
	}
	if req.Ipv6 {
		enableV6 = true
		var itemIpam network.IPAMConfig
		if len(req.AuxAddress) != 0 {
			itemIpam.AuxAddress = make(map[string]string)
		}
		if len(req.SubnetV6) != 0 {
			itemIpam.Subnet = req.SubnetV6
		}
		if len(req.GatewayV6) != 0 {
			itemIpam.Gateway = req.GatewayV6
		}
		if len(req.IPRangeV6) != 0 {
			itemIpam.IPRange = req.IPRangeV6
		}
		for _, addr := range req.AuxAddressV6 {
			itemIpam.AuxAddress[addr.Key] = addr.Value
		}
		ipams = append(ipams, itemIpam)
	}

	labelsMap := make(map[string]string)
	for _, kv := range req.Labels {
		labelsMap[kv.Key] = kv.Value
	}

	options := network.CreateOptions{
		EnableIPv6: &enableV6,
		Driver:     req.Driver,
		Options:    common.StringsToMap(req.Options),
		Labels:     labelsMap,
	}
	if len(ipams) != 0 {
		options.IPAM = &network.IPAM{Config: ipams}
	}
	if _, err := c.cli.NetworkCreate(context.TODO(), req.Name, options); err != nil {
		return err
	}
	return nil
}
