package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/proto"
	"github.com/sensdata/idb/core/model"
	"google.golang.org/grpc"
)

type Redis interface {
	GetComposes(hostID uint64, req model.GetComposesRequest) (*model.GetComposesResponse, error)
	Operation(hostID uint64, req model.OperateRequest) error
	SetPort(hostID uint64, req model.SetPortRequest) error
	GetConf(hostID uint64, req model.GetConfRequest) (*model.GetConfResponse, error)
	SetConf(hostID uint64, req model.SetConfRequest) error
	GetRemoteAccess(hostID uint64, req model.GetRemoteAccessRequest) (*model.GetRemoteAccessResponse, error)
	SetRemoteAccess(hostID uint64, req model.SetRemoteAccessRequest) error
	GetRootPassword(hostID uint64, req model.GetRootPasswordRequest) (*model.GetRootPasswordResponse, error)
	SetRootPassword(hostID uint64, req model.SetRootPasswordRequest) error
}

type RedisPlugin struct {
	plugin.Plugin
	Impl Redis
}

func (p *RedisPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterRedisServer(s, &RedisGRPCServer{Impl: p.Impl})
	return nil
}

func (p *RedisPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &RedisGRPCClient{client: proto.NewRedisClient(c)}, nil
}

var _ plugin.GRPCPlugin = &RedisPlugin{}

type RedisGRPCClient struct {
	client proto.RedisClient
}

var _ Redis = (*RedisGRPCClient)(nil)

func (c *RedisGRPCClient) GetComposes(hostID uint64, req model.GetComposesRequest) (*model.GetComposesResponse, error) {
	resp, err := c.client.RedisGetComposes(context.Background(), &proto.RedisGetComposesRequest{
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

func (c *RedisGRPCClient) Operation(hostID uint64, req model.OperateRequest) error {
	_, err := c.client.RedisOperation(context.Background(), &proto.RedisOperationRequest{
		HostId:    uint32(hostID),
		Name:      req.Name,
		Operation: req.Operation,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisGRPCClient) SetPort(hostID uint64, req model.SetPortRequest) error {
	_, err := c.client.RedisSetPort(context.Background(), &proto.RedisSetPortRequest{
		HostId: uint32(hostID),
		Name:   req.Name,
		Port:   req.Port,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisGRPCClient) GetConf(hostID uint64, req model.GetConfRequest) (*model.GetConfResponse, error) {
	resp, err := c.client.RedisGetConf(context.Background(), &proto.RedisGetConfRequest{
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

func (c *RedisGRPCClient) SetConf(hostID uint64, req model.SetConfRequest) error {
	_, err := c.client.RedisSetConf(context.Background(), &proto.RedisSetConfRequest{
		HostId:  uint32(hostID),
		Name:    req.Name,
		Content: req.Content,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisGRPCClient) GetRemoteAccess(hostID uint64, req model.GetRemoteAccessRequest) (*model.GetRemoteAccessResponse, error) {
	resp, err := c.client.RedisGetRemoteAccess(context.Background(), &proto.RedisGetRemoteAccessRequest{
		HostId: uint32(hostID),
		Name:   req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &model.GetRemoteAccessResponse{
		RemoteAccess: resp.RemoteAccess,
	}, nil
}

func (c *RedisGRPCClient) SetRemoteAccess(hostID uint64, req model.SetRemoteAccessRequest) error {
	_, err := c.client.RedisSetRemoteAccess(context.Background(), &proto.RedisSetRemoteAccessRequest{
		HostId:       uint32(hostID),
		Name:         req.Name,
		RemoteAccess: req.RemoteAccess,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisGRPCClient) GetRootPassword(hostID uint64, req model.GetRootPasswordRequest) (*model.GetRootPasswordResponse, error) {
	resp, err := c.client.RedisGetRootPassword(context.Background(), &proto.RedisGetRootPasswordRequest{
		HostId: uint32(hostID),
		Name:   req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &model.GetRootPasswordResponse{
		Password: resp.Pass,
	}, nil
}

func (c *RedisGRPCClient) SetRootPassword(hostID uint64, req model.SetRootPasswordRequest) error {
	_, err := c.client.RedisSetRootPassword(context.Background(), &proto.RedisSetRootPasswordRequest{
		HostId:  uint32(hostID),
		Name:    req.Name,
		NewPass: req.NewPass,
	})
	if err != nil {
		return err
	}
	return nil
}

type RedisGRPCServer struct {
	Impl Redis
	*proto.UnimplementedRedisServer
}

func (s *RedisGRPCServer) RedisGetComposes(ctx context.Context, req *proto.RedisGetComposesRequest) (*proto.RedisGetComposesResponse, error) {
	var resp proto.RedisGetComposesResponse

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
		resp.Items = append(resp.Items, &proto.RedisComposesInfo{
			Name:    item.Name,
			Version: item.Version,
			Port:    item.Port,
			Status:  item.Status,
		})
	}
	resp.Total = int64(result.Total)
	return &resp, nil
}

func (s *RedisGRPCServer) RedisOperation(ctx context.Context, req *proto.RedisOperationRequest) (*proto.RedisCommonResponse, error) {
	var resp proto.RedisCommonResponse

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

func (s *RedisGRPCServer) RedisSetPort(ctx context.Context, req *proto.RedisSetPortRequest) (*proto.RedisCommonResponse, error) {
	var resp proto.RedisCommonResponse

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

func (s *RedisGRPCServer) RedisGetConf(ctx context.Context, req *proto.RedisGetConfRequest) (*proto.RedisGetConfResponse, error) {
	result, err := s.Impl.GetConf(
		uint64(req.HostId),
		model.GetConfRequest{
			Name: req.Name,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.RedisGetConfResponse{
		Content: result.Content,
	}, nil
}

func (s *RedisGRPCServer) RedisSetConf(ctx context.Context, req *proto.RedisSetConfRequest) (*proto.RedisCommonResponse, error) {
	var resp proto.RedisCommonResponse

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

func (s *RedisGRPCServer) RedisGetRemoteAccess(ctx context.Context, req *proto.RedisGetRemoteAccessRequest) (*proto.RedisGetRemoteAccessResponse, error) {
	result, err := s.Impl.GetRemoteAccess(
		uint64(req.HostId),
		model.GetRemoteAccessRequest{
			Name: req.Name,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.RedisGetRemoteAccessResponse{
		RemoteAccess: result.RemoteAccess,
	}, nil
}

func (s *RedisGRPCServer) RedisSetRemoteAccess(ctx context.Context, req *proto.RedisSetRemoteAccessRequest) (*proto.RedisCommonResponse, error) {
	var resp proto.RedisCommonResponse

	err := s.Impl.SetRemoteAccess(
		uint64(req.HostId),
		model.SetRemoteAccessRequest{
			Name:         req.Name,
			RemoteAccess: req.RemoteAccess,
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}

func (s *RedisGRPCServer) RedisGetRootPassword(ctx context.Context, req *proto.RedisGetRootPasswordRequest) (*proto.RedisGetRootPasswordResponse, error) {
	result, err := s.Impl.GetRootPassword(
		uint64(req.HostId),
		model.GetRootPasswordRequest{
			Name: req.Name,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.RedisGetRootPasswordResponse{
		Pass: result.Password,
	}, nil
}

func (s *RedisGRPCServer) RedisSetRootPassword(ctx context.Context, req *proto.RedisSetRootPasswordRequest) (*proto.RedisCommonResponse, error) {
	var resp proto.RedisCommonResponse

	err := s.Impl.SetRootPassword(
		uint64(req.HostId),
		model.SetRootPasswordRequest{
			Name:    req.Name,
			NewPass: req.NewPass,
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}
