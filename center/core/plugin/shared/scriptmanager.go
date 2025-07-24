package shared

import (
	"context"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/proto"
	"github.com/sensdata/idb/core/model"
	"google.golang.org/grpc"
)

type ScriptManager interface {
	ListCategories(hostID uint64, req model.QueryGitFile) (*model.ScriptList, error)
	CreateCategory(hostID uint64, req model.CreateGitCategory) error
	UpdateCategory(hostID uint64, req model.UpdateGitCategory) error
	DeleteCategory(hostID uint64, req model.DeleteGitCategory) error
	ListScripts(hostID uint64, req model.QueryGitFile) (*model.ScriptList, error)
	GetScriptDetail(hostID uint64, req model.GetGitFileDetail) (*model.ScriptInfo, error)
	CreateScript(hostID uint64, req model.CreateGitFile) error
	UpdateScript(hostID uint64, req model.UpdateGitFile) error
	DeleteScript(hostID uint64, req model.DeleteGitFile) error
	RestoreScript(hostID uint64, req model.RestoreGitFile) error
	GetScriptHistory(hostID uint64, req model.GitFileLog) (*model.ScriptHistoryList, error)
	GetScriptDiff(hostID uint64, req model.GitFileDiff) (string, error)
	ScriptSync(hostID uint) error
	ScriptExec(hostID uint, req model.ExecuteScript) (*model.ScriptResult, error)
	GetScriptRunLogs(hostID uint, scriptPath string, page int, pageSize int) (*model.RunLogList, error)
	GetScriptRunLogDetail(hostID uint, logPath string) (*model.RunLogDetail, error)
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
	for _, item := range resp.Items {
		result.Items = append(result.Items, &model.ScriptInfo{
			Source:    item.Source,
			Name:      item.Name,
			Extension: item.Extension,
			Content:   item.Content,
			Size:      item.Size,
			ModTime:   item.ModTime,
			Linked:    item.Linked,
		})
	}
	result.Total = resp.Total
	return &result, nil
}

func (c *ScriptManagerGRPCClient) CreateCategory(hostID uint64, req model.CreateGitCategory) error {
	_, err := c.client.CreateCategory(context.Background(), &proto.CreateScriptCategoryRequest{
		HostId:   uint32(hostID),
		Type:     req.Type,
		Category: req.Category,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *ScriptManagerGRPCClient) UpdateCategory(hostID uint64, req model.UpdateGitCategory) error {
	_, err := c.client.UpdateCategory(context.Background(), &proto.UpdateScriptCategoryRequest{
		HostId:   uint32(hostID),
		Type:     req.Type,
		Category: req.Category,
		NewName:  req.NewName,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *ScriptManagerGRPCClient) DeleteCategory(hostID uint64, req model.DeleteGitCategory) error {
	_, err := c.client.DeleteCategory(context.Background(), &proto.CreateScriptCategoryRequest{
		HostId:   uint32(hostID),
		Type:     req.Type,
		Category: req.Category,
	})
	if err != nil {
		return err
	}
	return nil
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
	for _, item := range resp.Items {
		result.Items = append(result.Items, &model.ScriptInfo{
			Source:    item.Source,
			Name:      item.Name,
			Extension: item.Extension,
			Content:   item.Content,
			Size:      item.Size,
			ModTime:   item.ModTime,
			Linked:    item.Linked,
		})
	}
	result.Total = resp.Total
	return &result, nil
}

func (c *ScriptManagerGRPCClient) GetScriptDetail(hostID uint64, req model.GetGitFileDetail) (*model.ScriptInfo, error) {
	resp, err := c.client.GetScriptDetail(context.Background(), &proto.ScriptRequest{
		HostId:   uint32(hostID),
		Type:     req.Type,
		Category: req.Category,
		Name:     req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &model.ScriptInfo{
		Source:    resp.Source,
		Name:      resp.Name,
		Extension: resp.Extension,
		Content:   resp.Content,
		Size:      resp.Size,
		ModTime:   resp.ModTime,
		Linked:    resp.Linked,
	}, nil
}
func (c *ScriptManagerGRPCClient) CreateScript(hostID uint64, req model.CreateGitFile) error {
	_, err := c.client.CreateScript(context.Background(), &proto.CreateScriptRequest{
		HostId:   uint32(hostID),
		Type:     req.Type,
		Category: req.Category,
		Name:     req.Name,
		Content:  req.Content,
	})
	if err != nil {
		return err
	}
	return nil
}
func (c *ScriptManagerGRPCClient) UpdateScript(hostID uint64, req model.UpdateGitFile) error {
	_, err := c.client.UpdateScript(context.Background(), &proto.UpdateScriptRequest{
		HostId:      uint32(hostID),
		Type:        req.Type,
		Category:    req.Category,
		NewCategory: req.NewCategory,
		Name:        req.Name,
		NewName:     req.NewName,
		Content:     req.Content,
	})
	if err != nil {
		return err
	}
	return nil
}
func (c *ScriptManagerGRPCClient) DeleteScript(hostID uint64, req model.DeleteGitFile) error {
	_, err := c.client.DeleteScript(context.Background(), &proto.ScriptRequest{
		HostId:   uint32(hostID),
		Type:     req.Type,
		Category: req.Category,
		Name:     req.Name,
	})
	if err != nil {
		return err
	}
	return nil
}
func (c *ScriptManagerGRPCClient) RestoreScript(hostID uint64, req model.RestoreGitFile) error {
	_, err := c.client.RestoreScript(context.Background(), &proto.RestoreScriptRequest{
		HostId:     uint32(hostID),
		Type:       req.Type,
		Category:   req.Category,
		Name:       req.Name,
		CommitHash: req.CommitHash,
	})
	if err != nil {
		return err
	}
	return nil
}
func (c *ScriptManagerGRPCClient) GetScriptHistory(hostID uint64, req model.GitFileLog) (*model.ScriptHistoryList, error) {
	resp, err := c.client.GetScriptHistory(context.Background(), &proto.ScriptHistoryRequest{
		HostId:   uint32(hostID),
		Type:     req.Type,
		Category: req.Category,
		Name:     req.Name,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	var result model.ScriptHistoryList
	for _, history := range resp.Items {
		date, _ := time.Parse("2006-01-02T15:04:05+08:00", history.Date)
		result.Items = append(result.Items, &model.ScriptHistory{
			CommitHash:    history.CommitHash,
			Author:        history.Author,
			Email:         history.Email,
			Time:          date,
			CommitMessage: history.CommitMessage,
		})
	}
	result.Total = resp.Total

	return &result, nil
}
func (c *ScriptManagerGRPCClient) GetScriptDiff(hostID uint64, req model.GitFileDiff) (string, error) {
	resp, err := c.client.GetScriptDiff(context.Background(), &proto.ScriptDiffRequest{
		HostId:     uint32(hostID),
		Type:       req.Type,
		Category:   req.Category,
		Name:       req.Name,
		CommitHash: req.CommitHash,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (c *ScriptManagerGRPCClient) ScriptSync(hostID uint) error {
	_, err := c.client.ScriptSync(context.Background(), &proto.ScriptSyncRequest{
		HostId: uint32(hostID),
	})
	if err != nil {
		return err
	}
	return nil
}
func (c *ScriptManagerGRPCClient) ScriptExec(hostID uint, req model.ExecuteScript) (*model.ScriptResult, error) {
	resp, err := c.client.ScriptExec(context.Background(), &proto.ScriptExecRequest{
		HostId:     uint32(hostID),
		ScriptPath: req.ScriptPath,
	})
	if err != nil {
		return nil, err
	}
	start, _ := time.Parse("2006-01-02T15:04:05+08:00", resp.Start)
	end, _ := time.Parse("2006-01-02T15:04:05+08:00", resp.End)
	return &model.ScriptResult{
		LogHost: uint(resp.LogHost),
		LogPath: resp.LogPath,
		Start:   start,
		End:     end,
		Out:     resp.Out,
		Err:     resp.End,
	}, nil
}
func (c *ScriptManagerGRPCClient) GetScriptRunLogs(hostID uint, scriptPath string, page int, pageSize int) (*model.RunLogList, error) {
	resp, err := c.client.GetScriptRunLogs(context.Background(), &proto.ScriptRunLogsRequest{
		HostId:     uint32(hostID),
		ScriptPath: scriptPath,
		Page:       int32(page),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		return nil, err
	}
	var result model.RunLogList

	for _, file := range resp.Items {
		result.Items = append(result.Items, &model.RunLogInfo{
			Path:      file.Path,
			Name:      file.Name,
			Extension: file.Extension,
			Size:      int64(file.Size),
			IsDir:     file.IsDir,
			IsHidden:  file.IsHidden,
			CreatedAt: file.CreatedAt,
		})
	}
	result.Total = resp.Total

	return &result, nil
}
func (c *ScriptManagerGRPCClient) GetScriptRunLogDetail(hostID uint, logPath string) (*model.RunLogDetail, error) {
	resp, err := c.client.GetScriptRunLogDetail(context.Background(), &proto.ScriptRunLogDetailRequest{
		HostId:  uint32(hostID),
		LogPath: logPath,
	})
	if err != nil {
		return nil, err
	}
	return &model.RunLogDetail{
		Source:    resp.Source,
		Name:      resp.Name,
		Extension: resp.Extension,
		Content:   resp.Content,
		Size:      resp.Size,
		ModTime:   resp.ModTime,
	}, nil
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
	resp.Total = result.Total
	return &resp, nil
}

func (s *ScriptManagerGRPCServer) CreateCategory(ctx context.Context, req *proto.CreateScriptCategoryRequest) (*proto.CommonResponse, error) {
	var resp proto.CommonResponse

	err := s.Impl.CreateCategory(
		uint64(req.HostId),
		model.CreateGitCategory{
			Type:     req.Type,
			Category: req.Category,
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}

func (s *ScriptManagerGRPCServer) UpdateCategory(ctx context.Context, req *proto.UpdateScriptCategoryRequest) (*proto.CommonResponse, error) {
	var resp proto.CommonResponse

	err := s.Impl.UpdateCategory(
		uint64(req.HostId),
		model.UpdateGitCategory{
			Type:     req.Type,
			Category: req.Category,
			NewName:  req.NewName,
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
	return &resp, nil
}
func (s *ScriptManagerGRPCServer) DeleteCategory(ctx context.Context, req *proto.CreateScriptCategoryRequest) (*proto.CommonResponse, error) {
	var resp proto.CommonResponse

	err := s.Impl.DeleteCategory(
		uint64(req.HostId),
		model.DeleteGitCategory{
			Type:     req.Type,
			Category: req.Category,
		},
	)
	if err != nil {
		return &resp, err
	}
	resp.Success = true
	resp.Error = ""
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
	resp.Total = result.Total
	return &resp, nil
}

func (s *ScriptManagerGRPCServer) GetScriptDetail(ctx context.Context, req *proto.ScriptRequest) (*proto.ScriptInfo, error) {
	result, err := s.Impl.GetScriptDetail(
		uint64(req.HostId),
		model.GetGitFileDetail{
			Type:     req.Type,
			Category: req.Category,
			Name:     req.Name,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.ScriptInfo{
		Source:    result.Source,
		Name:      result.Name,
		Extension: result.Extension,
		Content:   result.Content,
		Size:      result.Size,
		ModTime:   result.ModTime,
		Linked:    result.Linked,
	}, nil
}
func (s *ScriptManagerGRPCServer) CreateScript(ctx context.Context, req *proto.CreateScriptRequest) (*proto.CommonResponse, error) {
	err := s.Impl.CreateScript(
		uint64(req.HostId),
		model.CreateGitFile{
			Type:     req.Type,
			Category: req.Category,
			Name:     req.Name,
			Content:  req.Content,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.CommonResponse{
		Success: true,
		Error:   "",
	}, nil
}
func (s *ScriptManagerGRPCServer) UpdateScript(ctx context.Context, req *proto.UpdateScriptRequest) (*proto.CommonResponse, error) {
	err := s.Impl.UpdateScript(
		uint64(req.HostId),
		model.UpdateGitFile{
			Type:        req.Type,
			Category:    req.Category,
			NewCategory: req.NewCategory,
			Name:        req.Name,
			NewName:     req.NewName,
			Content:     req.Content,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.CommonResponse{
		Success: true,
		Error:   "",
	}, nil
}
func (s *ScriptManagerGRPCServer) DeleteScript(ctx context.Context, req *proto.ScriptRequest) (*proto.CommonResponse, error) {
	err := s.Impl.DeleteScript(
		uint64(req.HostId),
		model.DeleteGitFile{
			Type:     req.Type,
			Category: req.Category,
			Name:     req.Name,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.CommonResponse{
		Success: true,
		Error:   "",
	}, nil
}
func (s *ScriptManagerGRPCServer) RestoreScript(ctx context.Context, req *proto.RestoreScriptRequest) (*proto.CommonResponse, error) {
	err := s.Impl.RestoreScript(
		uint64(req.HostId),
		model.RestoreGitFile{
			Type:       req.Type,
			Category:   req.Category,
			Name:       req.Name,
			CommitHash: req.CommitHash,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.CommonResponse{
		Success: true,
		Error:   "",
	}, nil
}
func (s *ScriptManagerGRPCServer) GetScriptHistory(ctx context.Context, req *proto.ScriptHistoryRequest) (*proto.ScriptHistoryResponse, error) {
	var resp proto.ScriptHistoryResponse
	result, err := s.Impl.GetScriptHistory(
		uint64(req.HostId),
		model.GitFileLog{
			Type:     req.Type,
			Category: req.Category,
			Name:     req.Name,
			Page:     int(req.Page),
			PageSize: int(req.PageSize),
		},
	)
	if err != nil {
		return &resp, err
	}
	for _, item := range result.Items {
		resp.Items = append(resp.Items, &proto.ScriptHistory{
			CommitHash:    item.CommitHash,
			Author:        item.Author,
			Email:         item.Email,
			Date:          item.Time.Format("2006-01-02T15:04:05+08:00"),
			CommitMessage: item.CommitMessage,
		})
	}
	resp.Total = result.Total
	return &resp, nil
}
func (s *ScriptManagerGRPCServer) GetScriptDiff(ctx context.Context, req *proto.ScriptDiffRequest) (*proto.ScriptDiffResponse, error) {
	result, err := s.Impl.GetScriptDiff(
		uint64(req.HostId),
		model.GitFileDiff{
			Type:       req.Type,
			Category:   req.Category,
			Name:       req.Name,
			CommitHash: req.CommitHash,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.ScriptDiffResponse{
		Content: result,
	}, nil
}
func (s *ScriptManagerGRPCServer) ScriptSync(ctx context.Context, req *proto.ScriptSyncRequest) (*proto.CommonResponse, error) {
	err := s.Impl.ScriptSync(uint(req.HostId))
	if err != nil {
		return nil, err
	}
	return &proto.CommonResponse{
		Success: true,
		Error:   "",
	}, nil
}
func (s *ScriptManagerGRPCServer) ScriptExec(ctx context.Context, req *proto.ScriptExecRequest) (*proto.ScriptExecResponse, error) {
	result, err := s.Impl.ScriptExec(
		uint(req.HostId),
		model.ExecuteScript{
			ScriptPath: req.ScriptPath,
		},
	)
	if err != nil {
		return nil, err
	}
	return &proto.ScriptExecResponse{
		LogHost: uint32(result.LogHost),
		LogPath: result.LogPath,
		Start:   result.Start.Format("2006-01-02T15:04:05+08:00"),
		End:     result.End.Format("2006-01-02T15:04:05+08:00"),
		Out:     result.Out,
		Err:     result.Err,
	}, nil
}
func (s *ScriptManagerGRPCServer) GetScriptRunLogs(ctx context.Context, req *proto.ScriptRunLogsRequest) (*proto.ScriptRunLogsResponse, error) {
	result, err := s.Impl.GetScriptRunLogs(
		uint(req.HostId),
		req.ScriptPath,
		int(req.Page),
		int(req.PageSize),
	)
	if err != nil {
		return nil, err
	}
	var resp proto.ScriptRunLogsResponse
	for _, item := range result.Items {
		resp.Items = append(resp.Items, &proto.RunLogInfo{
			Path:      item.Path,
			Name:      item.Name,
			Extension: item.Extension,
			Size:      int32(item.Size),
			IsDir:     item.IsDir,
			IsHidden:  item.IsHidden,
			CreatedAt: item.CreatedAt,
		})
	}
	resp.Total = result.Total
	return &resp, nil
}
func (s *ScriptManagerGRPCServer) GetScriptRunLogDetail(ctx context.Context, req *proto.ScriptRunLogDetailRequest) (*proto.ScriptRunLogDetailResponse, error) {
	result, err := s.Impl.GetScriptRunLogDetail(
		uint(req.HostId),
		req.LogPath,
	)
	if err != nil {
		return nil, err
	}
	return &proto.ScriptRunLogDetailResponse{
		Source:    result.Source,
		Name:      result.Name,
		Extension: result.Extension,
		Content:   result.Content,
		Size:      result.Size,
		ModTime:   result.ModTime,
	}, nil
}
