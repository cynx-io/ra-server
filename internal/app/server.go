package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/cynxees/cynx-core/src/logger"
	"github.com/cynxees/ra-server/internal/dependencies/config"
	"github.com/cynxees/ra-server/internal/grpc"
	"golang.org/x/sync/errgroup"
)

type Servers struct {
	grpcServer *grpc.Server
}

func (app *App) NewServers() (*Servers, error) {
	services := app.Services

	// Create gRPC server
	grpcServer := &grpc.Server{
		VirtualMachineService: services.VirtualMachineService,
	}

	return &Servers{
		grpcServer: grpcServer,
	}, nil
}

func (s *Servers) Start(ctx context.Context) error {
	var g errgroup.Group

	g.Go(func() error {
		logger.Info(ctx, "Starting gRPC server")
		address := config.Config.App.Address + ":" + strconv.Itoa(config.Config.App.Port)
		if err := s.grpcServer.Start(ctx, address); err != nil {
			return fmt.Errorf("failed to start gRPC server: %w", err)
		}
		return nil
	})

	return g.Wait()
}

func (s *Servers) Stop() error {
	return errors.New("stop not implemented")
}
