package grpc

import (
	"context"
	"github.com/cynxees/ra-server/internal/service/virtualmachineservice"
	"net"

	"github.com/cynxees/cynx-core/src/logger"
	pb "github.com/cynxees/ra-server/api/proto/gen/ra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedVirtualMachineServiceServer

	VirtualMachineService *virtualmachineservice.Service
}

func (s *Server) Start(ctx context.Context, address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	pb.RegisterVirtualMachineServiceServer(server, s)
	reflection.Register(server)

	logger.Info(ctx, "Starting gRPC server on ", address)
	return server.Serve(lis)
}
