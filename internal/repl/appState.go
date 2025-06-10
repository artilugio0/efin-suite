package repl

import (
	"github.com/artilugio0/efin-suite/internal/grpc/proto"
	"google.golang.org/grpc"
)

type appState struct {
	proxyClient proto.ProxyServiceClient
	proxyConn   *grpc.ClientConn
}
