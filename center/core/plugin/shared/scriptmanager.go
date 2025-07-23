package shared

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/proto"
	"github.com/sensdata/idb/core/model"
	"google.golang.org/grpc"
)

type ScriptManager interface {
	ListCategories(hostID uint64, req model.QueryGitFile) (*model.ScriptList, error)
	ListScripts(hostID uint64, req model.QueryGitFile) (*model.ScriptList, error)
}

type ScriptMangerPlugin struct {
	plugin.Plugin
	Impl ScriptManager
}

func (p *ScriptMangerPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterScriptManagerServer(s, &ScriptManagerGRPCServer{Impl: p.Impl})
	return nil
}

func (p *ScriptMangerPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &ScriptManagerGRPCClient{client: proto.NewScriptManagerClient(c)}, nil
}

var _ plugin.GRPCPlugin = &ScriptMangerPlugin{}

type ScriptManagerGRPCClient struct {
	client proto.ScriptManagerClient
}

var _ ScriptManager = (*ScriptManagerGRPCClient)(nil)

func (c *ScriptManagerGRPCClient) ListCategories(hostID uint64, req model.QueryGitFile) (*model.ScriptList, error) {
	resp, err := c.client.ListCategories(context.Background(), &proto.ListScriptsRequest{
		HostId:   uint32(hostID),
		Type:     req.Type,
		Category: req.Category,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	var result model.ScriptList
	result.Total = resp.Total
	scripts := make([]*model.ScriptInfo, 0)
	for _, item := range resp.Items {
		scripts = append(scripts, &model.ScriptInfo{
			Source:    item.Source,
			Name:      item.Name,
			Extension: item.Extension,
			Content:   item.Content,
			Size:      item.Size,
			ModTime:   item.ModTime,
			Linked:    item.Linked,
		})
	}
	result.Items = scripts
	return &result, nil
}

func (c *ScriptManagerGRPCClient) ListScripts(hostID uint64, req model.QueryGitFile) (*model.ScriptList, error) {
	resp, err := c.client.ListScripts(context.Background(), &proto.ListScriptsRequest{
		HostId:   uint32(hostID),
		Type:     req.Type,
		Category: req.Category,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	var result model.ScriptList
	result.Total = resp.Total
	scripts := make([]*model.ScriptInfo, 0)
	for _, item := range resp.Items {
		scripts = append(scripts, &model.ScriptInfo{
			Source:    item.Source,
			Name:      item.Name,
			Extension: item.Extension,
			Content:   item.Content,
			Size:      item.Size,
			ModTime:   item.ModTime,
			Linked:    item.Linked,
		})
	}
	result.Items = scripts
	return &result, nil
}

type ScriptManagerGRPCServer struct {
	Impl ScriptManager
	*proto.UnimplementedScriptManagerServer
}

func (s *ScriptManagerGRPCServer) ListCategories(ctx context.Context, req *proto.ListScriptsRequest) (*proto.ListScriptsResponse, error) {
	var resp proto.ListScriptsResponse

	result, err := s.Impl.ListCategories(
		uint64(req.HostId),
		model.QueryGitFile{
			Type:     req.Type,
			Category: req.Category,
			Page:     int(req.Page),
			PageSize: int(req.PageSize),
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Total = result.Total
	for _, item := range result.Items {
		resp.Items = append(resp.Items, &proto.ScriptInfo{
			Source:    item.Source,
			Name:      item.Name,
			Extension: item.Extension,
			Content:   item.Content,
			Size:      item.Size,
			ModTime:   item.ModTime,
			Linked:    item.Linked,
		})
	}
	return &resp, nil
}

func (s *ScriptManagerGRPCServer) ListScripts(ctx context.Context, req *proto.ListScriptsRequest) (*proto.ListScriptsResponse, error) {
	var resp proto.ListScriptsResponse

	result, err := s.Impl.ListScripts(
		uint64(req.HostId),
		model.QueryGitFile{
			Type:     req.Type,
			Category: req.Category,
			Page:     int(req.Page),
			PageSize: int(req.PageSize),
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Total = result.Total
	for _, item := range result.Items {
		resp.Items = append(resp.Items, &proto.ScriptInfo{
			Source:    item.Source,
			Name:      item.Name,
			Extension: item.Extension,
			Content:   item.Content,
			Size:      item.Size,
			ModTime:   item.ModTime,
			Linked:    item.Linked,
		})
	}
	return &resp, nil
}
