package repl

import (
	"github.com/artilugio0/efin-proxy/pkg/grpc/proto"
	"google.golang.org/grpc"
)

type appState struct {
	proxyClient proto.ProxyServiceClient
	proxyConn   *grpc.ClientConn
}
