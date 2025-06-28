package manager

import (
	"google.golang.org/grpc"
)

type PluginFactory func(cc *grpc.ClientConn) interface{}

var PluginFactories = map[string]PluginFactory{}

func RegisterFactory(name string, factory PluginFactory) {
	PluginFactories[name] = factory
}
