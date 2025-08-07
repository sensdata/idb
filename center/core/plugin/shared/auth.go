package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/proto"
	"github.com/sensdata/idb/core/model"
	"google.golang.org/grpc"
)

type Auth interface {
	VerifyFingerprint(req model.VerifyRequest) (*model.VerifyResponse, error)
}

type AuthPlugin struct {
	plugin.Plugin
	Impl Auth
}

func (p *AuthPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterAuthServer(s, &AuthGRPCServer{Impl: p.Impl})
	return nil
}

func (p *AuthPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &AuthGRPCClient{client: proto.NewAuthClient(c)}, nil
}

var _ plugin.GRPCPlugin = &AuthPlugin{}

type AuthGRPCClient struct {
	client proto.AuthClient
}

var _ Auth = (*AuthGRPCClient)(nil)

func (c *AuthGRPCClient) VerifyFingerprint(req model.VerifyRequest) (*model.VerifyResponse, error) {
	resp, err := c.client.VerifyFingerprint(context.Background(), &proto.VerifyFingerprintRequest{
		Fingerprint: req.Fingerprint,
		Ip:          req.IP,
		Mac:         req.MAC,
	})
	if err != nil {
		return nil, err
	}
	return &model.VerifyResponse{
		Fingerprint: resp.Fingerprint,
		Result:      resp.Result,
		VerifyTime:  resp.VerifyTime,
		ExpireTime:  resp.ExpireTime,
	}, nil
}

type AuthGRPCServer struct {
	Impl Auth
	*proto.UnimplementedAuthServer
}

func (s *AuthGRPCServer) VerifyFingerprint(ctx context.Context, req *proto.VerifyFingerprintRequest) (*proto.VerifyFingerprintResponse, error) {
	resp, err := s.Impl.VerifyFingerprint(model.VerifyRequest{
		Fingerprint: req.Fingerprint,
		IP:          req.Ip,
		MAC:         req.Mac,
	})
	if err != nil {
		return nil, err
	}
	return &proto.VerifyFingerprintResponse{
		Fingerprint: resp.Fingerprint,
		Result:      resp.Result,
		VerifyTime:  resp.VerifyTime,
		ExpireTime:  resp.ExpireTime,
	}, nil
}
