package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/proto"
	"github.com/sensdata/idb/core/model"
	"google.golang.org/grpc"
)

type Pma interface {
	GetComposes(hostID uint64, req model.GetComposesRequest) (*model.GetComposesResponse, error)
	Operation(hostID uint64, req model.OperateRequest) error
	SetPort(hostID uint64, req model.SetPortRequest) error
	GetServers(hostID uint64, req model.GetServersRequest) (*model.GetServersResponse, error)
	AddServer(hostID uint64, req model.AddServerRequest) error
	UpdateServer(hostID uint64, req model.AddServerRequest) error
	RemoveServer(hostID uint64, req model.RemoveServerRequest) error
}

type PmaPlugin struct {
	plugin.Plugin
	Impl Pma
}

func (p *PmaPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterPmaServer(s, &PmaGRPCServer{Impl: p.Impl})
	return nil
}

func (p *PmaPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &PmaGRPCClient{client: proto.NewPmaClient(c)}, nil
}

var _ plugin.GRPCPlugin = &PmaPlugin{}

type PmaGRPCClient struct {
	client proto.PmaClient
}

var _ Pma = (*PmaGRPCClient)(nil)

func (c *PmaGRPCClient) GetComposes(hostID uint64, req model.GetComposesRequest) (*model.GetComposesResponse, error) {
	resp, err := c.client.PmaGetComposes(context.Background(), &proto.PmaGetComposesRequest{
		HostId:   uint32(hostID),
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	var result model.GetComposesResponse
	for _, item := range resp.Items {
		result.Composes = append(result.Composes, &model.ComposeBrief{
			Name:    item.Name,
			Version: item.Version,
			Port:    item.Port,
			Status:  item.Status,
		})
	}
	result.Total = int(resp.Total)
	return &result, nil
}

func (c *PmaGRPCClient) Operation(hostID uint64, req model.OperateRequest) error {
	_, err := c.client.PmaOperation(context.Background(), &proto.PmaOperationRequest{
		HostId:    uint32(hostID),
		Name:      req.Name,
		Operation: req.Operation,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *PmaGRPCClient) SetPort(hostID uint64, req model.SetPortRequest) error {
	_, err := c.client.PmaSetPort(context.Background(), &proto.PmaSetPortRequest{
		HostId: uint32(hostID),
		Name:   req.Name,
		Port:   req.Port,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *PmaGRPCClient) GetServers(hostID uint64, req model.GetServersRequest) (*model.GetServersResponse, error) {
	resp, err := c.client.PmaGetServers(context.Background(), &proto.PmaGetServersRequest{
		HostId:   uint32(hostID),
		Name:     req.Name,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	var result model.GetServersResponse
	for _, item := range resp.Items {
		result.Servers = append(result.Servers, &model.PmaServerInfo{
			Verbose: item.Verbose,
			Host:    item.Host,
			Port:    item.Port,
		})
	}
	result.Total = int(resp.Total)
	return &result, nil
}

func (c *PmaGRPCClient) AddServer(hostID uint64, req model.AddServerRequest) error {
	_, err := c.client.PmaAddServer(context.Background(), &proto.PmaAddServerRequest{
		HostId:  uint32(hostID),
		Name:    req.Name,
		Verbose: req.Verbose,
		Host:    req.Host,
		Port:    req.Port,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *PmaGRPCClient) UpdateServer(hostID uint64, req model.AddServerRequest) error {
	_, err := c.client.PmaUpdateServer(context.Background(), &proto.PmaAddServerRequest{
		HostId:  uint32(hostID),
		Name:    req.Name,
		Verbose: req.Verbose,
		Host:    req.Host,
		Port:    req.Port,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *PmaGRPCClient) RemoveServer(hostID uint64, req model.RemoveServerRequest) error {
	_, err := c.client.PmaRemoveServer(context.Background(), &proto.PmaRemoveServerRequest{
		HostId: uint32(hostID),
		Name:   req.Name,
		Host:   req.Host,
		Port:   req.Port,
	})
	if err != nil {
		return err
	}
	return nil
}

type PmaGRPCServer struct {
	Impl Pma
	*proto.UnimplementedPmaServer
}

func (s *PmaGRPCServer) PmaGetComposes(ctx context.Context, req *proto.PmaGetComposesRequest) (*proto.PmaGetComposesResponse, error) {
	var resp proto.PmaGetComposesResponse

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
		resp.Items = append(resp.Items, &proto.PmaComposesInfo{
			Name:    item.Name,
			Version: item.Version,
			Port:    item.Port,
			Status:  item.Status,
		})
	}
	resp.Total = int64(result.Total)
	return &resp, nil
}

func (s *PmaGRPCServer) PmaOperation(ctx context.Context, req *proto.PmaOperationRequest) (*proto.PmaCommonResponse, error) {
	var resp proto.PmaCommonResponse

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

func (s *PmaGRPCServer) PmaSetPort(ctx context.Context, req *proto.PmaSetPortRequest) (*proto.PmaCommonResponse, error) {
	var resp proto.PmaCommonResponse

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

func (s *PmaGRPCServer) PmaGetServers(ctx context.Context, req *proto.PmaGetServersRequest) (*proto.PmaGetServersResponse, error) {
	var resp proto.PmaGetServersResponse

	result, err := s.Impl.GetServers(
		uint64(req.HostId),
		model.GetServersRequest{
			Name:     req.Name,
			Page:     int(req.Page),
			PageSize: int(req.PageSize),
		},
	)
	if err != nil {
		return &resp, err
	}

	for _, item := range result.Servers {
		resp.Items = append(resp.Items, &proto.PmaServerInfo{
			Verbose: item.Verbose,
			Host:    item.Host,
			Port:    item.Port,
		})
	}
	resp.Total = int64(result.Total)
	return &resp, nil
}

func (s *PmaGRPCServer) PmaAddServer(ctx context.Context, req *proto.PmaAddServerRequest) (*proto.PmaCommonResponse, error) {
	var resp proto.PmaCommonResponse

	err := s.Impl.AddServer(
		uint64(req.HostId),
		model.AddServerRequest{
			Name:    req.Name,
			Verbose: req.Verbose,
			Host:    req.Host,
			Port:    req.Port,
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}

func (s *PmaGRPCServer) PmaUpdateServer(ctx context.Context, req *proto.PmaAddServerRequest) (*proto.PmaCommonResponse, error) {
	var resp proto.PmaCommonResponse

	err := s.Impl.UpdateServer(
		uint64(req.HostId),
		model.AddServerRequest{
			Name:    req.Name,
			Verbose: req.Verbose,
			Host:    req.Host,
			Port:    req.Port,
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}

func (s *PmaGRPCServer) PmaRemoveServer(ctx context.Context, req *proto.PmaRemoveServerRequest) (*proto.PmaCommonResponse, error) {
	var resp proto.PmaCommonResponse

	err := s.Impl.RemoveServer(
		uint64(req.HostId),
		model.RemoveServerRequest{
			Name: req.Name,
			Host: req.Host,
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
