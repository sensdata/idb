package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/proto"
	"github.com/sensdata/idb/core/model"
	"google.golang.org/grpc"
)

type MysqlManager interface {
	GetComposes(hostID uint64, req model.GetComposesRequest) (*model.GetComposesResponse, error)
	Operation(hostID uint64, req model.OperateRequest) error
	SetPort(hostID uint64, req model.SetPortRequest) error
	GetConf(hostID uint64, req model.GetConfRequest) (*model.GetConfResponse, error)
	SetConf(hostID uint64, req model.SetConfRequest) error
}

type MysqlManagerPlugin struct {
	plugin.Plugin
	Impl MysqlManager
}

func (p *MysqlManagerPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterMysqlManagerServer(s, &MysqlManagerGRPCServer{Impl: p.Impl})
	return nil
}

func (p *MysqlManagerPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &MysqlManagerGRPCClient{client: proto.NewMysqlManagerClient(c)}, nil
}

var _ plugin.GRPCPlugin = &MysqlManagerPlugin{}

type MysqlManagerGRPCClient struct {
	client proto.MysqlManagerClient
}

var _ MysqlManager = (*MysqlManagerGRPCClient)(nil)

func (c *MysqlManagerGRPCClient) GetComposes(hostID uint64, req model.GetComposesRequest) (*model.GetComposesResponse, error) {
	resp, err := c.client.GetComposes(context.Background(), &proto.GetComposesRequest{
		HostId:   uint32(hostID),
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	var result model.GetComposesResponse
	for _, item := range resp.Items {
		result.Composes = append(result.Composes, &model.ComposesInfo{
			Name:    item.Name,
			Version: item.Version,
			Port:    item.Port,
		})
	}
	result.Total = int(resp.Total)
	return &result, nil
}

func (c *MysqlManagerGRPCClient) Operation(hostID uint64, req model.OperateRequest) error {
	_, err := c.client.Operation(context.Background(), &proto.OperationRequest{
		HostId:    uint32(hostID),
		Name:      req.Name,
		Operation: req.Operation,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *MysqlManagerGRPCClient) SetPort(hostID uint64, req model.SetPortRequest) error {
	_, err := c.client.SetPort(context.Background(), &proto.SetPortRequest{
		HostId: uint32(hostID),
		Name:   req.Name,
		Port:   req.Port,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *MysqlManagerGRPCClient) GetConf(hostID uint64, req model.GetConfRequest) (*model.GetConfResponse, error) {
	resp, err := c.client.GetConf(context.Background(), &proto.GetConfRequest{
		HostId: uint32(hostID),
		Name:   req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &model.GetConfResponse{
		Content: resp.Content,
	}, nil
}

func (c *MysqlManagerGRPCClient) SetConf(hostID uint64, req model.SetConfRequest) error {
	_, err := c.client.SetConf(context.Background(), &proto.SetConfRequest{
		HostId:  uint32(hostID),
		Name:    req.Name,
		Content: req.Content,
	})
	if err != nil {
		return err
	}
	return nil
}

type MysqlManagerGRPCServer struct {
	Impl MysqlManager
	*proto.UnimplementedMysqlManagerServer
}

func (s *MysqlManagerGRPCServer) GetComposes(ctx context.Context, req *proto.GetComposesRequest) (*proto.GetComposesResponse, error) {
	var resp proto.GetComposesResponse

	result, err := s.Impl.GetComposes(
		uint64(req.HostId),
		model.GetComposesRequest{
			Page:     int(req.Page),
			PageSize: int(req.PageSize),
		},
	)
	if err != nil {
		return &resp, err
	}
	for _, item := range result.Composes {
		resp.Items = append(resp.Items, &proto.ComposesInfo{
			Name:    item.Name,
			Version: item.Version,
			Port:    item.Port,
		})
	}
	resp.Total = int64(result.Total)
	return &resp, nil
}

func (s *MysqlManagerGRPCServer) Operation(ctx context.Context, req *proto.OperationRequest) (*proto.MysqlCommonResponse, error) {
	var resp proto.MysqlCommonResponse

	err := s.Impl.Operation(
		uint64(req.HostId),
		model.OperateRequest{
			Name:      req.Name,
			Operation: req.Operation,
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}

func (s *MysqlManagerGRPCServer) SetPort(ctx context.Context, req *proto.SetPortRequest) (*proto.MysqlCommonResponse, error) {
	var resp proto.MysqlCommonResponse

	err := s.Impl.SetPort(
		uint64(req.HostId),
		model.SetPortRequest{
			Name: req.Name,
			Port: req.Port,
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}

func (s *MysqlManagerGRPCServer) GetConf(ctx context.Context, req *proto.GetConfRequest) (*proto.GetConfResponse, error) {
	result, err := s.Impl.GetConf(
		uint64(req.HostId),
		model.GetConfRequest{
			Name: req.Name,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.GetConfResponse{
		Content: result.Content,
	}, nil
}

func (s *MysqlManagerGRPCServer) SetConf(ctx context.Context, req *proto.SetConfRequest) (*proto.MysqlCommonResponse, error) {
	var resp proto.MysqlCommonResponse

	err := s.Impl.SetConf(
		uint64(req.HostId),
		model.SetConfRequest{
			Name:    req.Name,
			Content: req.Content,
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}
