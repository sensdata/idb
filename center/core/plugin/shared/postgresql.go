package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/proto"
	"github.com/sensdata/idb/core/model"
	"google.golang.org/grpc"
)

type PostgreSql interface {
	GetComposes(hostID uint64, req model.GetComposesRequest) (*model.GetComposesResponse, error)
	Operation(hostID uint64, req model.OperateRequest) error
	SetPort(hostID uint64, req model.SetPortRequest) error
	GetConf(hostID uint64, req model.GetConfRequest) (*model.GetConfResponse, error)
	SetConf(hostID uint64, req model.SetConfRequest) error
}

type PostgreSqlPlugin struct {
	plugin.Plugin
	Impl PostgreSql
}

func (p *PostgreSqlPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterPostgreSqlServer(s, &PostgreSqlGRPCServer{Impl: p.Impl})
	return nil
}

func (p *PostgreSqlPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &PostgreSqlGRPCClient{client: proto.NewPostgreSqlClient(c)}, nil
}

var _ plugin.GRPCPlugin = &PostgreSqlPlugin{}

type PostgreSqlGRPCClient struct {
	client proto.PostgreSqlClient
}

var _ PostgreSql = (*PostgreSqlGRPCClient)(nil)

func (c *PostgreSqlGRPCClient) GetComposes(hostID uint64, req model.GetComposesRequest) (*model.GetComposesResponse, error) {
	resp, err := c.client.PGGetComposes(context.Background(), &proto.PGGetComposesRequest{
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

func (c *PostgreSqlGRPCClient) Operation(hostID uint64, req model.OperateRequest) error {
	_, err := c.client.PGOperation(context.Background(), &proto.PGOperationRequest{
		HostId:    uint32(hostID),
		Name:      req.Name,
		Operation: req.Operation,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *PostgreSqlGRPCClient) SetPort(hostID uint64, req model.SetPortRequest) error {
	_, err := c.client.PGSetPort(context.Background(), &proto.PGSetPortRequest{
		HostId: uint32(hostID),
		Name:   req.Name,
		Port:   req.Port,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *PostgreSqlGRPCClient) GetConf(hostID uint64, req model.GetConfRequest) (*model.GetConfResponse, error) {
	resp, err := c.client.PGGetConf(context.Background(), &proto.PGGetConfRequest{
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

func (c *PostgreSqlGRPCClient) SetConf(hostID uint64, req model.SetConfRequest) error {
	_, err := c.client.PGSetConf(context.Background(), &proto.PGSetConfRequest{
		HostId:  uint32(hostID),
		Name:    req.Name,
		Content: req.Content,
	})
	if err != nil {
		return err
	}
	return nil
}

type PostgreSqlGRPCServer struct {
	Impl PostgreSql
	*proto.UnimplementedPostgreSqlServer
}

func (s *PostgreSqlGRPCServer) PGGetComposes(ctx context.Context, req *proto.PGGetComposesRequest) (*proto.PGGetComposesResponse, error) {
	var resp proto.PGGetComposesResponse

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
		resp.Items = append(resp.Items, &proto.PGComposesInfo{
			Name:    item.Name,
			Version: item.Version,
			Port:    item.Port,
			Status:  item.Status,
		})
	}
	resp.Total = int64(result.Total)
	return &resp, nil
}

func (s *PostgreSqlGRPCServer) PGOperation(ctx context.Context, req *proto.PGOperationRequest) (*proto.PGCommonResponse, error) {
	var resp proto.PGCommonResponse

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

func (s *PostgreSqlGRPCServer) PGSetPort(ctx context.Context, req *proto.PGSetPortRequest) (*proto.PGCommonResponse, error) {
	var resp proto.PGCommonResponse

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

func (s *PostgreSqlGRPCServer) PGGetConf(ctx context.Context, req *proto.PGGetConfRequest) (*proto.PGGetConfResponse, error) {
	result, err := s.Impl.GetConf(
		uint64(req.HostId),
		model.GetConfRequest{
			Name: req.Name,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.PGGetConfResponse{
		Content: result.Content,
	}, nil
}

func (s *PostgreSqlGRPCServer) PGSetConf(ctx context.Context, req *proto.PGSetConfRequest) (*proto.PGCommonResponse, error) {
	var resp proto.PGCommonResponse

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
