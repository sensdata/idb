package scriptmanager

import (
	"context"

	smpb "github.com/sensdata/idb/center/plugin/scriptmanager/pb"

	"google.golang.org/grpc"
)

type ScriptManager interface {
	ListScripts(ctx context.Context, req *smpb.ListScriptsRequest) ([]string, error)
}

type ScriptManagerWrapper struct {
	Client smpb.ScriptManagerClient
}

func NewGRPCClient(cc *grpc.ClientConn) smpb.ScriptManagerClient {
	return smpb.NewScriptManagerClient(cc)
}

func (w *ScriptManagerWrapper) ListScripts(ctx context.Context, req *smpb.ListScriptsRequest) ([]string, error) {
	resp, err := w.Client.ListScripts(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Scripts, nil
}
