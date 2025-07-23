package main

import (
	"context"
	"github.com/cynxees/cynx-core/src/logger"
	"github.com/cynxees/ra-server/internal/app"
	"log"
)

func main() {

	log.Println("Starting ra")
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	log.Println("Initializing App")
	application, err := app.NewApp(ctx)
	if err != nil {
		panic(err)
	}

	logger.Info(ctx, "Creating servers")
	servers, err := application.NewServers()
	if err != nil {
		panic(err)
	}

	logger.Info(ctx, "Starting servers")
	if err := servers.Start(ctx); err != nil {
		panic(err)
	}

	<-ctx.Done()
}
