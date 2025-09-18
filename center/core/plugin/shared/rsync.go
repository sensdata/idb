package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/proto"
	"github.com/sensdata/idb/core/model"
	"google.golang.org/grpc"
)

type Rsync interface {
	ListTask(req *model.RsyncListTaskRequest) (*model.RsyncListTaskResponse, error)
	CreateTask(req *model.RsyncCreateTaskRequest) (*model.RsyncCreateTaskResponse, error)
	QueryTask(req *model.RsyncQueryTaskRequest) (*model.RsyncTaskInfo, error)
	CancelTask(req *model.RsyncCancelTaskRequest) error
	DeleteTask(req *model.RsyncDeleteTaskRequest) error
	RetryTask(req *model.RsyncRetryTaskRequest) error
}

type RsyncPlugin struct {
	plugin.Plugin
	Impl Rsync
}

func (p *RsyncPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterRsyncServer(s, &RsyncGRPCServer{Impl: p.Impl})
	return nil
}

func (p *RsyncPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &RsyncGRPCClient{client: proto.NewRsyncClient(c)}, nil
}

var _ plugin.GRPCPlugin = &RsyncPlugin{}

type RsyncGRPCClient struct {
	client proto.RsyncClient
}

var _ Rsync = (*RsyncGRPCClient)(nil)

func (c *RsyncGRPCClient) ListTask(req *model.RsyncListTaskRequest) (*model.RsyncListTaskResponse, error) {
	resp, err := c.client.RsyncListTask(context.Background(), &proto.RsyncListTaskRequest{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	var result model.RsyncListTaskResponse
	for _, item := range resp.Items {
		result.Tasks = append(result.Tasks, &model.RsyncTaskInfo{
			ID:        item.Id,
			Src:       item.Src,
			Dst:       item.Dst,
			CacheDir:  item.CacheDir,
			Mode:      item.Mode,
			Status:    item.Status,
			Progress:  int(item.Progress),
			Step:      item.Step,
			StartTime: item.StartTime,
			EndTime:   item.EndTime,
			Error:     item.Error,
			LastLog:   item.LastLog,
		})
	}
	result.Total = result.Total
	return &result, nil
}

func (c *RsyncGRPCClient) CreateTask(req *model.RsyncCreateTaskRequest) (*model.RsyncCreateTaskResponse, error) {
	resp, err := c.client.RsyncCreateTask(context.Background(), &proto.RsyncCreateTaskRequest{
		Src: req.Src,
		SrcHost: &proto.RsyncHost{
			Host:     req.SrcHost.Host,
			Port:     int32(req.SrcHost.Port),
			User:     req.SrcHost.User,
			KeyPath:  req.SrcHost.KeyPath,
			Password: req.SrcHost.Password,
		},
		Dst: req.Dst,
		DstHost: &proto.RsyncHost{
			Host:     req.DstHost.Host,
			Port:     int32(req.DstHost.Port),
			User:     req.DstHost.User,
			KeyPath:  req.DstHost.KeyPath,
			Password: req.DstHost.Password,
		},
		Mode: req.Mode,
	})
	if err != nil {
		return nil, err
	}
	return &model.RsyncCreateTaskResponse{
		ID: resp.Id,
	}, nil
}

func (c *RsyncGRPCClient) QueryTask(req *model.RsyncQueryTaskRequest) (*model.RsyncTaskInfo, error) {
	resp, err := c.client.RsyncQueryTask(context.Background(), &proto.RsyncQueryTaskRequest{
		Id: req.ID,
	})
	if err != nil {
		return nil, err
	}
	return &model.RsyncTaskInfo{
		ID:        resp.Id,
		Src:       resp.Src,
		Dst:       resp.Dst,
		CacheDir:  resp.CacheDir,
		Mode:      resp.Mode,
		Status:    resp.Status,
		Progress:  int(resp.Progress),
		Step:      resp.Step,
		StartTime: resp.StartTime,
		EndTime:   resp.EndTime,
		Error:     resp.Error,
		LastLog:   resp.LastLog,
	}, nil
}

func (c *RsyncGRPCClient) CancelTask(req *model.RsyncCancelTaskRequest) error {
	_, err := c.client.RsyncCancelTask(context.Background(), &proto.RsyncCancelTaskRequest{
		Id: req.ID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *RsyncGRPCClient) DeleteTask(req *model.RsyncDeleteTaskRequest) error {
	_, err := c.client.RsyncDeleteTask(context.Background(), &proto.RsyncDeleteTaskRequest{
		Id: req.ID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *RsyncGRPCClient) RetryTask(req *model.RsyncRetryTaskRequest) error {
	_, err := c.client.RsyncRetryTask(context.Background(), &proto.RsyncRetryTaskRequest{
		Id: req.ID,
	})
	if err != nil {
		return err
	}
	return nil
}

type RsyncGRPCServer struct {
	Impl Rsync
	*proto.UnimplementedRsyncServer
}

func (s *RsyncGRPCServer) RsyncListTask(ctx context.Context, req *proto.RsyncListTaskRequest) (*proto.RsyncListTaskResponse, error) {
	var resp proto.RsyncListTaskResponse

	result, err := s.Impl.ListTask(&model.RsyncListTaskRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return &resp, err
	}
	for _, item := range result.Tasks {
		resp.Items = append(resp.Items, &proto.RsyncTaskInfo{
			Id:        item.ID,
			Src:       item.Src,
			Dst:       item.Dst,
			CacheDir:  item.CacheDir,
			Mode:      item.Mode,
			Status:    item.Status,
			Progress:  int32(item.Progress),
			Step:      item.Step,
			StartTime: item.StartTime,
			EndTime:   item.EndTime,
			Error:     item.Error,
			LastLog:   item.LastLog,
		})
	}
	resp.Total = int64(result.Total)
	return &resp, nil
}

func (s *RsyncGRPCServer) RsyncCreateTask(ctx context.Context, req *proto.RsyncCreateTaskRequest) (*proto.RsyncCreateTaskResponse, error) {
	var resp proto.RsyncCreateTaskResponse

	result, err := s.Impl.CreateTask(&model.RsyncCreateTaskRequest{
		Src: req.Src,
		SrcHost: model.RsyncHost{
			Host:     req.SrcHost.Host,
			Port:     int(req.SrcHost.Port),
			User:     req.SrcHost.User,
			KeyPath:  req.SrcHost.KeyPath,
			Password: req.SrcHost.Password,
		},
		Dst: req.Dst,
		DstHost: model.RsyncHost{
			Host:     req.DstHost.Host,
			Port:     int(req.DstHost.Port),
			User:     req.DstHost.User,
			KeyPath:  req.DstHost.KeyPath,
			Password: req.DstHost.Password,
		},
		Mode: req.Mode,
	})
	if err != nil {
		return &resp, err
	}
	resp.Id = result.ID
	return &resp, nil
}

func (s *RsyncGRPCServer) RsyncQueryTask(ctx context.Context, req *proto.RsyncQueryTaskRequest) (*proto.RsyncTaskInfo, error) {
	var resp proto.RsyncTaskInfo

	result, err := s.Impl.QueryTask(&model.RsyncQueryTaskRequest{
		ID: req.Id,
	})
	if err != nil {
		return &resp, err
	}
	resp.Id = result.ID
	resp.Src = result.Src
	resp.Dst = result.Dst
	resp.CacheDir = result.CacheDir
	resp.Mode = result.Mode
	resp.Status = result.Status
	resp.Progress = int32(result.Progress)
	resp.Step = result.Step
	resp.StartTime = result.StartTime
	resp.EndTime = result.EndTime
	resp.Error = result.Error
	resp.LastLog = result.LastLog
	return &resp, nil
}

func (s *RsyncGRPCServer) RsyncCancelTask(ctx context.Context, req *proto.RsyncCancelTaskRequest) (*proto.RsyncCommonResponse, error) {
	var resp proto.RsyncCommonResponse
	err := s.Impl.CancelTask(&model.RsyncCancelTaskRequest{
		ID: req.Id,
	})
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}

func (s *RsyncGRPCServer) RsyncDeleteTask(ctx context.Context, req *proto.RsyncDeleteTaskRequest) (*proto.RsyncCommonResponse, error) {
	var resp proto.RsyncCommonResponse
	err := s.Impl.DeleteTask(&model.RsyncDeleteTaskRequest{
		ID: req.Id,
	})
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}

func (s *RsyncGRPCServer) RsyncRetryTask(ctx context.Context, req *proto.RsyncRetryTaskRequest) (*proto.RsyncCommonResponse, error) {
	var resp proto.RsyncCommonResponse
	err := s.Impl.RetryTask(&model.RsyncRetryTaskRequest{
		ID: req.Id,
	})
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}
