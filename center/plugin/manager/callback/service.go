package callback

import (
	"context"

	"github.com/sensdata/idb/center/plugin/manager/callback/pb"
)

type CallbackServer struct {
	pb.UnimplementedCenterCallbackServer
	// 可以注入 db、agent manager 等资源
}

func (s *CallbackServer) GetHostInfo(ctx context.Context, req *pb.HostInfoRequest) (*pb.HostInfoResponse, error) {
	// 模拟数据库查 host 信息
	return &pb.HostInfoResponse{
		HostId:   1,
		Hostname: "host-123",
	}, nil
}

func (s *CallbackServer) ExecuteCommand(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	// 模拟调用 agent 执行命令
	return &pb.CommandResponse{
		ExitCode: 0,
		Stdout:   "output",
		Stderr:   "",
	}, nil
}
