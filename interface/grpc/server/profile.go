package server

import (
	"context"

	"refina-profile/config/log"
	"refina-profile/interface/grpc/interceptor"
	"refina-profile/internal/service"
	"refina-profile/internal/types/dto"
	"refina-profile/internal/utils/data"

	ppb "github.com/MuhammadMiftaa/Refina-Protobuf/profile"
)

type profileServer struct {
	ppb.UnimplementedProfileServiceServer
	profileService service.ProfileService
}

func NewProfileServer(profileService service.ProfileService) ppb.ProfileServiceServer {
	return &profileServer{
		profileService: profileService,
	}
}

// GetProfile retrieves user profile
func (s *profileServer) GetProfile(ctx context.Context, req *ppb.GetProfileRequest) (*ppb.Profile, error) {
	userID := req.GetUserId()
	if userID == "" {
		userID = interceptor.UserIDFromContext(ctx)
	}

	profile, err := s.profileService.GetProfile(ctx, userID)
	if err != nil {
		log.Error(data.LogGetProfileFailed, map[string]any{
			data.LogFieldService: data.GRPCService,
			data.LogFieldUserID:  userID,
			data.LogFieldError:   err.Error(),
		})
		return nil, err
	}

	log.Info(data.LogGetProfileSuccess, map[string]any{
		data.LogFieldService:   data.GRPCService,
		data.LogFieldUserID:    userID,
		data.LogFieldProfileID: profile.ID,
	})

	return toProtoProfile(profile), nil
}

// CreateProfile creates a new user profile
func (s *profileServer) CreateProfile(ctx context.Context, req *ppb.CreateProfileRequest) (*ppb.Profile, error) {
	userID := req.GetUserId()
	if userID == "" {
		userID = interceptor.UserIDFromContext(ctx)
	}

	profile, err := s.profileService.CreateProfile(ctx, userID, req.GetFullname())
	if err != nil {
		log.Error(data.LogCreateProfileFailed, map[string]any{
			data.LogFieldService: data.GRPCService,
			data.LogFieldUserID:  userID,
			data.LogFieldError:   err.Error(),
		})
		return nil, err
	}

	log.Info(data.LogCreateProfileSuccess, map[string]any{
		data.LogFieldService:   data.GRPCService,
		data.LogFieldUserID:    userID,
		data.LogFieldProfileID: profile.ID,
	})

	return toProtoProfile(profile), nil
}

// UpdateProfile updates user profile
func (s *profileServer) UpdateProfile(ctx context.Context, req *ppb.UpdateProfileRequest) (*ppb.Profile, error) {
	userID := req.GetUserId()
	if userID == "" {
		userID = interceptor.UserIDFromContext(ctx)
	}

	updateReq := dto.UpdateProfileRequest{
		Fullname: req.GetFullname(),
	}

	profile, err := s.profileService.UpdateProfile(ctx, userID, updateReq)
	if err != nil {
		log.Error(data.LogUpdateProfileFailed, map[string]any{
			data.LogFieldService: data.GRPCService,
			data.LogFieldUserID:  userID,
			data.LogFieldError:   err.Error(),
		})
		return nil, err
	}

	log.Info(data.LogUpdateProfileSuccess, map[string]any{
		data.LogFieldService:   data.GRPCService,
		data.LogFieldUserID:    userID,
		data.LogFieldProfileID: profile.ID,
	})

	return toProtoProfile(profile), nil
}

// UploadProfilePhoto uploads user profile photo
func (s *profileServer) UploadProfilePhoto(ctx context.Context, req *ppb.UploadProfilePhotoRequest) (*ppb.UploadProfilePhotoResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		userID = interceptor.UserIDFromContext(ctx)
	}

	result, err := s.profileService.UploadPhoto(ctx, userID, req.GetBase64Image())
	if err != nil {
		log.Error(data.LogUploadPhotoFailed, map[string]any{
			data.LogFieldService: data.GRPCService,
			data.LogFieldUserID:  userID,
			data.LogFieldError:   err.Error(),
		})
		return &ppb.UploadProfilePhotoResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Info(data.LogUploadPhotoSuccess, map[string]any{
		data.LogFieldService: data.GRPCService,
		data.LogFieldUserID:  userID,
	})

	return &ppb.UploadProfilePhotoResponse{
		Success:  result.Success,
		PhotoUrl: result.PhotoURL,
		Message:  result.Message,
	}, nil
}

// DeleteProfilePhoto deletes user profile photo
func (s *profileServer) DeleteProfilePhoto(ctx context.Context, req *ppb.DeleteProfilePhotoRequest) (*ppb.DeleteProfilePhotoResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		userID = interceptor.UserIDFromContext(ctx)
	}

	result, err := s.profileService.DeletePhoto(ctx, userID)
	if err != nil {
		log.Error(data.LogDeletePhotoFailed, map[string]any{
			data.LogFieldService: data.GRPCService,
			data.LogFieldUserID:  userID,
			data.LogFieldError:   err.Error(),
		})
		return &ppb.DeleteProfilePhotoResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Info(data.LogDeletePhotoSuccess, map[string]any{
		data.LogFieldService: data.GRPCService,
		data.LogFieldUserID:  userID,
	})

	return &ppb.DeleteProfilePhotoResponse{
		Success: result.Success,
		Message: result.Message,
	}, nil
}

// Helper function to convert DTO to proto
func toProtoProfile(profile *dto.ProfileResponse) *ppb.Profile {
	return &ppb.Profile{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Fullname:  profile.Fullname,
		PhotoUrl:  profile.PhotoURL,
		CreatedAt: profile.CreatedAt,
		UpdatedAt: profile.UpdatedAt,
	}
}
