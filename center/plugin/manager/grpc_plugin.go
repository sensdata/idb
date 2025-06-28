package manager

import (
	"net/rpc"

	hplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type GRPCPlugin struct {
	hplugin.GRPCPlugin
	NewClient PluginFactory
}

func (p *GRPCPlugin) GRPCClient(b *hplugin.GRPCBroker, cc *grpc.ClientConn) (interface{}, error) {
	return p.NewClient(cc), nil
}

func (p *GRPCPlugin) GRPCServer(*hplugin.GRPCBroker, *grpc.Server) error {
	return nil
}

func (p *GRPCPlugin) Client(*hplugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, nil
}

func (p *GRPCPlugin) Server(*hplugin.MuxBroker) (interface{}, error) {
	return nil, nil
}
