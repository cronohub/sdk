package sdk

import (
	"context"

	proto "github.com/cronohub/protoc/cronoprot"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion: 1,
	MagicCookieKey:  "CRONOHUB_PLUGINS",
	// Never ever change this.
	MagicCookieValue: "ce118ec2-6c69-48a2-a7c5-02787052ec95",
}

/*
 *
 * Plugin interface declarations.
 *
 */

// Archive takes a locate to a file and archives it giving back a
// a bool as a result of the archiving procedure.
type Archive interface {
	Execute(payload string) (bool, error)
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

// GRPCArchiveClient is an implementation of Archive that talks over RPC.
type GRPCArchiveClient struct {
	client proto.ArchiveClient
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

// Execute is the GRPC implementation of the Execute function for the
// Archive plugin definition. This will talk over GRPC.
func (m *GRPCArchiveClient) Execute(filename string) (bool, error) {
	p, err := m.client.Execute(context.Background(), &proto.Payload{
		File: filename,
	})
	if err != nil {
		return false, err
	}
	return p.Success, nil
}

// GRPCArchiveServer is the gRPC server that GRPCArchiveClient talks to.
type GRPCArchiveServer struct {
	// This is the real implementation
	Impl Archive
}

// Execute is the execute function of the GRPCServer which will rely the information to the
// underlying implementation of this interface.
func (m *GRPCArchiveServer) Execute(ctx context.Context, req *proto.Payload) (*proto.Status, error) {
	res, err := m.Impl.Execute(req.File)
	return &proto.Status{Success: res}, err
}
