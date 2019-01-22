package sdk

import (
	"context"

	"github.com/cronohub/protoc/proto"
	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

/*
 *
 * Plugin interface declarations.
 *
 */

// Archive takes a locate to a file and archives it giving back a
// a bool as a result of the archiving procedure.
type Archive interface {
	Execute(payload string) bool
}

/*
 *
 * Archive Plugin structs and functions.
 *
 */

// ArchiveGRPCPlugin is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type ArchiveGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Archive
}

// GRPCServer is the grpc server implementation which calls the
// protoc generated code to register it.
func (p *ArchiveGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterArchiveServer(s, &GRPCArchiveServer{Impl: p.Impl})
	return nil
}

// GRPCClient is the grpc client that will talk to the GRPC Server
// and calls into the generated protoc code.
func (p *ArchiveGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCArchiveClient{client: proto.NewArchiveClient(c)}, nil
}

// GRPCArchiveClient is an implementation of Archive that talks over RPC.
type GRPCArchiveClient struct{ client proto.ArchiveClient }

// Execute is the GRPC implementation of the Execute function for the
// Archive plugin definition. This will talk over GRPC.
func (m *GRPCArchiveClient) Execute(filename string) bool {
	p, err := m.client.Execute(context.Background(), &proto.Payload{
		File: filename,
	})
	if err != nil {
		return false
	}
	return p.Success
}

// GRPCArchiveServer is the gRPC server that GRPCArchiveClient talks to.
type GRPCArchiveServer struct {
	// This is the real implementation
	Impl Archive
}

// Execute is the execute function of the GRPCServer which will rely the information to the
// underlying implementation of this interface.
func (m *GRPCArchiveServer) Execute(ctx context.Context, req *proto.Payload) (*proto.Status, error) {
	res := m.Impl.Execute(req.File)
	return &proto.Status{Success: res}, nil
}
