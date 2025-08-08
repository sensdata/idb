package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/proto"
	"github.com/sensdata/idb/core/model"
	"google.golang.org/grpc"
)

type Auth interface {
	RegisterFingerprint(req model.RegisterFingerprintReq) (*model.RegisterFingerprintRsp, error)
	VerifyLicense(req model.VerifyLicenseRequest) (*model.VerifyLicenseResponse, error)
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

func (c *AuthGRPCClient) RegisterFingerprint(req model.RegisterFingerprintReq) (*model.RegisterFingerprintRsp, error) {
	resp, err := c.client.RegisterFingerprint(context.Background(), &proto.RegisterFingerprintRequest{
		Ip:          req.Ip,
		Mac:         req.Mac,
		Fingerprint: req.Fingerprint,
	})
	if err != nil {
		return nil, err
	}
	return &model.RegisterFingerprintRsp{
		License:   resp.License,
		Signature: resp.Signature,
	}, nil
}

func (c *AuthGRPCClient) VerifyLicense(req model.VerifyLicenseRequest) (*model.VerifyLicenseResponse, error) {
	resp, err := c.client.VerifyLicense(context.Background(), &proto.VerifyLicenseRequest{
		License:   req.License,
		Signature: req.Signature,
	})
	if err != nil {
		return nil, err
	}
	return &model.VerifyLicenseResponse{
		Valid:       resp.Valid,
		LicenseType: resp.LicenseType,
		ExpireAt:    resp.ExpireAt,
	}, nil
}

type AuthGRPCServer struct {
	Impl Auth
	*proto.UnimplementedAuthServer
}

func (s *AuthGRPCServer) RegisterFingerprint(ctx context.Context, req *proto.RegisterFingerprintRequest) (*proto.RegisterFingerprintResponse, error) {
	resp, err := s.Impl.RegisterFingerprint(model.RegisterFingerprintReq{
		Ip:          req.Ip,
		Mac:         req.Mac,
		Fingerprint: req.Fingerprint,
	})
	if err != nil {
		return nil, err
	}
	return &proto.RegisterFingerprintResponse{
		License:   resp.License,
		Signature: resp.Signature,
	}, nil
}

func (s *AuthGRPCServer) VerifyLicense(ctx context.Context, req *proto.VerifyLicenseRequest) (*proto.VerifyLicenseResponse, error) {
	resp, err := s.Impl.VerifyLicense(model.VerifyLicenseRequest{
		License:   req.License,
		Signature: req.Signature,
	})
	if err != nil {
		return nil, err
	}
	return &proto.VerifyLicenseResponse{
		Valid:       resp.Valid,
		LicenseType: resp.LicenseType,
		ExpireAt:    resp.ExpireAt,
	}, nil
}
