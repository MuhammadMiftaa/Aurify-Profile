package server

import (
	"net"

	"refina-profile/config/db"
	"refina-profile/config/env"
	"refina-profile/config/miniofs"
	"refina-profile/interface/grpc/interceptor"
	"refina-profile/internal/repository"
	"refina-profile/internal/service"

	ppb "github.com/MuhammadMiftaa/Refina-Protobuf/profile"
	"google.golang.org/grpc"
)

func SetupGRPCServer(dbInstance db.DatabaseClient, minioInstance *miniofs.MinIOManager) (*grpc.Server, *net.Listener, error) {
	lis, err := net.Listen("tcp", ":"+env.Cfg.Server.GRPCPort)
	if err != nil {
		return nil, nil, err
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryServerInterceptor()),
		grpc.StreamInterceptor(interceptor.StreamServerInterceptor()),
	)

	// ── Repositories ──
	profileRepo := repository.NewProfileRepository(dbInstance.GetDB())

	// ── Services ──
	profileService := service.NewProfileService(profileRepo, minioInstance)

	// ── Register gRPC Server ──
	profileServer := NewProfileServer(profileService)
	ppb.RegisterProfileServiceServer(s, profileServer)

	return s, &lis, nil
}
