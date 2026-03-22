package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"refina-profile/config/db"
	"refina-profile/config/env"
	logger "refina-profile/config/log"
	"refina-profile/config/miniofs"
	grpcserver "refina-profile/interface/grpc/server"
	"refina-profile/internal/utils"
	"refina-profile/internal/utils/data"
)

func init() {
	var err error
	var missing []string
	if missing, err = env.LoadByViper(); err != nil {
		log.Printf("Failed to read JSON config file: %v", err)
		if missing, err = env.LoadNative(); err != nil {
			log.Fatalf("Failed to load environment variables: %v", err)
		}
		log.Printf("Environment variables by .env file loaded successfully")
	} else {
		log.Printf("Environment variables by Viper loaded successfully")
	}

	logger.SetupLogger()

	if len(missing) > 0 {
		for _, envVar := range missing {
			logger.Warn(data.LogEnvVarMissing, map[string]any{"service": data.EnvService, "env_var": envVar})
		}
	}
}

func main() {
	// Setup Database Connection
	startTime := time.Now()
	dbInstance := db.GetInstance(env.Cfg.Database)
	logger.Info(data.LogDBSetupSuccess, map[string]any{"service": data.DatabaseService, "duration": utils.Ms(time.Since(startTime))})

	// Setup MinIO Connection
	startTime = time.Now()
	minioInstance := miniofs.SetupMinio(env.Cfg.Minio)
	logger.Info(data.LogMinioSetupSuccess, map[string]any{"service": data.MinioService, "duration": utils.Ms(time.Since(startTime))})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up the gRPC server
	startTime = time.Now()
	grpcServer, listener, err := grpcserver.SetupGRPCServer(dbInstance, minioInstance)
	if err != nil {
		logger.Fatal("Failed to setup gRPC server", map[string]any{"service": data.GRPCService, "error": err.Error()})
	}

	go func() {
		logger.Info(data.LogGRPCStarted, map[string]any{
			"service": data.GRPCService,
			"port":    env.Cfg.Server.GRPCPort,
		})
		if err := grpcServer.Serve(*listener); err != nil {
			logger.Fatal("gRPC server failed", map[string]any{"service": data.GRPCService, "error": err.Error()})
		}
	}()
	logger.Info(data.LogGRPCStarted, map[string]any{"service": data.GRPCService, "duration": utils.Ms(time.Since(startTime))})

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down servers...", map[string]any{"service": data.ProfileService})

	// Shutdown context with timeout
	_, cancelShutdown := context.WithTimeout(ctx, 30*time.Second)
	defer cancelShutdown()

	// Stop gRPC server
	grpcServer.GracefulStop()
	logger.Info("gRPC server stopped", map[string]any{"service": data.GRPCService})

	// Close database connection
	if err := dbInstance.Close(); err != nil {
		logger.Error("Failed to close database", map[string]any{"service": data.DatabaseService, "error": err.Error()})
	} else {
		logger.Info("Database connection closed", map[string]any{"service": data.DatabaseService})
	}

	logger.Info("Server shutdown complete", map[string]any{"service": data.ProfileService})
}

// Placeholder for HTTP server if needed in the future
func setupHTTPServer() *http.Server {
	return nil
}
