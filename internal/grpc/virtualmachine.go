package grpc

import (
	"context"

	grpccore "github.com/cynxees/cynx-core/src/grpc"
	pb "github.com/cynxees/ra-server/api/proto/gen/ra"
)

func (s *Server) GetVirtualMachine(ctx context.Context, req *pb.GetVirtualMachineRequest) (resp *pb.VirtualMachineResponse, err error) {
	return grpccore.HandleGrpc(ctx, req, resp, s.VirtualMachineService.GetVirtualMachine)
}
